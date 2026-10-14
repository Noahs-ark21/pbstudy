# pbstudy — Next list

The project's one living list of follow-ups. Sessions edit this file in place
(`/sess-save`) instead of copying a Next section forward from the previous
session doc, so an item that leaves the list does so as a visible edit in git
history rather than by being left out of a paste. `/next-list` re-checks every
item's premise against the code and restores anything that lapsed.

Seeded 2026-10-12 from all seven session docs (`2026-0805-1823` through
`2026-0807-0954`), whose Next sections were titled "Carried forward".

## Conventions

- **IDs are permanent and never reused.** `N-###`, one sequence for the whole
  file. A closed or declined item keeps its ID.
- **`raised`** is the stem of the session doc an item first appeared in, even
  if a later doc reworded it.
- **Age is computed, never stored:** the number of session docs since `raised`
  (an item raised in the newest doc has age 0). It counts sessions, not days.
- **Value** is the payoff of doing the item, not the effort:
  - `high`: something is being worked around today, or a second independent
    consumer has arrived
  - `medium`: it blocks one named thing, or it is a visible defect nobody has
    to route around yet
  - `low`: a gap nobody has bumped into, or contingent on something that does
    not exist
- **Sections:**
  - **Open** is the next build work: code to write, a defect to fix, a
    decision to make.
  - **Validate** is the next testing: run, look, hear, measure, or write or
    repair a test. No product change is planned unless the check finds a
    defect.
  - **Roadmap** is wanted, but later.
  - **Non-goals** is likely never, kept so the decision stays visible.
- **Nothing leaves Open, Validate or Roadmap without a line in another
  section** (Closed, Non-goals, or a move between the three).
- **Open, Validate and Roadmap stay in ID order.** Never renumber.

**Next ID:** N-020

## Open

- **N-003** · raised `2026-0805-1917-phase2-notes-tags-xrefs` · value low
  The reader loads `ChapterMarks` on every chapter render
  (`web/handlers_read.go:143`). Fine at present scale; if it ever shows up, the
  fix is a per-chapter cache invalidated on study writes, not a schema change.
  Lapsed once: dropped from `2026-0805-1935` without being done.

- **N-005** · raised `2026-0805-2048-phase5-sync` · value low
  Live-tag duplicate files still accumulate, one per machine per name.
  Compaction clears them once they are *tombstoned* (the orphan branch in
  `syncer/compact.go`), but a shared live tag keeps its redundant file on every
  machine. A different cleanup than compaction: remove a file whose id no
  longer exists anywhere.

- **N-006** · raised `2026-0805-2048-phase5-sync` · value low
  Only the syncer path installs a signal handler (`installShutdownFlush`,
  called at `main.go:147` inside the sync branch). With no sync directory
  configured, `SIGTERM` still kills the process without closing the databases,
  so the deferred `st.Close()` at `main.go:73` never runs. bytdb's default
  durability covers it.

- **N-007** · raised `2026-0805-2238-compaction` · value low
  Compaction is per machine and has to be run on each. A machine that has not
  compacted re-exports its tombstone files; the others import and remove them
  again next pass. Bounded churn, converges once both compact, and what comes
  back is a tombstone, invisible to every read in `study/`.

- **N-017** · raised `2026-1012-0131-bytdb-v0212-upgrade-next-list-seed` · value low
  `pbstudy compact` removes tombstoned rows but leaves the database file
  bigger than before: each `DELETE` appends to bytdb's log (measured
  17,867 → 18,868 bytes after removing two rows). bytdb v0.19+ has `VACUUM`
  (`Engine.Compact`, reachable via `stdlib.Engine`), which rewrites the log on
  demand; running it at the end of a non-dry-run `compact` would make the
  command shrink the file it just purged. Not a transaction-block statement,
  so it must run outside one. See N-012 for the related decision.

- **N-018** · raised `2026-1012-0131-bytdb-v0212-upgrade-next-list-seed` · value low
  Every `pbstudy backup` now leaves a `study-<stamp>.bytdb.lock` sidecar beside
  the snapshot in `<SyncDir>/backups`, because bytdb v0.18+ `Backup` locks its
  destination and btypedb deliberately never deletes lock sidecars. Harmless
  (the lock is the flock, not the file; `recentBackups` lists only `*.bytdb`;
  the syncer reads only `*.json` record files), but one stray file per backup is
  uploaded to every machine. Either remove the sidecar once `Backup` returns
  (no other process ever opens a fresh snapshot, so btypedb's reason for
  keeping sidecars doesn't apply) or accept the clutter.

- **N-019** · raised `2026-1012-0131-bytdb-v0212-upgrade-next-list-seed` · value low
  The AI draft uses `claude-sonnet-5` (`ai/draft.go:41`), which is valid but
  now the previous generation; `claude-sonnet-5-5` is the current Sonnet at the
  same price. `draft.go` already sends none of the parameters Sonnet 5.5
  rejects (no sampling params, no `thinking` config, no forced `tool_choice`),
  so the swap looks like one constant plus re-checking the request-shape test.
  Best done together with the first live run (N-004), so the run observes
  the model that ships.

## Validate

Items whose remaining work is purely testing: run, look, hear, measure, or
write or repair a test. No product change is planned unless the check finds a
defect.

- **N-001** · raised `2026-0805-1823-phase1-skeleton-reader` · value medium
  ScriptTagger hover popups in a real browser, inside note bodies. Unverified
  since Phase 2: the Chrome instance available then could not reach this
  machine's localhost (`ERR_CONNECTION_REFUSED` while curl got 200). The served
  HTML is right: the `BLB.Tagger` config precedes the script tag, and
  `blb-no-tag` is on scripture containers but not on note bodies. It needs a
  human (or a browser that can reach localhost) to look:
  `PBSTUDY_DATA_DIR=/tmp/pbstudy-dev go run . serve`.

- **N-004** · raised `2026-0805-2012-phase4-sermon-builder-ai` · value medium
  A successful live AI draft has never been observed (no `ANTHROPIC_API_KEY`
  in any session so far). The request shape, the streaming parse, the framing
  and every error path are covered offline, and the real API took the request
  as far as rejecting the key. The happy path against the real model wants
  one run with a real key.

## Roadmap

Wanted, but deferred. Parked, not declined. No session doc has deferred an
item yet; items move here from Open when the user decides they are "later".

## Non-goals

- **N-008** · declined `2026-0805-2048-phase5-sync` — Two machines editing the
  same note while both are offline still loses one side (last write wins).
  Accepted for a single user; the snapshots are the net.
- **N-009** · declined `2026-0805-2048-phase5-sync` — Translation download
  buttons on the settings page. Downloading 66 books from a web handler needs
  its own progress stream and concurrency guard, and `pbstudy download`
  already does the job.
- **N-010** · declined `2026-0805-2012-phase4-sermon-builder-ai` — JavaScript
  reorder for outline sections. The up/down buttons are plain POST forms that
  work with JS disabled and are keyboard-accessible for free.
- **N-011** · declined `2026-0805-2238-compaction` — A compaction button in the
  settings page. Compaction is neither additive nor reversible, and the
  retention window that makes it safe is one only the user can choose; it
  stays a command.
- **N-012** · declined `2026-0805-2238-compaction` — A database-only
  compaction mode. bytdb's file is append-only, so `DELETE` reclaims nothing on
  disk; the saving is in the sync folder. A smaller database is
  `pbstudy backup`. *Premise loosened 2026-10-12:* bytdb v0.19+ has `VACUUM`
  (`Engine.Compact`), which does shrink the file on demand, and btypedb
  auto-compacts the log once it has grown enough. The decision may deserve a
  second look; `compact` does not run `VACUUM` today.

## Closed

Closures before this file existed are recorded in the session docs' bodies.
The entries below are the ones found done while seeding the file.

- **N-002** · raised `2026-0805-1917-phase2-notes-tags-xrefs` · closed
  2026-10-14 — No framework backstop for the escaping rule. Added
  `web/ui/escape_test.go`: it walks every view struct with reflection,
  poisons every exported string it can reach with `"'><script>`, renders,
  and fails on any raw payload. Fields that are constrained upstream (validated
  translations, `NormalizeScope`, handler-built URLs, sync-report counts) are
  listed in `trustedFields`, each with where the constraint lives. A new field
  is poisoned automatically. The first run found four sinks, all fixed:
  `ErrorPage` heading/detail (echo URL path segments; latent because browsers
  percent-encode `<` in paths, but reproduced with a raw request on a scratch
  server), Settings `kv()` (environment paths), `kindLabel`'s fallback (a
  sync-file kind), and record ids in hrefs/actions (sync import accepts any
  non-empty id). Ids now go through `pathSeg()` (`url.PathEscape`) in
  `NoteURL`/`TagURL`/`SermonURL`, section buttons and xref delete; that is a
  no-op on UUIDs.
- **N-013** · raised `2026-0805-1823-phase1-skeleton-reader` · closed
  2026-10-12 — Confirm the exact model ID for the AI draft rather than
  trusting `claude-sonnet-5` from `PLAN.md`. Checked against the current model
  list: `claude-sonnet-5` (`ai/draft.go:41`, `DefaultModel`) is a valid,
  still-served ID. It is no longer the newest Sonnet; see the session report.
- **N-014** · raised `2026-0805-1917-phase2-notes-tags-xrefs` · closed
  `2026-0807-0954-bytdb-v091-workaround-removal` — bytdb's `DISTINCT` +
  qualified `ORDER BY` constraint. Fixed upstream in bytdb v0.9.1; the
  workaround was removed, and `study/queries_test.go` guards it.
- **N-015** · raised `2026-0805-2048-phase5-sync` · closed
  `2026-0805-2238-compaction` — Nothing prunes tombstones. `pbstudy compact`
  removes expired tombstones from the database and the sync folder.
- **N-016** · raised `2026-0805-2012-phase4-sermon-builder-ai` · closed
  `2026-0805-2048-phase5-sync` — `status` can stick at `drafting` if the
  process dies mid-draft; sync must not treat it as meaningful. Import
  normalizes it (`study/sync.go:558`): `drafting` becomes `drafted` or
  `outline`. Locally the stuck value stays harmless. No UI shows sermon status
  and nothing gates on it; the draft lock is in memory (`s.drafts.begin`).
