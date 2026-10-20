package bible

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/rohanthewiz/bytdb/stdlib"
)

// openPrefsTestDB creates a bible database with just the prefs table. The DDL
// duplicates store/schema.go on purpose, as study's sync tests do: importing
// store would pull the whole app's startup, lock included, into this test.
func openPrefsTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("bytdb", filepath.Join(t.TempDir(), "bible.bytdb"))
	if err != nil {
		t.Fatalf("sql.Open() error: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(`CREATE TABLE prefs (key TEXT PRIMARY KEY, value TEXT)`); err != nil {
		t.Fatalf("create prefs: %v", err)
	}
	return db
}

func TestPreferredTranslation(t *testing.T) {
	db := openPrefsTestDB(t)

	got, err := PreferredTranslation(db)
	if err != nil || got != "" {
		t.Fatalf("unset preference = %q, %v; want \"\", nil", got, err)
	}

	// Setting twice exercises the replace path: the second value must win
	// and must not collide with the first row's key.
	for _, abbrev := range []string{"web", "asv"} {
		if err := SetPreferredTranslation(db, abbrev); err != nil {
			t.Fatalf("SetPreferredTranslation(%q) error: %v", abbrev, err)
		}
		got, err := PreferredTranslation(db)
		if err != nil || got != abbrev {
			t.Fatalf("after setting %q, got %q, %v", abbrev, got, err)
		}
	}

	if err := SetPreferredTranslation(db, "niv"); err == nil {
		t.Fatal("SetPreferredTranslation(\"niv\") succeeded; want an unknown-translation error")
	}
	if got, _ := PreferredTranslation(db); got != "asv" {
		t.Fatalf("a rejected set changed the preference to %q", got)
	}
}
