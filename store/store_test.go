package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rohanthewiz/pbstudy/cfg"
)

// openTestStore opens a whole Store (both databases, the directory lock) over
// a fresh data directory, the way every command does.
func openTestStore(t *testing.T) *Store {
	t.Helper()

	st, err := Open(cfg.Config{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

// TestBackupStudyLeavesOnlyTheSnapshot: the backups folder sits in the sync
// directory, so anything written there beside the snapshot is uploaded to
// every machine. bytdb's Backup leaves a <dest>.lock sidecar; BackupStudy
// must clear it.
func TestBackupStudyLeavesOnlyTheSnapshot(t *testing.T) {
	st := openTestStore(t)
	dest := t.TempDir()

	path, err := st.BackupStudy(dest)
	if err != nil {
		t.Fatalf("BackupStudy() error: %v", err)
	}

	entries, err := os.ReadDir(dest)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if len(names) != 1 || names[0] != filepath.Base(path) {
		t.Errorf("backup folder holds %v, want only %s", names, filepath.Base(path))
	}
}

// TestBackupStudyThroughSymlink: btypedb names the sidecar after the
// symlink-resolved path, so a backups folder reached through a link must
// still have its sidecar found and removed.
func TestBackupStudyThroughSymlink(t *testing.T) {
	st := openTestStore(t)
	real := t.TempDir()
	link := filepath.Join(t.TempDir(), "backups")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("cannot create symlink here: %v", err)
	}

	if _, err := st.BackupStudy(link); err != nil {
		t.Fatalf("BackupStudy() error: %v", err)
	}

	entries, err := os.ReadDir(real)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".lock") {
			t.Errorf("lock sidecar %s left in the symlinked backups folder", e.Name())
		}
	}
}

// TestCompactStudyShrinksAfterDeletes is the reason CompactStudy exists:
// DELETE appends to bytdb's log, so a purge alone grows the file. Compaction
// has to bring it back below where it started, without losing a live row.
func TestCompactStudyShrinksAfterDeletes(t *testing.T) {
	st := openTestStore(t)
	now := time.Now().UTC()

	// Enough rows, with enough body, that the deleted ones dominate the file;
	// a handful of tiny rows would leave compaction's saving inside the
	// noise of the catalog records.
	body := strings.Repeat("grace upon grace ", 64)
	for i := range 200 {
		if _, err := st.Study.Exec(
			`INSERT INTO notes (id, title, body_md, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
			fmt.Sprintf("n%03d", i), "note", body, now, now); err != nil {
			t.Fatalf("insert %d: %v", i, err)
		}
	}
	if _, err := st.Study.Exec(`DELETE FROM notes WHERE id <> $1`, "n000"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	before, after, err := st.CompactStudy()
	if err != nil {
		t.Fatalf("CompactStudy() error: %v", err)
	}
	if before == 0 || after == 0 {
		t.Fatalf("CompactStudy() sizes = %d → %d, want both non-zero", before, after)
	}
	if after >= before {
		t.Errorf("CompactStudy() did not shrink the file: %d → %d bytes", before, after)
	}

	// The survivor is still there, through the same handle: compaction swaps
	// the log underneath an open engine.
	var n int
	if err := st.Study.QueryRow(`SELECT COUNT(*) FROM notes`).Scan(&n); err != nil {
		t.Fatalf("count after compaction: %v", err)
	}
	if n != 1 {
		t.Errorf("after compaction notes holds %d rows, want 1", n)
	}
}
