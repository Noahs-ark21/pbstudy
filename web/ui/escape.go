package ui

import (
	"html"
	"net/url"
)

// The element builder does not escape anything.
//
// Both b.T("...") and the attribute values passed to b.Div("title", "...") are
// written to the output buffer verbatim. That is a reasonable choice for a
// library whose job is to emit exactly the markup you describe, but it means
// escaping is the caller's responsibility — ours — and getting it wrong is an
// injection hole rather than a cosmetic bug.
//
// The rule for this package, with no exceptions:
//
//	Anything that originated outside the binary goes through esc().
//
// That covers note titles and bodies, tag names and descriptions,
// cross-reference comments, search queries echoed back, and scripture text
// (which comes from getbible.net over the network, so it is not ours either).
// Text that is a Go string literal, a strconv.Itoa result, or a value from the
// compiled canon table does not need it.
//
// Rendered Markdown is the one deliberate exception: study.RenderMarkdown
// returns HTML that goldmark already escaped, and it is written with
// b.WriteString rather than b.T. See renderMarkdownBody.

// esc escapes a string for either HTML text or a double-quoted attribute
// value. html.EscapeString covers both contexts because it escapes the quote
// characters as well as the angle brackets and ampersand.
//
// It is a one-line wrapper on purpose: the short name is what makes the calls
// short enough that wrapping every user-supplied string stays readable inside
// a render tree, and a named function is greppable in a way that a scattered
// html.EscapeString is not.
func esc(s string) string { return html.EscapeString(s) }

// urlQueryEscape percent-encodes a value destined for a query string.
//
// Distinct from esc() and not interchangeable with it: a reference like
// "John 3:16" going into ?refs= needs percent-encoding, and a URL built this
// way is then placed in an href attribute where the ampersands separating
// parameters must survive as-is — which is why the escaping happens here, on
// the value, rather than on the assembled URL.
func urlQueryEscape(s string) string { return url.QueryEscape(s) }

// pathSeg percent-encodes one path segment — in practice a record id spliced
// into a URL like "/notes/<id>/edit".
//
// Ids are minted by uuid.NewString, so for every id this binary created the
// call is a no-op. It exists for the ones it did not: sync imports accept any
// non-empty id a record file carries (study.ApplyNote and friends), and a
// hand-edited or hostile file in the sync folder could carry `"><script>`.
// Escaping at the URL helpers was chosen over rejecting non-UUID ids at import
// because import is deliberately permissive (it repairs rather than rejects,
// see decodeOutline), and because the helpers' output lands in two places —
// href/action attributes and redirect Location headers — that percent-encoding
// makes safe at once, where esc() would only cover the first.
//
// url.PathEscape encodes the quote characters and angle brackets (and '/', so
// an id cannot add path segments); it leaves '&' alone, which is harmless
// inside a quoted attribute value.
func pathSeg(id string) string { return url.PathEscape(id) }
