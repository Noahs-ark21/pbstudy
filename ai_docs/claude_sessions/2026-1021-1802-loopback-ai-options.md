# pbstudy — loopback by default (N-022), AI assistant options parked

Session id: `26334965-cf08-468f-8861-0448c59e2996`

Continues the session in `2026-1019-2255-preferred-translation-fork` (same
session id). That doc's commit was pushed to the fork at the start.

---

## N-022: listen on loopback by default (`d8241a0`)

`pbstudy serve` bound `":8000"` (every interface) and the app has no login,
so anyone on the same network could read and edit the study database. The
running server confirmed it: `lsof` showed `TCP *:8000`. The upstream comment
on `cfg.Address` called this deliberate, for phone access, so that stays
possible as an opt-in rather than being removed.

- `cfg/cfg.go`: `PBSTUDY_HOST`, default `127.0.0.1`. It must be an IP or
  `localhost`; a hostname is rejected at startup rather than bound to
  whatever it resolves to. `Address()` uses `net.JoinHostPort`, so `::`
  works. `NetworkReachable()` reports a non-loopback bind. An empty `Host`
  reads as loopback (`listenHost`), so a `Config` built without `Load` (e.g.
  `web/draft_stream_test.go`) cannot fall open by omission.
- `main.go`: a startup warning when network-reachable; the `--help` text
  documents the variable.
- Settings page: "Listen address" replaces "Listen port"; an "Open to your
  network" notice shows when reachable.
- README: the variable, plus a paragraph on when to opt in.
- `TestListenAddress`: default, empty, localhost, `::1`, `0.0.0.0`, `::`, a
  LAN IP, and two rejects.

Verified against real binds. Default: `lsof` shows `127.0.0.1:8000`;
localhost 200, the LAN IP (`192.168.1.173`) refused. A scratch data dir with
`PBSTUDY_HOST=0.0.0.0` on port 8011: reachable via the LAN IP, both warnings
shown. `PBSTUDY_HOST=my-mac.local` refused at startup. The scratch dir was
removed.

## AI assistant for notes — discussed, parked (N-029, N-030)

Three backends were considered for a "Study with AI" note draft on the
verse hub, which would reuse `ai/` streaming and `web/drafts.go`:

- **API key**: what the sermon drafting already uses (N-004 still unrun).
- **The user's Claude subscription** via `claude -p --output-format json`.
  Checked against the docs: non-bare `-p` uses the subscription login;
  `--bare` does not (it reads only `ANTHROPIC_API_KEY`/`apiKeyHelper`), so
  containment would have to come from other flags (no tools, empty working
  dir). The Agent SDK overview says third-party developers may not offer
  claude.ai login in their products without approval, so this is for the
  personal fork only, never proposed upstream.
- **A free local model through Ollama.** This machine is an M2 with 8 GB of
  RAM and 16 GB of free disk, so that means a 3–4B model: fine for
  explaining and summarizing, weak on facts. Embedding llama.cpp in the
  binary was advised against, since the weights are a separate multi-GB
  file either way.

Also raised: public-domain resources (TSK, Matthew Henry, Strong's) as a
reliable, offline, non-AI source. The user put both on the back burner;
they went straight to Roadmap (`84e75e0`).

## Housekeeping

- The `sw` skill points to `~/.claude/commands/sess-wrap.md`, which does
  not exist. Both wraps this session followed `sess-save` instead.

## Next

Closed: N-022. Declined: None. Raised: N-029, N-030. Deferred: N-029, N-030.
Promoted: None. Moved: None. Updated: None.
Full list: `ai_docs/todo/next-list.md`.
