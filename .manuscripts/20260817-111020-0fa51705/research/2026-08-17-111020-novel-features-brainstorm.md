# Novel-features brainstorm (executed inline; no Task tool in this runtime)

## Customer model

### Commuter Mira
- **Today:** Opens DB Navigator, types her station, scrolls a noisy board, misses a platform change. Sometimes she greps a Python one-liner against fchg XML.
- **Weekly ritual:** Before leaving home, check "is my usual ICE still on platform 7, and is it late?"
- **Frustration:** `plan` lies (it is static). She has to mentally merge fchg. No one command answers "what cancelled this hour at my station?"

### Station-side traveler Jonas
- **Today:** Looks up at the split-flap / bahn.de board. If he is scripting, he calls four XML endpoints and decodes `pp`/`cp`.
- **Weekly ritual:** Arrive at a hub, find the current platform for the next long-distance departures.
- **Frustration:** MCP tools give a merged board but not a dedicated platform-change list. He wants `--json --select train,planned_platform,changed_platform`.

### Agent/script Riley
- **Today:** Wraps `deutsche-bahn-api` or an MCP server. No durable local store. Re-fetches fchg for every follow-up and burns the 60/min quota.
- **Weekly ritual:** Sync one station, then answer "cancellations" / "platforms" from cache.
- **Frustration:** Existing tools are libraries or MCP-only. No Go CLI with SQLite + `--agent`.

### Disruption watcher Anika
- **Today:** Re-downloads full fchg every 30s or scrapes bahn.de (we will not).
- **Weekly ritual:** After a disruption, watch one EVA for new cancels/platform moves.
- **Frustration:** Official docs already say "load fchg once, then rchg" — no CLI implements that as `watch`.

## Candidates (pre-cut)
1. board — overlay plan+fchg (persona Mira/Jonas, source a+b). KEEP.
2. platforms — platform-change list (Jonas, a). KEEP.
3. cancellations — cancelled this hour (Mira, a). KEEP.
4. delays — delay_minutes filter (Mira, a). KEEP.
5. watch — rchg onto cached snapshot (Anika, a+c). KEEP.
6. insight — generated station-to-command map (prior). KILL: hollow, not a traveler ritual.
7. health/coverage/stale — generated store stats (prior). KILL as product thesis: useful plumbing, not the reason to install.
8. journey — A to B (user vision anti-goal). KILL: not in spec; would require scrape.
9. fares/book — KILL: no official self-serve API.
10. raw-xml dump — KILL: wrapper, not transcendence.
11. map-view — KILL: scope creep / TUI.
12. multi-station compare — borderline; cut as weekly-use weak vs board.
13. message-feed (distributor messages) — useful but niche; cut vs cancellations.
14. eva-resolve alias — already `station`; wrapper.

## Survivors and kills

### Survivors
See absorb manifest transcendence table (5 rows, all >= 7/10, all hand-code).

### Killed candidates
| feature | kill reason | closest-surviving-sibling |
|---------|-------------|---------------------------|
| insight/health/coverage/stale as thesis | hollow generated-only; not a weekly traveler ritual | board / watch |
| journey A to B | not in official API; scrape forbidden | none (anti-trigger) |
| fares/book | no official self-serve API | none |
| raw-xml dump | thin wrapper | board |
| map-view | scope creep | platforms |
| multi-station compare | weak weekly use | board |
| message-feed | niche vs cancel/delay | cancellations |

## Reprint verdicts
Prior research.json listed insight/health/coverage/stale. Verdict: **drop** as transcendence / product thesis (hollow). Framework commands may still be generated; they are not why this CLI exists.
