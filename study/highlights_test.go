package study

import (
	"testing"
	"time"

	"github.com/rohanthewiz/pbstudy/bible"
)

var (
	hlJohn316 = bible.Ref{BookNum: 43, Chapter: 3, VerseStart: 16, VerseEnd: 16}
	hlJohn317 = bible.Ref{BookNum: 43, Chapter: 3, VerseStart: 17, VerseEnd: 17}
)

// TestHighlightLifecycle walks one verse through set, recolour, clear and set
// again, and checks that it stays one row throughout. The derived id is what
// sync convergence rests on, so a second row for the same verse is the failure
// this test is really guarding.
func TestHighlightLifecycle(t *testing.T) {
	db := openSyncTestDB(t)

	if c, err := HighlightOf(db, 43, 3, 16); err != nil || c != "" {
		t.Fatalf("unhighlighted verse = %q, %v; want \"\", nil", c, err)
	}

	steps := []struct {
		color string // "" means clear
		want  string
	}{
		{"yellow", "yellow"},
		{"blue", "blue"},
		{"", ""},
		{"green", "green"}, // revives the tombstoned row
	}
	for _, s := range steps {
		var err error
		if s.color == "" {
			err = ClearHighlight(db, hlJohn316)
		} else {
			err = SetHighlight(db, hlJohn316, s.color)
		}
		if err != nil {
			t.Fatalf("set %q: %v", s.color, err)
		}
		if got, err := HighlightOf(db, 43, 3, 16); err != nil || got != s.want {
			t.Fatalf("after set %q: HighlightOf = %q, %v; want %q", s.color, got, err, s.want)
		}
	}

	var rows int
	if err := db.QueryRow(`SELECT COUNT(*) FROM highlights`).Scan(&rows); err != nil {
		t.Fatalf("count: %v", err)
	}
	if rows != 1 {
		t.Errorf("one verse left %d highlight rows, want 1", rows)
	}

	// Clearing a verse with no highlight is a no-op, not an error.
	if err := ClearHighlight(db, hlJohn317); err != nil {
		t.Errorf("ClearHighlight on an unhighlighted verse: %v", err)
	}
}

func TestSetHighlightRejects(t *testing.T) {
	db := openSyncTestDB(t)

	if err := SetHighlight(db, hlJohn316, "purple"); err == nil {
		t.Error("an unknown colour was accepted")
	}
	chapter := bible.Ref{BookNum: 43, Chapter: 3}
	if err := SetHighlight(db, chapter, "yellow"); err == nil {
		t.Error("a whole-chapter highlight was accepted")
	}
}

func TestChapterHighlights(t *testing.T) {
	db := openSyncTestDB(t)

	for ref, c := range map[bible.Ref]string{hlJohn316: "yellow", hlJohn317: "pink"} {
		if err := SetHighlight(db, ref, c); err != nil {
			t.Fatalf("SetHighlight(%s): %v", ref, err)
		}
	}
	// Another chapter, and a cleared verse, must not appear.
	if err := SetHighlight(db, bible.Ref{BookNum: 43, Chapter: 4, VerseStart: 1}, "blue"); err != nil {
		t.Fatal(err)
	}
	if err := ClearHighlight(db, hlJohn317); err != nil {
		t.Fatal(err)
	}

	got, err := ChapterHighlights(db, 43, 3)
	if err != nil {
		t.Fatalf("ChapterHighlights: %v", err)
	}
	if len(got) != 1 || got[16] != "yellow" {
		t.Errorf("John 3 highlights = %v, want map[16:yellow]", got)
	}
}

// TestApplyHighlightRoundTrip: what ExportHighlights writes, ApplyHighlight
// reproduces — tombstone included.
func TestApplyHighlightRoundTrip(t *testing.T) {
	src := openSyncTestDB(t)
	dst := openSyncTestDB(t)

	if err := SetHighlight(src, hlJohn316, "blue"); err != nil {
		t.Fatal(err)
	}
	if err := SetHighlight(src, hlJohn317, "pink"); err != nil {
		t.Fatal(err)
	}
	if err := ClearHighlight(src, hlJohn317); err != nil {
		t.Fatal(err)
	}

	recs, err := ExportHighlights(src)
	if err != nil {
		t.Fatalf("ExportHighlights: %v", err)
	}
	for _, rec := range recs {
		if err := ApplyHighlight(dst, rec); err != nil {
			t.Fatalf("ApplyHighlight(%s): %v", rec.ID, err)
		}
	}

	again, err := ExportHighlights(dst)
	if err != nil {
		t.Fatalf("ExportHighlights(dst): %v", err)
	}
	if len(again) != len(recs) {
		t.Fatalf("round trip produced %d records, want %d", len(again), len(recs))
	}
	for i := range recs {
		a, b := recs[i], again[i]
		if a.ID != b.ID || a.At != b.At || a.Color != b.Color ||
			!a.UpdatedAt.Equal(b.UpdatedAt) || !a.SyncDeleted().Equal(b.SyncDeleted()) {
			t.Errorf("record %d changed in the round trip:\n got %+v\nwant %+v", i, b, a)
		}
	}
	if c, _ := HighlightOf(dst, 43, 3, 17); c != "" {
		t.Errorf("a cleared highlight arrived live as %q", c)
	}
}

// TestApplyHighlightRejectsForeignID: an id that is not the verse's own would
// let a second row colour the same verse, undoing the convergence the derived
// id exists for.
func TestApplyHighlightRejectsForeignID(t *testing.T) {
	db := openSyncTestDB(t)
	now := time.Now().UTC().Truncate(time.Microsecond)

	err := ApplyHighlight(db, HighlightRecord{
		Kind: SyncKindHighlight, V: SyncVersion, ID: "hl-1-1-1",
		At: hlJohn316, Color: "yellow", CreatedAt: now, UpdatedAt: now,
	})
	if err == nil {
		t.Fatal("a highlight whose id names another verse was applied")
	}
}
