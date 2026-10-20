package web

import (
	"testing"

	"github.com/rohanthewiz/pbstudy/bible"
)

// TestPickTranslation covers the fallback when the saved preference cannot be
// honoured — most often because that translation's cache was removed.
func TestPickTranslation(t *testing.T) {
	all := []bible.Translation{{Abbrev: "kjv"}, {Abbrev: "web"}, {Abbrev: "asv"}}
	onlyWeb := []bible.Translation{{Abbrev: "web"}}

	cases := []struct {
		name       string
		pref       string
		downloaded []bible.Translation
		want       string
	}{
		{"no preference takes the first downloaded", "", all, "kjv"},
		{"a downloaded preference wins", "asv", all, "asv"},
		{"a preference no longer downloaded falls back", "kjv", onlyWeb, "web"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := pickTranslation(c.pref, c.downloaded); got != c.want {
				t.Errorf("pickTranslation(%q) = %q, want %q", c.pref, got, c.want)
			}
		})
	}
}
