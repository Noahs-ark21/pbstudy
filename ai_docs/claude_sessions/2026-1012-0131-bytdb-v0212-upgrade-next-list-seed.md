# pbstudy — seeding the living Next list; bytdb v0.9.1 → v0.21.2

Session id: `caac009b-56d4-44ff-bcf0-debd7f6d83f7`

Two pieces of work: `/next-list seed` turned the seven session docs' "Carried
forward" sections into a living list at `ai_docs/todo/next-list.md`, and the
bytdb dependency moved twelve minor versions, proved against a database the
old engine wrote.

---

## Starting point

- `main` at `f1cc6f6`, clean. bytdb pinned at v0.9.1 (btypedb v0.7.0).
- No `ai_docs/todo/` yet. Every session doc ended in "State at end of session
  / Carried forward" rather than `## Next`, and each carried its list forward
  by hand.

---

## Part 1: the living Next list

`/next-list seed` ran History mode over all 7 docs (`2026-0805-1823` →
`2026-0807-0954`), so no age was floored by the window.

### What the rebuild caught

- **Two silent lapses.**
  - The `web/ui/escape.go` "no framework backstop" item ran from Phase 2 to
    Phase 4, then vanished from Phase 5's list without being done. Restored
    as **N-002**. `element` v0.6.0's `T()` still writes strings verbatim
    (`builder_funcs.go:56`), so the premise holds.
  - The `ChapterMarks`-per-render item appeared only in Phase 2. Restored
    as **N-003**. The premise holds (`web/handlers_read.go:143`).
- **One lapse that was actually done.** Phase 1's "confirm the exact model ID
  at Phase 4" was never carried past Phase 1 and never recorded as checked.
  It was checked here against the current model list: `claude-sonnet-5` is
  valid and still served. Closed as **N-013**.
- **One premise that needed reading closely.** "`status` can stick at
  `drafting`" was raised for the sync engine's sake. Phase 5 normalizes it on
  import (`study/sync.go:558`). Locally it is inert: no UI renders sermon
  status, and the draft lock is in memory. Closed as **N-016**.
- **Closed as found done:**
  - **N-014**, the DISTINCT/ORDER BY limit, fixed upstream in v0.9.1.
  - **N-015**, tombstone pruning, done by `pbstudy compact`.

### Shape of the seeded file

- **Open:** N-002, N-003, N-005, N-006, N-007.
- **Validate:** N-001 (ScriptTagger popups in a real browser) and N-004 (a
  live AI draft).
- **Roadmap:** empty, because no doc deferred anything.
- **Non-goals, N-008 to N-012:**
  - two-machine offline edits (explicitly accepted)
  - settings download buttons
  - JS outline reorder
  - a compaction button
  - a database-only compaction mode
- **Closed:** N-013 to N-016.

The informational carries were left out of the item list on purpose: the dev
sandbox path, the `Ref.Compare` and `NoteHits` seams, and `DROP INDEX IF
EXISTS` being available.

---

## Part 2: bytdb v0.9.1 → v0.21.2

`go get github.com/rohanthewiz/bytdb@v0.21.2` also took btypedb v0.7.0 →
v0.9.0. No pbstudy code changed: the build, `go vet` and every test passed
first time. The work was in proving it and correcting what the upgrade made
untrue.

### What changed upstream that pbstudy can feel

bytdb ships no changelog. A research agent read bytdb's session docs and
diffed the module source. The findings:

- **No on-disk format change.** btypedb's diff touches no header or record
  encoding, and bytdb's descriptor structs and row encoding are unchanged.
  pbstudy's whole bytdb surface is the stdlib driver, whose `stdlib.go` is
  byte-identical apart from one comment. The DSN grammar is unchanged.
- **Per-file locking (bytdb v0.17, moved into btypedb in v0.18).**
  - Each opened database gets an advisory `flock` on a `<db>.lock` sidecar.
    The sidecar is **deliberately never deleted**, because removing it would
    reopen a two-holder race. A stale sidecar never blocks anything: the
    flock dies with its process.
  - pbstudy's own `pbstudy.lock` is a different file, so the two locks don't
    interact.
- **`Backup` locks its destination (v0.18).** Every `pbstudy backup` now
  leaves `study-<stamp>.bytdb.lock` beside the snapshot in the sync folder's
  `backups/`. It's harmless: `recentBackups` lists only `*.bytdb`, and the
  syncer reads only `*.json`. It is clutter that syncs, so it's raised as
  **N-018**.
- **`VACUUM` / `Engine.Compact` (v0.19)** shrinks the file on demand.
  btypedb was already auto-compacting the log at v0.7.0 (`maybeAutoCompact`),
  so pbstudy's comments saying a DELETE "reclaims nothing on disk" were loose
  even before the upgrade.
- **SQL:** nothing pbstudy uses changed. TIMESTAMP is still microsecond. The
  DATE and BOOL input changes in v0.21.1/v0.21.2 hit no pbstudy column. The
  error-text changes matter only to string-matchers, and pbstudy has none.

### End-to-end proof: a database the old engine wrote

The tests only ever build fresh databases, so they cannot show that v0.21.2
opens a v0.9.1 file. No real database survived on this machine
(`/tmp/pbstudy-dev` was gone and `~/.pbstudy` never existed), so one was
manufactured:

1. A detached `git worktree` of HEAD built **pbstudy-old** (bytdb v0.9.1).
   The working tree built **pbstudy-new**.
2. pbstudy-old ran `download kjv` (31,102 verses, `sync=never` bulk path) and
   `serve`. The data went in through the real HTML forms:
   - 4 notes across 3 tags: a 16–18 range anchor, a chapter-level anchor,
     and one note then deleted
   - a cross-reference
   - a sermon with heading, passage and note sections
   - a sync run
3. A 23-route sweep was snapshotted from the old server. It covered the
   dashboard, notes (including the deleted note's 404), tags, reader, verse
   hubs, four search shapes (including the `John 3:16` 302 fast-path),
   sermons, both exports, settings, and all three tag pages.

| Check | Result |
| --- | --- |
| pbstudy-new on the legacy files, same 23 routes | **22 byte-identical**; `/settings` differs only by the old process's in-memory "last sync" report |
| Writes through the new engine (create, edit, delete note; xref; sermon point; sync; backup) | all 303; dashboard 3 notes · 2 xrefs · 1 sermon; zero log errors |
| `compact 0 --dry-run`, then `compact 0` | 2 tombstones + 2 files (one deleted under each engine); file 17,867 → 18,868 bytes |
| Reopen with pbstudy-new | data intact, 0 × 5xx, 0 log errors |
| **Downgrade:** pbstudy-old on a copy of the new-written file | data intact, 0 × 5xx, 0 log errors |

The downgrade check matters because it shows the upgrade is reversible: a
v0.9.1 binary still reads what v0.21.2 wrote.

Both servers were stopped with SIGTERM. Each time the PID was resolved from
the port and its command line confirmed first, so the graceful flush-and-close
path ran. The scratch worktree was removed afterwards.

### Comments and docs the upgrade made false

| File | Was | Now |
| --- | --- | --- |
| `store/store.go` (`Open`) | "bytdb does no locking of its own, so this is the only thing standing between…" | bytdb locks each file since v0.17. Why the dir lock is kept anyway: it claims both DBs as one unit before either opens, and its refusal names pbstudy and the holding PID |
| `store/store.go` (`BackupStudy`) | "bytdb does no file locking, so a sync daemon…" | the lock is advisory, and daemons and `cp` never ask for it. Documents the backup's `.lock` sidecar |
| `store/lock_unix.go` | "bytdb does no file locking of its own" | history plus the two-layer explanation |
| `main.go` usage text | "the live databases are unlocked and append-only" | "written in place, and a sync daemon ignores their locks" |
| `main.go` `cmdCompact` comment and hint | "a DELETE reclaims nothing on disk" / "append-only file does not shrink" | DELETE appends. Space returns on auto-compaction or `VACUUM`, which this command does not run |
| `store/schema.go` | "there is no `ALTER TABLE ADD PRIMARY KEY`… no second chance" | no *lone* ADD PRIMARY KEY. Since v0.14 an existing key can be replaced in one ALTER |
| `README.md` "Never copy a live database" | "bytdb does no file locking" | advisory lock, so the same three rules still hold |
| `PLAN.md` bytdb summary line | "No file locking"; "no ALTER ADD PK" | versioned statements of both |

PLAN.md's dated "Phase outcome" and "Compaction" sections were left as
written. They record what was true when each decision was made.

---

## State at end of session

| File | Change |
| --- | --- |
| `go.mod` / `go.sum` | bytdb v0.9.1 → v0.21.2; btypedb v0.7.0 → v0.9.0 (indirect) |
| `store/store.go`, `store/lock_unix.go`, `store/schema.go`, `main.go` | comments and one error hint corrected (above); no behaviour change |
| `README.md`, `PLAN.md` | locking and primary-key statements corrected |
| `ai_docs/todo/next-list.md` | new: the living list |

`gofmt` clean, `go vet` clean, `go test -count=1 ./...` all pass.

---

## Next

Closed: None. Declined: None. Raised: N-017, N-018, N-019.
Deferred: None. Promoted: None. Moved: None.
Updated: N-012. Full list: `ai_docs/todo/next-list.md`.

The list itself was seeded this session (N-001 to N-016); see Part 1.
