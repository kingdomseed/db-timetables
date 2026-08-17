# Deutsche Bahn Timetables CLI Brief

## API Identity
- Domain: travel / rail operations (station boards, not journey planning)
- Product: Deutsche Bahn API Marketplace **Timetables 1.0.274**
- Official product: https://developers.deutschebahn.com/db-api-marketplace/apis/product/timetables
- Server: `https://apis.deutschebahn.com/db-api-marketplace/apis/timetables/v1`
- Spec source: official OpenAPI 3.0.1 extracted from the Marketplace product (`/workspace/printing-press/timetables-openapi.json`)
- Users:
  - **Commuter Mira** — leaves for work from a named station (Frankfurt Hbf / FF / 8000105). Before she walks to the platform she needs: is my train delayed, cancelled, or moved?
  - **Station-side traveler Jonas** — already on the concourse. He needs the live board for *this hour* and a list of platform changes, not a journey planner.
  - **Agent/script Riley** — a local helper that syncs one station's plan+fchg slice and answers follow-ups offline (`--json` / `--agent`).
  - **Disruption watcher Anika** — polls `rchg` every ~30-90s after an initial `fchg` snapshot so she only pays for the last two minutes of changes.
- Data profile: four XML GET endpoints. Highest-gravity entities are **station** (EVA / DS100 / name), **hourly plan slice**, **full-day changes (fchg)**, **recent changes (rchg)**. Responses are IRIS XML with single-letter attributes (`pt`/`ct` time, `pp`/`cp` platform, `cs` cancel). This is official station-board data (delay / platform / cancel). It is **not** A→B, fares, or booking.

## User Vision
Jason Holt (kingdomseed) wants a shippable local print of the official Timetables API: a working CLI under a managed run dir, promoted to `$HOME/printing-press/library/db-timetables/`, with dual-header Marketplace auth (`DB-Client-Id` + `DB-Api-Key`) and Phase 5 full live dogfood. The CLI must say honestly what it is and is not. Do not scrape bahn.de. Do not pretend this is a journey planner.

## Reachability Risk
- Low for the official Marketplace host when both headers are present.
- High if only one header is sent (the generator's default single-scheme path). Community wrappers and Stack Overflow (`DB-Client-Id` + `DB-Api-Key` on every GET) confirm both are required.
- Probe-safe endpoint: `GET /station/{pattern}` with pattern `FF` or `8000105` (documented default EVA `8000105`, DS100 default `BLS`).
- Rate limit (Marketplace / community docs): about 60 calls/minute. `plan` is static and cacheable; `fchg`/`rchg` refresh ~30s.
- License: timetable data CC BY 4.0 (name Deutsche Bahn). Code license is separate (Apache-2.0 for the printed CLI).
- No official self-serve API for A→B, fares, or booking. That is a product boundary, not a gap to scrape around.

## Top Workflows
1. **Resolve a station** — name / EVA / DS100 (`Frankfurt`, `8000105`, `FF`) → EVA number.
2. **Hourly planned board** — `GET /plan/{evaNo}/{YYMMDD}/{HH}` for the trains that *should* arrive/depart this hour.
3. **Live overlay** — fetch `fchg` (all known changes from now on) and merge onto the plan: delay minutes, changed platform, cancellation, extra stops.
4. **Cheap poll** — after a full `fchg` snapshot, poll `rchg` (changes known in the last ~2 minutes) instead of re-downloading the day.
5. **Ask follow-ups offline** — cache station + plan + fchg slices in SQLite, then list platform changes / cancellations / delays without another Marketplace call.

## Table Stakes
- Dual-header Marketplace auth on every call (`DB-Client-Id` + `DB-Api-Key`).
- Station search by name, EVA, DS100/RL100, wildcard.
- Planned hourly slice (`plan`) with real `YYMMDD` + `HH` values, never `example-value`.
- Full changes (`fchg`) and recent changes (`rchg`).
- XML → structured JSON for agents (`--json`, `--select`, `--agent`).
- Merge plan + fchg into a usable live board (every serious wrapper does this; raw `plan` has no delays).
- Honest "this is not A→B / fares / booking" framing.

## Data Layer
- Primary entities: `station`, `plan` (hourly slice keyed by eva+date+hour), `fchg` (eva, operating day), `rchg` (eva, poll window).
- Sync cursor: last successful fetch per (resource, eva, slice). `plan` is static (cache hours). `fchg` stale after ~2-5 minutes. `rchg` is a delta, not a full replace.
- FTS/search: station name / DS100 / EVA. Stop-level search on cached plan+fchg (train id, line, destination).
- Local store value: an agent can `sync` one station and then ask "which platforms moved?" / "what cancelled this hour?" without burning quota.

## Competitors
- **DB Navigator / bahn.de** — the incumbent consumer apps. Journey planning, tickets, full UI. Not a CLI; not this API.
- **Third-party scrapers / HAFAS unofficial clients** — we will not use or absorb scrape paths.
- **eric134422/db-fahrplan-mcp** (Python MCP) — `find_station`, merged `get_departures` (plan+fchg), `get_live_changes`, `get_recent_changes`. Parses XML to compact JSON. No local store.
- **jorekai/db-timetable-mcp** (TypeScript MCP) — `getStationBoard` (plan+fchg), planned/current/recent, `findStations`. Documents 60 req/min. No SQLite.
- **mbbrueckner/deutsche-bahn** / `deutsche-bahn-py` — Python client with `get_timetable_with_changes` (the method they call "most useful").
- **pypi `deutsche-bahn-api`** — `get_timetable` + `get_timetable_changes` merge. Library, not a CLI.

## Product Thesis
- Name: **db-timetables** (binary `db-timetables-pp-cli`)
- Why it should exist: the official Marketplace surface is four XML endpoints that are useless to an agent until you (1) send both headers, (2) resolve EVA, (3) overlay plan+fchg, and (4) keep a local slice so follow-ups are free. No existing tool is a Go CLI with `--json`/`--agent`, a SQLite cache, and traveler-shaped commands (`board`, `platforms`, `cancellations`). Hollow generated `insight`/`health`/`coverage`/`stale` are not the product.

## Build Priorities
1. Dual-header auth that actually works (`DB_TIMETABLES_CLIENT_ID` + `DB_TIMETABLES_API_KEY`, aliases `DB_CLIENT_ID` / `DB_API_KEY`). Treat auth as configured only when both are present.
2. Generated endpoint commands: `station`, `plan`, `fchg`, `rchg` with real examples (`8000105`, `FF`/`BLS`, hour `12`, date `YYMMDD`).
3. Hand-coded traveler commands: live `board` (plan+fchg overlay), `platforms`, `cancellations`, `delays`, `watch` (rchg onto a cached snapshot).
4. Local store of station/plan/fchg slices + `sync` so overlay commands can run `--data-source local`.
5. Phase 5 full live dogfood against Frankfurt Hbf (`8000105`) with TestsFailed == 0.

## Auth
- Wire format: two raw header apiKeys, AND, every request.
  - `DB-Client-Id` ← `DB_TIMETABLES_CLIENT_ID` (also `DB_CLIENT_ID`)
  - `DB-Api-Key` ← `DB_TIMETABLES_API_KEY` (also `DB_API_KEY`)
- Not Bearer. Not a single token. Do not invent `TIMETABLES_CLIENT_ID` as a required extra env.
- Marketplace: create an Anwendung, subscribe it to **Timetables**, copy Client ID + API Key.

## Source Priority
- Single source: official Timetables API. No combo CLI.

## Reachability Gate
- PASS: GET /station/FF against the official Marketplace host returned HTTP 200 with a `<stations>` XML body when both Marketplace headers were sent.
- Probe-safe endpoint used: GET /station/{pattern}
