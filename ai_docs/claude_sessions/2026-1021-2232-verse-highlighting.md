# pbstudy — verse highlighting (N-025)

Session id: `26334965-cf08-468f-8861-0448c59e2996`

Continues the session in `2026-1021-1802-loopback-ai-options` (same session
id). That doc's commit was pushed to the fork at the start.

---

## N-025: verse highlighting (`ee63215`)

The user picked N-025 from the next list. A verse can be highlighted yellow,
green, blue or pink, separately from notes.

- **Table**: `highlights` in `study.bytdb` (`book_num, chapter, verse, color`,
  plus `created_at / updated_at / deleted_at`), indexed by chapter.
- **Derived id** `hl-<book>-<chapter>-<verse>`. The same verse highlighted
  on two machines is the same row, so sync converges to the newer colour
  instead of producing two rows (the duplicate problem N-005 has with tags).
  `ApplyHighlight` rejects a record whose id isn't its verse's own.
- **`study/highlights.go`**: `SetHighlight` (UPDATE, else INSERT; an update
  also clears `deleted_at`, so it revives a tombstone), `ClearHighlight`
  (tombstones the row; a no-op on an unhighlighted verse), `HighlightOf`,
  `ChapterHighlights`. Whole-chapter refs and unknown colours are rejected.
- **Sync**: `HighlightRecord` / `ExportHighlights` / `ApplyHighlight`, a
  `highlights` sync folder, report label and compaction, wired in like the
  other four entities.
- **Web**: `POST /highlights` (`at`, `color`, `return`; `color=none`
  clears; 400 on a bad colour or a chapter ref; 303 back). The verse hub has
  a one-form swatch bar (each swatch is a submit button with
  `aria-pressed`, plus Clear when highlighted) and tints its translations
  card. The reader adds `hl-<colour>` to highlighted verses.
- **Escaping**: `TestEveryViewEscapes` caught the colour reaching a class
  attribute unchecked. `highlightClass` now allow-lists it in the view,
  for both the reader and the hub.
- **Tests**: lifecycle (stays one row through set, recolour, clear, set),
  rejects, chapter query, export/apply round trip with a tombstone,
  foreign-id rejection, and `TestHighlightsConvergeByVerse` in the syncer
  (two hosts, one file, newer colour wins, a clear propagates).

Verified end to end on a scratch server (port 8012, scratch data dir with
a copy of `bible.bytdb`): set, recolour, clear, both 400s, the reader
classes and hub swatches, and headless-Chrome screenshots of both pages.
The scratch dir was removed. The real server on `127.0.0.1:8000` was
restarted on the new build.

Not built: highlighting straight from the reader (it takes a trip to the
verse hub now), and highlighting a verse range.

## Housekeeping

- The `sw` skill still points to the missing
  `~/.claude/commands/sess-wrap.md`; this wrap followed `sess-save`.

## Next

Closed: N-025. Declined: None. Raised: None. Deferred: None.
Promoted: None. Moved: None. Updated: None.
Full list: `ai_docs/todo/next-list.md`.
