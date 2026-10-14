package ui

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/rohanthewiz/element"

	"github.com/rohanthewiz/pbstudy/bible"
	"github.com/rohanthewiz/pbstudy/study"
)

// hostilePayload is written into every string a view receives. It opens with
// a quote and an angle bracket so it breaks out of an unescaped attribute
// value as well as unescaped element content; either failure leaves the
// literal "<script>pwn" in the output, which is the one thing the test greps
// for.
const hostilePayload = `"'><script>pwn()</script>`

// rawHTMLFields are the string fields that carry HTML on purpose and are
// written with b.WriteString rather than b.T. They hold goldmark output, which
// goldmark has already escaped (see renderMarkdownBody in notes.go), so the
// test leaves them empty rather than asserting something the view is
// documented not to do.
var rawHTMLFields = map[string]bool{
	"BodyHTML":  true,
	"DraftHTML": true,
}

// trustedFields are strings that reach a view already constrained, so the
// views write them without esc() by design. Each is given a realistic value
// instead of the payload. Keys are "Type.Field" rather than bare field names
// because the same name can be trusted on one type and user text on another
// (ChapterLink.Label is a canon book name; OutlineRow.Label is not).
//
// Every entry must say where the constraint is enforced. Adding one is a
// claim that the value can never carry markup, so it deserves the same
// scrutiny as removing an esc().
var trustedFields = map[string]string{
	// Checked against bible.Known by every handler that takes one from a
	// request (bible.IsKnownTranslation), or taken from the Known table.
	"Dashboard.Translation":     "kjv",
	"Reader.Translation":        "kjv",
	"Search.Translation":        "kjv",
	"NotesList.Translation":     "kjv",
	"NoteDetail.Translation":    "kjv",
	"TagPage.Translation":       "kjv",
	"SermonBuilder.Translation": "kjv",
	"VerseHub.ReadTranslation":  "kjv",
	"Verse.Translation":         "kjv",
	"Translation.Abbrev":        "kjv",
	"Translation.Name":          "King James Version",
	"Translation.Lang":          "en",

	// Passed through NormalizeScope, which maps anything unknown to ScopeAll.
	"Search.Scope": ScopeAll,

	// Built by the handlers from canon book names, ui.*URL helpers and ids.
	"ChapterLink.URL":      "/read/kjv/1/1",
	"ChapterLink.Label":    "Genesis 1",
	"OutlineRow.Href":      "/notes/" + testUUID,
	"NoteEditor.Action":    "/notes/" + testUUID,
	"NoteEditor.CancelURL": "/notes/" + testUUID,

	// Assembled by syncer's report types from counts and fixed words.
	"SyncStatus.LastHeadline": "Everything was already in step.",
	"SyncEntityLine.Label":    "Notes",
	"SyncEntityLine.Summary":  "nothing to do (1 already in step)",
}

const testUUID = "0b8f9c1e-2d3a-4b5c-8d7e-6f5a4b3c2d1e"

// TestEveryViewEscapes is the framework backstop the element builder lacks.
//
// element's T() and attribute values are written verbatim (see escape.go), so
// the only thing standing between a note title and the page is a hand-written
// esc(). The per-view tests (TestSermonBuilderEscapes, TestHighlightEscapes)
// pin the riskiest spots by name; this one covers the rest by construction:
// it walks each view struct with reflection, fills *every* exported string it
// can reach (nested structs, slices, maps, pointers) with hostilePayload, and
// renders it. A field added later is poisoned automatically, so a new b.T()
// that forgets esc() fails here without anyone remembering to extend a test.
//
// The walk also fills ints with 1 and gives every slice and map one element,
// so the branches that render lists and counts actually execute. 1 is chosen
// because it is a valid book number, chapter and verse, which keeps the
// canon lookups (bible.MustByNum and friends) from panicking.
func TestEveryViewEscapes(t *testing.T) {
	// noUserText marks a view that renders nothing a user can type. It
	// still runs through the walk (so a user-text field added to it later is
	// poisoned automatically), but it is excused from the "payload must
	// appear" guard, which would otherwise fail for want of input.
	views := []struct {
		name       string
		view       element.Component
		noUserText bool
	}{
		// Counts and the Known translation table only.
		{"Dashboard", &Dashboard{}, true},
		{"Settings", &Settings{}, false},
		{"ErrorPage", &ErrorPage{}, false},
		{"Reader", &Reader{}, false},
		{"Search", &Search{}, false},
		{"TagsList", &TagsList{}, false},
		{"TagPage", &TagPage{}, false},
		{"NotesList", &NotesList{}, false},
		{"NoteDetail", &NoteDetail{}, false},
		{"NoteEditor", &NoteEditor{}, false},
		{"SermonsList", &SermonsList{}, false},
		{"SermonBuilder", &SermonBuilder{}, false},
		{"VerseHub", &VerseHub{}, false},
	}

	for _, tc := range views {
		t.Run(tc.name, func(t *testing.T) {
			poison(reflect.ValueOf(tc.view).Elem(), 0)

			// Render the pointee: the views implement Render on value
			// receivers, and the pointer was only needed so poison could
			// set fields.
			b := element.NewBuilder()
			reflect.ValueOf(tc.view).Elem().Interface().(element.Component).Render(b)
			assertEscaped(t, tc.name, b.String(), !tc.noUserText)
		})
	}

	// The shell is not a view struct poison can fill blindly: Body is an
	// interface and Active/Translation select nav state. Its user-supplied
	// input is the title (a note or tag name), so that is what is poisoned.
	t.Run("Page", func(t *testing.T) {
		b := element.NewBuilder()
		Page{Title: hostilePayload, Body: ErrorPage{Status: 404, Heading: "x"}}.Render(b)
		assertEscaped(t, "Page", b.String(), true)
	})
}

func assertEscaped(t *testing.T, name, got string, wantPayload bool) {
	t.Helper()
	if strings.Contains(got, "<script>pwn") {
		// Print only the neighbourhood of the first hit; a whole page of
		// HTML buries the line that matters.
		i := strings.Index(got, "<script>pwn")
		lo, hi := max(0, i-160), min(len(got), i+60)
		t.Errorf("%s rendered a user string unescaped:\n…%s…", name, got[lo:hi])
	}
	if wantPayload && !strings.Contains(got, "&lt;script&gt;pwn") {
		// Guards the test itself: if the payload never reached the output
		// at all (every field skipped, every branch empty), the absence check
		// above would pass vacuously.
		t.Errorf("%s never rendered the poisoned text, so the check above proved nothing", name)
	}
}

// poison fills v in place. depth bounds the recursion so a self-referential
// type added later cannot hang the test; nothing in the views nests anywhere
// near this deep.
func poison(v reflect.Value, depth int) {
	if depth > 8 || !v.CanSet() {
		return
	}

	// Types with a meaning poison must not overwrite with junk.
	switch v.Type() {
	case reflect.TypeFor[time.Time]():
		v.Set(reflect.ValueOf(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)))
		return
	case reflect.TypeFor[*bible.Book]():
		// Book names come from the compiled canon table, not from a user, so
		// a real book is the honest input; poisoning it would flag escaping
		// the package deliberately does not do.
		v.Set(reflect.ValueOf(bible.MustByNum(1)))
		return
	case reflect.TypeFor[bible.Ref]():
		v.Set(reflect.ValueOf(bible.Ref{BookNum: 1, Chapter: 1, VerseStart: 1}))
		return
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(hostilePayload)
	case reflect.Int, reflect.Int64:
		v.SetInt(1)
	case reflect.Bool:
		// true reveals more of the page (AIEnabled, IsExisting, Sync.Enabled)
		// and therefore more places to forget esc().
		v.SetBool(true)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			name := v.Type().Field(i).Name
			if rawHTMLFields[name] {
				continue
			}
			if safe, ok := trustedFields[v.Type().Name()+"."+name]; ok && v.Field(i).CanSet() {
				v.Field(i).SetString(safe)
				continue
			}
			poison(v.Field(i), depth+1)
		}
	case reflect.Slice:
		s := reflect.MakeSlice(v.Type(), 1, 1)
		poison(s.Index(0), depth+1)
		v.Set(s)
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		k := reflect.New(v.Type().Key()).Elem()
		e := reflect.New(v.Type().Elem()).Elem()
		poison(k, depth+1)
		poison(e, depth+1)
		m.SetMapIndex(k, e)
		v.Set(m)
	case reflect.Pointer:
		p := reflect.New(v.Type().Elem())
		poison(p.Elem(), depth+1)
		v.Set(p)
	}
}

// Compile-time reminder of what the walk relies on: study types reach the
// views by value, and their user text lives in exported string fields.
var _ = study.Note{Title: hostilePayload}
