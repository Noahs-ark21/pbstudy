// Package store owns database lifecycle: opening the two bytdb files,
// bootstrapping their schemas, and taking snapshot backups.
//
// # Why two databases
//
//	bible.bytdb  — ~31k rows per translation of public-domain scripture.
//	               Bulky, wholly rebuildable from getbible.net, and therefore
//	               never synced and never backed up.
//	study.bytdb  — notes, tags, cross-references, sermons. Small, irreplaceable.
//	               This is what gets exported to JSON and snapshotted.
//
// Splitting them means a backup is small enough to take often, and a corrupt
// or stale scripture cache can be deleted without a second thought.
package store

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/rohanthewiz/bytdb/stdlib"
	"github.com/rohanthewiz/logger"
	"github.com/rohanthewiz/serr"

	"github.com/rohanthewiz/pbstudy/cfg"
)

// Store holds the open database handles for the process lifetime.
//
// bytdb takes the database file for itself — one path is one engine per
// process — so these handles are opened once at startup and shared. *sql.DB
// is safe for concurrent use; each pooled connection gets its own bytdb
// session.
type Store struct {
	// Bible is the scripture cache (read-heavy; written only by `download`).
	Bible *sql.DB

	// Study is the notes/tags/xrefs/sermons database.
	Study *sql.DB

	// lock keeps a second pbstudy process out of this data directory.
	lock *dirLock

	cfg cfg.Config
}

// Open opens both databases and brings their schemas up to date.
//
// The bible DB is opened with sync=never: every row in it can be re-fetched
// with `pbstudy download`, so trading crash durability for a materially
// faster 31k-row bulk import is the right side of that bargain. The study DB
// keeps bytdb's default fsync behaviour — that data has no upstream.
func Open(conf cfg.Config) (*Store, error) {
	st := &Store{cfg: conf}

	// Take the data-directory lock before touching either database.
	//
	// Since v0.17 bytdb also locks each file it opens (an flock on a
	// <db>.lock sidecar), so a second process can no longer interleave
	// appends into study.bytdb even without this. The directory lock is kept
	// because it still does two things bytdb's cannot: it claims both
	// databases as one unit before either is opened, so a refused second
	// process never gets as far as half-opening the pair, and it fails with
	// a message that names pbstudy and the holding PID rather than a
	// bytdb ErrLocked from inside sql.Open.
	lock, err := acquireDirLock(conf.DataDir)
	if err != nil {
		return nil, err
	}
	st.lock = lock

	bibleDB, err := sql.Open("bytdb", conf.BibleDBPath()+"?sync=never")
	if err != nil {
		_ = st.Close()
		return nil, serr.Wrap(err, "cannot open bible db", "path", conf.BibleDBPath())
	}
	st.Bible = bibleDB

	studyDB, err := sql.Open("bytdb", conf.StudyDBPath())
	if err != nil {
		_ = st.Close()
		return nil, serr.Wrap(err, "cannot open study db", "path", conf.StudyDBPath())
	}
	st.Study = studyDB

	if err := bootstrapBible(bibleDB); err != nil {
		_ = st.Close()
		return nil, serr.Wrap(err, "cannot bootstrap bible schema")
	}
	if err := bootstrapStudy(studyDB); err != nil {
		_ = st.Close()
		return nil, serr.Wrap(err, "cannot bootstrap study schema")
	}

	return st, nil
}

// Close releases both database handles and the data-directory lock.
//
// Every step is attempted even if an earlier one fails — the first error is
// reported, but a failed database close must not leave the lock held for the
// life of the process.
func (s *Store) Close() error {
	var firstErr error
	fail := func(err error, what string) {
		if err != nil && firstErr == nil {
			firstErr = serr.Wrap(err, what)
		}
	}

	if s.Bible != nil {
		fail(s.Bible.Close(), "closing bible db")
		s.Bible = nil
	}
	if s.Study != nil {
		fail(s.Study.Close(), "closing study db")
		s.Study = nil
	}
	if s.lock != nil {
		fail(s.lock.release(), "releasing data directory lock")
		s.lock = nil
	}
	return firstErr
}

// BackupStudy writes a consistent snapshot of the study database to destDir,
// named study-<RFC3339-ish timestamp>.bytdb, and returns the path written.
//
// This is the ONLY safe way to get a copy of a live bytdb file. bytdb's file
// lock is advisory: it keeps other bytdb openers out, but a sync daemon (or a
// plain `cp`) never asks for it, so one that reads the live file mid-append
// still captures a torn tail. Engine.Backup takes an internal snapshot and
// writes a self-consistent file, which is safe to hand to iCloud.
//
// Since bytdb v0.18, Backup also locks its destination (refusing one that is
// a live database), and like every bytdb lock that leaves a small
// study-<stamp>.bytdb.lock sidecar beside the snapshot. It was harmless (the
// lock is the flock, not the file, and the backup listing only reads *.bytdb
// names), but the backups folder lives in the sync directory, so every stray
// sidecar was uploaded to every machine. BackupStudy removes it once Backup
// has returned; see removeLockSidecar for why that is safe here when btypedb
// deliberately never does it in general.
func (s *Store) BackupStudy(destDir string) (string, error) {
	if destDir == "" {
		return "", serr.New("no backup destination configured")
	}
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return "", serr.Wrap(err, "cannot create backup dir", "dir", destDir)
	}

	// Reach past database/sql to the native engine — Backup() has no
	// SQL-level equivalent.
	eng, err := stdlib.Engine(context.Background(), s.Study)
	if err != nil {
		return "", serr.Wrap(err, "cannot reach bytdb engine for backup")
	}

	// Colons are legal on the filesystems we care about but make paths
	// awkward to type and illegal on FAT/exFAT sync volumes.
	stamp := time.Now().UTC().Format("20060102-150405")
	dest := filepath.Join(destDir, "study-"+stamp+".bytdb")

	if err := eng.Backup(dest); err != nil {
		return "", serr.Wrap(err, "backup failed", "dest", dest)
	}

	// The snapshot is complete and in place; a sidecar that cannot be
	// removed is clutter, not a failed backup, so it is logged and the
	// backup still reported as written.
	if err := removeLockSidecar(dest); err != nil {
		logger.LogErr(err, "backup written, but its lock sidecar was left behind", "dest", dest)
	}
	return dest, nil
}

// removeLockSidecar deletes the <dest>.lock file bytdb's Backup leaves beside
// a fresh snapshot.
//
// btypedb keeps sidecars on purpose (btypedb lock.go): unlinking one while
// another process is between opening and locking it lets two processes each
// lock a different inode for the same database. That race needs a second
// party opening the same path. Nothing opens a fresh snapshot: the name is
// minted here from the current second, the file is not a live database, and
// restoring one means renaming it to study.bytdb (whose own sidecar is
// unrelated). Backup has already released its lock by the time it returns, so
// the file is unheld when it is removed.
//
// The path is resolved through symlinks first because that is how btypedb
// names the sidecar (lockPathFor): a backups folder reached through a
// symlink gets its sidecar beside the real file, not beside the link.
func removeLockSidecar(dest string) error {
	resolved := dest
	if real, err := filepath.EvalSymlinks(dest); err == nil {
		resolved = real
	}
	sidecar := resolved + ".lock"
	if err := os.Remove(sidecar); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return serr.Wrap(err, "cannot remove backup lock sidecar", "path", sidecar)
	}
	return nil
}

// CompactStudy rewrites the study database's log as a minimal snapshot of
// the live rows, and reports the file size before and after.
//
// bytdb's file is append-only: a DELETE appends a tombstone rather than
// freeing space, so `pbstudy compact`, which deletes rows, used to leave the
// file larger than it found it. Engine.Compact is the same pass bytdb's
// background auto-compaction runs once the log outgrows its threshold;
// calling it here makes the command shrink the file it just purged instead of
// waiting for that heuristic. It is the native-engine form of the SQL VACUUM
// statement, reached the same way BackupStudy reaches Backup. Like VACUUM it
// must not run inside a transaction; it takes none, and readers and writers
// keep running (bytdb only pauses writers briefly).
//
// The sizes are best-effort: a failed Stat reports 0 rather than failing a
// compaction that succeeded.
func (s *Store) CompactStudy() (before, after int64, err error) {
	path := s.cfg.StudyDBPath()
	before = fileSize(path)

	eng, err := stdlib.Engine(context.Background(), s.Study)
	if err != nil {
		return before, before, serr.Wrap(err, "cannot reach bytdb engine for compaction")
	}
	if err := eng.Compact(); err != nil {
		return before, before, serr.Wrap(err, "study database compaction failed", "path", path)
	}
	return before, fileSize(path), nil
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}
