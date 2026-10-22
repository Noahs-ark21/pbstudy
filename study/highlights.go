package study

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/rohanthewiz/serr"

	"github.com/rohanthewiz/pbstudy/bible"
)

// HighlightColors are the colours a verse can be marked with, in the order the
// verse hub offers them. Stored by name, not by CSS value, so the palette can
// be retuned without rewriting anyone's data or sync files.
var HighlightColors = []string{"yellow", "green", "blue", "pink"}

// IsHighlightColor reports whether c is one of HighlightColors. The colour
// arrives from a form and from sync files, and flows into a class attribute,
// so it is checked against the list rather than escaped.
func IsHighlightColor(c string) bool {
	for _, known := range HighlightColors {
		if c == known {
			return true
		}
	}
	return false
}

// highlightID derives a highlight's id from its verse.
//
// Every other syncable row gets a minted UUID (see newID), because two notes
// on the same verse are two different things. Two highlights on the same verse
// are the same thing — the verse's colour — so the id names the verse. That
// makes the sync engine's last-writer-wins do the right thing unaided: both
// machines write "hl-43-3-16", and the newer colour (or the newer clear) wins.
// A minted id would instead leave two live rows for one verse, the duplicate
// problem tags have (N-005).
func highlightID(bookNum, chapter, verse int) string {
	return fmt.Sprintf("hl-%d-%d-%d", bookNum, chapter, verse)
}

// SetHighlight colours one verse, replacing any colour it already had.
//
// A tombstoned row for the same verse is revived rather than duplicated — it
// has to be, since the id is the verse.
func SetHighlight(db *sql.DB, at bible.Ref, color string) error {
	if at.BookNum < 1 || at.Chapter < 1 || at.VerseStart < 1 {
		return serr.New("a highlight needs a specific verse")
	}
	if !IsHighlightColor(color) {
		return serr.New("unknown highlight colour", "color", color)
	}

	id := highlightID(at.BookNum, at.Chapter, at.VerseStart)
	now := nowUTC()

	res, err := db.Exec(
		`UPDATE highlights SET color = $1, updated_at = $2, deleted_at = NULL
		  WHERE id = $3`, color, now, id)
	if err != nil {
		return serr.Wrap(err, "cannot update highlight", "verse", at.String())
	}
	if n, err := res.RowsAffected(); err == nil && n > 0 {
		return nil
	}

	if _, err := db.Exec(
		`INSERT INTO highlights (id, book_num, chapter, verse, color,
		     created_at, updated_at, deleted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, at.BookNum, at.Chapter, at.VerseStart, color, now, now, nil); err != nil {
		return serr.Wrap(err, "cannot create highlight", "verse", at.String())
	}
	return nil
}

// ClearHighlight removes a verse's colour. Like every delete here it is a
// tombstone, so the clear reaches the other machines. Clearing a verse that is
// not highlighted is not an error: the verse already looks the way the user
// asked for.
func ClearHighlight(db *sql.DB, at bible.Ref) error {
	if _, err := db.Exec(
		`UPDATE highlights SET deleted_at = $1, updated_at = $1
		  WHERE id = $2 AND deleted_at IS NULL`,
		nowUTC(), highlightID(at.BookNum, at.Chapter, at.VerseStart)); err != nil {
		return serr.Wrap(err, "cannot clear highlight", "verse", at.String())
	}
	return nil
}

// HighlightOf returns one verse's colour, or "" when it has none.
func HighlightOf(db *sql.DB, bookNum, chapter, verse int) (string, error) {
	var color string
	err := db.QueryRow(
		`SELECT color FROM highlights WHERE id = $1 AND deleted_at IS NULL`,
		highlightID(bookNum, chapter, verse)).Scan(&color)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", serr.Wrap(err, "cannot read highlight",
			"book", strconv.Itoa(bookNum), "chapter", strconv.Itoa(chapter),
			"verse", strconv.Itoa(verse))
	}
	return color, nil
}

// ChapterHighlights returns the colour of every highlighted verse in one
// chapter, keyed by verse number. One query per chapter render, like
// ChapterMarks.
func ChapterHighlights(db *sql.DB, bookNum, chapter int) (map[int]string, error) {
	rows, err := db.Query(
		`SELECT verse, color FROM highlights
		  WHERE book_num = $1 AND chapter = $2 AND deleted_at IS NULL`,
		bookNum, chapter)
	if err != nil {
		return nil, serr.Wrap(err, "cannot read chapter highlights",
			"book", strconv.Itoa(bookNum), "chapter", strconv.Itoa(chapter))
	}
	defer rows.Close()

	out := map[int]string{}
	for rows.Next() {
		var verse int
		var color string
		if err := rows.Scan(&verse, &color); err != nil {
			return nil, serr.Wrap(err, "cannot scan highlight row")
		}
		// A colour this build does not know (a newer machine's palette) is
		// left out rather than rendered as an unstyled class.
		if IsHighlightColor(color) {
			out[verse] = color
		}
	}
	if err := rows.Err(); err != nil {
		return nil, serr.Wrap(err, "cannot iterate highlight rows")
	}
	return out, nil
}
