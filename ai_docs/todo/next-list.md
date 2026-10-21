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

**Next ID:** N-031

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
  *Premise corrected 2026-10-14 (`2026-1014-0741-work-next-list-n019-n005-n001`):*
  that cleanup cannot work for a live tag. The redundant file's id still
  exists on the machine that minted it (`ApplyTag` keeps both local ids, by
  design), and export rewrites any local row whose file is missing
  (`syncer.go`, `doExport`). So deleting it here just makes the other
  machine write it back on its next pass. The fix that would converge is
  **id convergence** on a name merge: both machines adopt one deterministic
  id (say, the lexicographically smaller UUID), re-point `note_tags`, and
  tombstone the losing id so its file is cleared by compaction. That changes
  tag identity across machines, which `ApplyTag`'s comment currently argues
  is unnecessary. **Question for the user:** worth doing for one small file
  per machine per shared tag, or move this to Non-goals?

- **N-007** · raised `2026-0805-2238-compaction` · value low
  Compaction is per machine and has to be run on each. A machine that has not
  compacted re-exports its tombstone files; the others import and remove them
  again next pass. Bounded churn, converges once both compact, and what comes
  back is a tombstone, invisible to every read in `study/`.

- **N-020** · raised `2026-1014-0741-work-next-list-n019-n005-n001` · value low
  Sonnet 5.5's safety classifiers can decline a request (`stop_reason:
  "refusal"`). `ai/draft.go` already reports that as a user-facing error
  rather than an empty draft. The API also offers server-side fallbacks
  (beta `server-side-fallback-2026-07-01` + `fallbacks: "default"`), which
  re-serve a declined request on another model inside the same call. Not
  added with N-019 because it adds a beta header to a hand-rolled client and
  changes how a refusal is billed. A sermon draft is unlikely to trip a
  classifier, so this matters only if refusals are ever seen (N-004's live
  run would be the first chance).

- **N-023** · raised `2026-1019-2255-preferred-translation-fork` · value low
  Resume reading: bare `/read` always lands on John 1
  (`handleReadDefault`). Remember the last chapter read, e.g. in the bible
  db's `prefs` table added for the preferred translation.

- **N-024** · raised `2026-1019-2255-preferred-translation-fork` · value medium
  Start on login. The server only runs while a terminal keeps it alive,
  worked around today by starting it by hand. A launchd agent (plist plus an
  install note or command) would keep it at `localhost:8000`.

- **N-025** · raised `2026-1019-2255-preferred-translation-fork` · value low
  Verse highlighting in a few colours from the reader, separate from notes.
  Needs a new synced study table (id / updated_at / deleted_at) and sync
  export.

- **N-026** · raised `2026-1019-2255-preferred-translation-fork` · value low
  Side-by-side chapter reading across translations. The verse hub already
  shows every translation for one verse; nothing does it for a chapter.

- **N-027** · raised `2026-1019-2255-preferred-translation-fork` · value low
  Reading plan / reading log: mark chapters read, track progress through a
  plan. Largest of this batch; synced like other study data.

## Validate

Items whose remaining work is purely testing: run, look, hear, measure, or
write or repair a test. No product change is planned unless the check finds a
defect.

- **N-004** · raised `2026-0805-2012-phase4-sermon-builder-ai` · value medium
  A successful live AI draft has never been observed (no `ANTHROPIC_API_KEY`
  in any session so far). The request shape, the streaming parse, the framing
  and every error path are covered offline, and the real API took the request
  as far as rejecting the key. The happy path against the real model wants
  one run with a real key. Since N-019 the model is `claude-sonnet-5-5`, so
  that run observes the model that ships: confirm a thinking status appears
  before the first prose and the draft saves.

## Roadmap

Wanted, but deferred. Parked, not declined. Items move here from Open when
the user decides they are "later".

- **N-029** · raised `2026-1021-1802-loopback-ai-options` · value low
  AI study assistant: a "Study with AI" button on the verse hub that drafts a
  note (context, key words as Strong's shortcodes, suggested cross-references
  validated with `bible.ParseRef`, reflection questions) from the verse in
  every cached translation plus the user's own notes and xrefs. Opens in the
  note editor, labelled as AI-written; nothing saves until the user does.
  Reuse `ai/` streaming and `web/drafts.go`. Three interchangeable backends,
  chosen in settings: the user's Claude subscription via `claude -p
  --output-format json` (personal fork only: Anthropic does not allow
  offering claude.ai login in products for others), `ANTHROPIC_API_KEY`, or a
  free local model through Ollama. This machine is an M2 with 8 GB, so local
  means a 3–4B model: fine for explaining, weak on facts. A chat panel is a
  possible second step.

- **N-030** · raised `2026-1021-1802-loopback-ai-options` · value low
  Public-domain study resources cached like the translations: Treasury of
  Scripture Knowledge cross-references, Matthew Henry's commentary, Strong's
  definitions. Reliable, free and offline; pairs with N-029.

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
  second look. Since 2026-10-14 (N-017), `compact` ends every real pass by
  compacting the file, but it still requires a sync directory, so there is
  no database-only mode.
- **N-028** · declined `2026-1019-2255-preferred-translation-fork` — NKJV.
  It is under copyright, so it cannot be cached like KJV/WEB/ASV; an NKJV
  option for the Blue Letter Bible link-out was offered and the user said no.

## Closed

Closures before this file existed are recorded in the session docs' bodies.
The entries below are the ones found done while seeding the file.

- **N-022** · raised `2026-1019-2255-preferred-translation-fork` · closed 2026-10-21, `2026-1021-1802-loopback-ai-options` —
  `pbstudy serve` listened on every interface with no authentication, so
  anyone on the same Wi-Fi could read and edit the study database. Now
  `PBSTUDY_HOST` defaults to `127.0.0.1` (an empty host also reads as
  loopback); `0.0.0.0` opts in for phone access and is announced at startup
  and on the settings page, which now shows the full listen address.
  Verified: refused via the LAN IP by default, reachable with the opt-in.
- **N-021** · raised `2026-1014-0741-work-next-list-n019-n005-n001` · closed 2026-10-14, `2026-1014-0741-work-next-list-n019-n005-n001` —
  Found by N-001's check: the ScriptTagger config never took effect. The
  tagger replaces `BLB.Tagger` with its defaults on load, so the inline
  object emitted *before* the script tag was discarded: popups quoted NKJV,
  `HyperLinks` stayed `"none"` (it takes `"all"/"none"/"hover"`, not a
  bool), and `blb-no-tag` was never excluded, so scripture containers were
  tagged too (a control page in Chrome tagged a `blb-no-tag` div). Now a
  deferred `/js/blb-tagger.js` runs right after the tagger, reads
  translation and exclusion class from data attributes, and sets the
  documented properties. Checked in Chrome on the real app: `Translation`
  KJV, `HyperLinks` "all", `NoSearchClasses` includes `blb-no-tag`, popup
  reads "Romans 8:28 (KJV)". `TestTaggerConfigFollowsTagger` pins the order.
- **N-001** · raised `2026-0805-1823-phase1-skeleton-reader` · closed
  2026-10-14, `2026-1014-0741-work-next-list-n019-n005-n001` — ScriptTagger hover popups in a real browser, inside
  note bodies. Chrome (via the extension) now reaches localhost. On a
  scratch server, a note body's "John 3:16", "Romans 8:28" and "Eph 2:8-9"
  were all tagged, and hovering one opened the BLB popup with the verse
  text. The app's own reference chip and the scripture card were not tagged.
  The check also found that the tagger config was being ignored, raised and
  fixed as N-021. After the fix the popup quotes KJV.
- **N-019** · raised `2026-1012-0131-bytdb-v0212-upgrade-next-list-seed` ·
  closed 2026-10-14, `2026-1014-0741-work-next-list-n019-n005-n001` — `DefaultModel` is now
  `claude-sonnet-5-5` (same price as `claude-sonnet-5`). Re-checked against
  Sonnet 5.5's rules: the request sends no sampling params, no `thinking`
  and no `tool_choice` (each would 400), and adaptive thinking still streams
  thinking blocks, so the status event is unchanged. `TestDraftRequestShape`
  now also bans `tool_choice`; `ai` and `web` tests green. Not observed live;
  that is N-004.
- **N-017** · raised `2026-1012-0131-bytdb-v0212-upgrade-next-list-seed` ·
  closed 2026-10-14, `2026-1014-0734-n002-n006-n017-n018-escape-backstop-sigterm-vacuum-sidecar` — `pbstudy compact` grew the file it purged. New
  `Store.CompactStudy` calls `Engine.Compact` (the native form of `VACUUM`,
  no transaction), and every non-dry-run `compact` ends with it and prints
  `Study database  before → after bytes`. Checked on a scratch data dir:
  deleting 4 of 6 notes left `study.bytdb` at 16,595 bytes after the purge,
  and 5,346 after compaction. `TestCompactStudyShrinksAfterDeletes` pins the
  shrink and that the surviving row reads back through the same handle.
  README and usage text updated.
- **N-018** · raised `2026-1012-0131-bytdb-v0212-upgrade-next-list-seed` ·
  closed 2026-10-14, `2026-1014-0734-n002-n006-n017-n018-escape-backstop-sigterm-vacuum-sidecar` — Backup left a `.lock` sidecar in the synced backups
  folder. `BackupStudy` now removes it once `Backup` returns
  (`removeLockSidecar`, symlink-resolved the way btypedb names it). The
  sidecar race btypedb guards against needs a second opener of the same path,
  and nothing opens a freshly minted snapshot. `store/store_test.go` checks
  that only the `*.bytdb` file remains, directly and through a symlinked
  folder; both tests fail with the removal disabled.
- **N-006** · raised `2026-0805-2048-phase5-sync` · closed 2026-10-14, `2026-1014-0734-n002-n006-n017-n018-escape-backstop-sigterm-vacuum-sidecar` —
  No signal handler without sync. `installShutdownFlush` now takes a nil
  syncer (skips the flush, still closes the store) and `cmdServe` installs it
  unconditionally. Checked on a scratch server with no `PBSTUDY_SYNC_DIR`:
  `SIGTERM` printed "Stopping (terminated) ..." and the process exited through
  `st.Close()`.
- **N-002** · raised `2026-0805-1917-phase2-notes-tags-xrefs` · closed
  2026-10-14, `2026-1014-0734-n002-n006-n017-n018-escape-backstop-sigterm-vacuum-sidecar` — No framework backstop for the escaping rule. Added
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
