# Absorb Manifest

API: db-timetables
Stamp: 2026-08-17-111020
Spec: official OpenAPI 3.0.1 Timetables 1.0.274

## Tools surveyed
- Official Marketplace Timetables 1.0.274 (4 GET endpoints)
- eric134422/db-fahrplan-mcp (Python MCP, GitHub)
- jorekai/db-timetable-mcp (TypeScript MCP, GitHub)
- mbbrueckner/deutsche-bahn (Python library, GitHub)
- pypi deutsche-bahn-api (Python library; station + plan + changes merge)
- Incumbents DB Navigator / bahn.de (consumer apps; not absorbed as scrape targets)

## Absorbed (match or beat everything that exists)
| # | Feature | Best Source | Our Implementation | Added Value |
|---|---------|-----------|-------------------|-------------|
| 1 | Station search by name / EVA / DS100 | official `/station/{pattern}`; db-fahrplan-mcp `find_station`; db-timetable-mcp `findStations` | db-timetables-pp-cli station --pattern Frankfurt | `--json`/`--agent`, cache the hit, real examples (`FF`, `8000105`, `BLS`) |
| 2 | Planned hourly board | official `/plan/{evaNo}/{date}/{hour}`; deutsche-bahn-api `get_timetable` | db-timetables-pp-cli plan 16 --eva-no 8000105 --date 260817 | typed date/hour, dry-run, local slice upsert |
| 3 | Full-day live changes | official `/fchg/{evaNo}`; MCP `get_live_changes` / `getCurrentTimetable` | db-timetables-pp-cli fchg --eva-no 8000105 | delay/platform/cancel XML to JSON, cache by EVA |
| 4 | Recent 2-minute changes | official `/rchg/{evaNo}`; MCP `get_recent_changes` | db-timetables-pp-cli rchg --eva-no 8000105 | cheap poll after a full snapshot |
| 5 | Merged live departures (plan+fchg) | db-fahrplan-mcp `get_departures`; db-timetable-mcp `getStationBoard`; mbbrueckner `get_timetable_with_changes` | db-timetables-pp-cli board --eva-no 8000105 | same merge, plus `--json --select` and optional local store |
| 6 | XML attribute decode (pt/ct/pp/cp/cs) | db-fahrplan-mcp parser.py; db-timetable-mcp | (behavior in db-timetables-pp-cli board) named JSON fields + delay_minutes | agents do not have to learn IRIS single-letter attrs |
| 7 | Dual Marketplace headers | official securitySchemes ClientID+ClientSecret; all wrappers | (behavior in db-timetables-pp-cli doctor) both headers required | doctor fails closed if either env is missing |
| 8 | Local station/plan/fchg cache | no competitor has this | (behavior in db-timetables-pp-cli sync) --resources station,plan,fchg | follow-ups without burning the 60/min quota |
| 9 | Agent output modes | printing-press framework | (behavior in db-timetables-pp-cli station) --json --agent --select --dry-run | scriptable; not an MCP-only surface |

No stubs. Scrape/HAFAS/bahn.de paths are explicitly out of scope.

## Transcendence (only possible with our approach)
| # | Feature | Command | Score | Buildability | How It Works | Evidence | Long Description |
|---|---------|---------|-------|--------------|--------------|----------|------------------|
| 1 | Live station board | board | 9/10 | hand-code | Calls `/plan` + `/fchg` (or reads synced slices) and overlays ct/cp/cs onto planned stops | Every serious wrapper (db-fahrplan-mcp get_departures, db-timetable-mcp getStationBoard, mbbrueckner get_timetable_with_changes) exists because raw plan has no delays | Use this command for the live board at one station this hour. Do NOT use it for A to B journeys; this API cannot plan trips. |
| 2 | Platform-change list | platforms | 8/10 | hand-code | Filters overlaid stops where changed platform (`cp`) differs from planned (`pp`) | Traveler Jonas + MCP docs call out planned vs changed platform as the #1 concourse question | Use this command for trains that moved platform. Do NOT use it for the full board; use 'board' instead. |
| 3 | Cancellations this hour | cancellations | 8/10 | hand-code | Filters overlaid plan+fchg (or local slices) for cancelled events (`cs` / `clt`) in the current hour | Commuter Mira's ritual; official fchg description lists cancel as a first-class change | Use this command for cancelled stops this hour. Do NOT use it for delays that still run; use 'delays'. |
| 4 | Delay list | delays | 7/10 | hand-code | Computes delay_minutes from planned `pt` vs changed `ct` on the overlaid hour | deutsche-bahn-py prints delay_minutes on the merged timetable; agents need a filter, not the full XML | Use this command for late trains. Do NOT use it for cancellations; use 'cancellations'. |
| 5 | Cheap change watch | watch | 7/10 | hand-code | Requires a cached fchg/plan snapshot in SQLite, then applies `/rchg` (last ~2 min) as a delta | Official rchg docs: load fchg once, then poll rchg to save bandwidth; Anika's ritual | Use this command to apply recent changes onto a cached snapshot. Do NOT use it as the first fetch; use 'board' or 'fchg' first. |

Minimum 5 transcendence rows. Hollow generated-only `insight` / `health` / `coverage` / `stale` are **not** the product thesis and are not listed here.

## Hand-code commitment
5 novel features, all `hand-code` (~50-150 LoC each plus `root.go` wiring). Generator will still emit `station` / `plan` / `fchg` / `rchg` plus framework `sync`/`search`/`doctor`.
