package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
