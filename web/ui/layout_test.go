package ui

import (
	"strings"
	"testing"

	"github.com/rohanthewiz/element"

	"github.com/rohanthewiz/pbstudy/blb"
)

// TestTaggerConfigFollowsTagger pins the ScriptTagger ordering contract.
//
// The tagger replaces BLB.Tagger with its defaults when it executes, so the
// config script must come AFTER the tagger's tag (both deferred, so they run
// in document order), and no inline BLB.Tagger assignment may precede it —
// that shape was silently discarded, which is how popups came to quote NKJV
// and scripture containers came to be tagged.
func TestTaggerConfigFollowsTagger(t *testing.T) {
	b := element.NewBuilder()
	Page{Title: "x", Translation: "asv", Body: ErrorPage{Status: 404, Heading: "x"}}.Render(b)
	got := b.String()

	tagger := strings.Index(got, `src="`+blb.ScriptTaggerURL+`"`)
	config := strings.Index(got, `src="`+blb.TaggerConfigScript+`"`)
	if tagger < 0 || config < 0 {
		t.Fatalf("tagger script at %d, config script at %d; want both present:\n%s", tagger, config, got)
	}
	if config < tagger {
		t.Errorf("config script precedes the tagger; the tagger would overwrite it")
	}
	if strings.Contains(got, "BLB.Tagger =") {
		t.Errorf("an inline BLB.Tagger assignment is back; the tagger discards it")
	}
	for _, want := range []string{`data-translation="ASV"`, `data-no-tag="` + blb.NoTagClass + `"`} {
		if !strings.Contains(got[config:], want) {
			t.Errorf("config script tag lacks %s", want)
		}
	}
}
