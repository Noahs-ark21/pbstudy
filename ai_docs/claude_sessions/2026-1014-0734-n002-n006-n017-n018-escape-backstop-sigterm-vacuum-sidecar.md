# pbstudy — `/work-next-list`, first batch: escaping backstop, SIGTERM close, compact shrinks the file, no backup sidecar

Session id: `ed96adb0-6fc8-411a-ba39-2f8f89b061e5`

An unattended run over `ai_docs/todo/next-list.md`. This doc covers the first
four items: N-002, N-006, N-018 and N-017. The run continues after this wrap
with N-019, N-005 and N-001.

---

## Starting point

- `main` at `868dbca`, clean. Build, vet and tests green.
- `staticcheck` at `~/bin` was built with go1.23 and cannot load a go1.26
  module, so lint was skipped. The repo has no other linter configured.
- No `ANTHROPIC_API_KEY` in the environment.

## Triage

| Doable | Plan |
| --- | --- |
| N-002 | reflection-driven escaping test over every view |
| N-006 | install the shutdown handler without sync too |
| N-017 | `Engine.Compact` at the end of a real `compact` |
| N-018 | remove the `.lock` sidecar after `Backup` |
| N-019 | model constant → `claude-sonnet-5-5` |
| N-005 | re-check the premise, then the cleanup if it is safe |
| N-001 | try the Chrome extension against a scratch server |

Skipped: **N-003**, deferred on purpose until it "ever shows up". **N-004**
needs a real API key. **N-007** describes accepted behaviour and proposes no
fix; whether it moves to Non-goals is the user's call.

---

## N-002: the escaping backstop (`f958d8e`)

`web/ui/escape_test.go` / `TestEveryViewEscapes`:

- It walks each view struct (13 views plus the `Page` shell) with
  reflection. Every reachable exported string is set to
  `"'><script>pwn()</script>`. Ints are set to 1, so book, chapter and verse
  numbers stay valid for `bible.MustByNum`. Bools are set to true, which shows
  more of the page. Each slice or map gets one element. `*bible.Book` gets
  Genesis, because book names come from the canon table.
- The test fails on any literal `<script>pwn` in the output. A second guard
  fails if the escaped payload never appears at all, so an empty render can't
  pass. `Dashboard` is exempt from that guard because it renders no user
  text.
- `rawHTMLFields` (`BodyHTML`, `DraftHTML`) holds goldmark output and is
  skipped.
- `trustedFields` is keyed `Type.Field`. Each entry gets a realistic value,
  and a comment says where its constraint is enforced:
  - translations, checked by `IsKnownTranslation`
  - `Search.Scope`, passed through `NormalizeScope`
  - URLs built by the handlers
  - the syncer's report strings, which are built from counts

The first run found four real sinks:

1. **`ErrorPage`** printed `Heading` and `Detail` unescaped, and the handlers
   echo URL segments into them (the unknown-translation and "no such script"
   messages). Browsers percent-encode `<` in paths, so this was latent. On a
   scratch server, a raw `curl /read/<b>x/1/1` showed it is real for a
   non-browser client. Both fields are now escaped.
2. **Settings `kv()`** printed environment paths unescaped. `kv()` now
   escapes both its key and its value. The other callers pass plain text,
   which esc() leaves unchanged.
3. **`kindLabel`**'s fallback returned an unknown section kind as is. That
   value can only come from a sync file. It is now `esc(kind)`.
4. **Record ids in hrefs and form actions.** Sync import accepts any non-empty
   id. Added `pathSeg()` (`url.PathEscape`) and applied it in
   `NoteURL`/`TagURL`/`SermonURL`, the section buttons and the xref delete
   action.

   *Decision:* escape ids at the URL helpers rather than reject non-UUID ids
   at import.
   - Import deliberately repairs rather than rejects (see `decodeOutline`).
   - The sync tests use ids like `"n1"`.
   - The helpers also feed redirect `Location` headers. Percent-encoding is
     safe there, and `esc()` would not be.
   - `pathSeg()` is a no-op on UUIDs.

## N-006: SIGTERM closes the databases without sync (`3406773`)

- `installShutdownFlush` now accepts a nil syncer. It skips the flush and
  still closes the store.
- `cmdServe` installs it outside the `if sync != nil` block.
- The comment says plainly that this is tidiness, not a data-loss fix. bytdb's
  default durability already loses nothing, and the kernel drops the flocks
  when the process dies.
- Verified on a scratch server with no `PBSTUDY_SYNC_DIR`: `SIGTERM` printed
  `Stopping (terminated) ...` and the process exited.

## N-018: no lock sidecar beside backups (`a36d4c0`)

- `BackupStudy` calls `removeLockSidecar(dest)` after `Backup` returns. The
  sidecar is `EvalSymlinks(dest)+".lock"`, which is how btypedb's
  `lockPathFor` names it.
- Why it is safe: btypedb keeps sidecars because of a race with a second
  process opening the same path. Nothing opens a freshly minted snapshot, and
  `Backup` has released its lock by the time it returns.
- A failed removal is logged and does not fail the backup.
- `store/store_test.go` has two tests: the plain case and a symlinked backups
  folder. Both were confirmed to fail with the removal disabled.
- End to end: `pbstudy backup` on the scratch sync dir left only
  `study-<stamp>.bytdb`.

## N-017: `compact` shrinks the file it purged (`4ac757e`)

- New `Store.CompactStudy() (before, after int64, err error)` calls
  `stdlib.Engine` and then `Engine.Compact`. That is the native form of
  `VACUUM`, the same route `BackupStudy` takes, and it runs outside any
  transaction.
- `cmdCompact` runs it on every non-dry-run pass and prints
  `Study database  N → M bytes`.
  - It runs even when no rows were removed, because ordinary edits also grow
    the log.
  - A failure here is the command's error, so the exit code is honest.
  - The dry run skips it.
- Scratch run (6 notes created, 4 deleted, then `compact 0`): 14,038 bytes
  after the edits, 16,595 after the purge, 5,346 after compaction.
- Added `TestCompactStudyShrinksAfterDeletes`.
- README's Compacting example, the README usage block and the `usage()` text
  were updated.
- N-012's text (Non-goals) now records that `compact` compacts the file. A
  database-only mode is still declined.

---

## State at end of this batch

- `main` at `4ac757e` plus this doc, pushed. All tests green; `go vet` clean.
- Scratch servers were stopped by PID. The scratch data dirs are in the
  session scratchpad only.

## Next

Closed: N-002, N-006, N-017, N-018. Declined: None. Raised: None.
Deferred: None. Promoted: None. Moved: None.
Updated: N-012 (premise note: `compact` now shrinks the file).
Full list: `ai_docs/todo/next-list.md`.
