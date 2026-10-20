package bible

import (
	"database/sql"
	"errors"

	"github.com/rohanthewiz/serr"
)

// prefTranslation is the prefs key holding the user's preferred translation.
const prefTranslation = "translation"

// PreferredTranslation returns the translation the user chose on the settings
// page, or "" when none has been chosen. The value is returned as stored; the
// caller decides what to do if that translation is no longer downloaded.
func PreferredTranslation(db *sql.DB) (string, error) {
	var abbrev string
	err := db.QueryRow(`SELECT value FROM prefs WHERE key = $1`, prefTranslation).Scan(&abbrev)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", serr.Wrap(err, "cannot read preferred translation")
	}
	return abbrev, nil
}

// SetPreferredTranslation records abbrev as the preferred translation. Only
// known translations are accepted; the value is reflected into URLs and
// forms, so it is validated on the way in.
func SetPreferredTranslation(db *sql.DB, abbrev string) error {
	if !IsKnownTranslation(abbrev) {
		return serr.New("unknown translation", "translation", abbrev)
	}

	tx, err := db.Begin()
	if err != nil {
		return serr.Wrap(err, "cannot begin transaction")
	}
	defer func() { _ = tx.Rollback() }()

	// Delete-then-insert, as in recordTranslation: "replace this row" with no
	// dependence on upsert semantics.
	if _, err := tx.Exec(`DELETE FROM prefs WHERE key = $1`, prefTranslation); err != nil {
		return serr.Wrap(err, "cannot clear preferred translation")
	}
	if _, err := tx.Exec(`INSERT INTO prefs (key, value) VALUES ($1, $2)`,
		prefTranslation, abbrev); err != nil {
		return serr.Wrap(err, "cannot record preferred translation")
	}
	if err := tx.Commit(); err != nil {
		return serr.Wrap(err, "cannot commit preferred translation")
	}
	return nil
}
