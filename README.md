# db-timetables

**Official station boards: delay, platform, cancel — not a journey planner.**

A Go CLI for Deutsche Bahn's Marketplace Timetables API. Resolve a station, pull the hourly plan, overlay live changes, and keep a local slice so an agent can ask follow-ups without burning quota. This is not A to B, fares, or booking.

Learn more at [Deutsche Bahn Timetables](https://developers.deutschebahn.com/db-api-marketplace/apis/product/timetables).

## What you get

Four Marketplace endpoints, plus traveler-facing overlays and a local store so an agent can cache a station slice and ask follow-ups offline:

| Command | What it hits | Use it for |
| --- | --- | --- |
| `station` | `GET /station/{pattern}` | Resolve a name or EVA number (`FF`, `Frankfurt`, `8000105`) |
| `plan` | `GET /plan/{evaNo}/{date}/{hour}` | The planned board for one hour (date is `YYMMDD`) |
| `fchg` | `GET /fchg/{evaNo}` | All known changes from now on: delay, platform, cancel, extra stops |
| `rchg` | `GET /rchg/{evaNo}` | Changes that became known in the last ~2 minutes |
| `board` | plan + fchg overlay | Live board this hour |
| `platforms` | overlay filter | Only trains whose platform moved |
| `cancellations` | overlay filter | Cancels this hour |
| `delays` | overlay filter | Late-but-running trains |
| `watch` | cached plan+fchg + rchg | Cheap poll after a snapshot |

`plan` is static. Live truth is `fchg` (or `rchg` if you already have a full snapshot and are polling). Combine them: look the station up, pull the hour's plan, then overlay full changes.

## Authentication

Marketplace Anwendung credentials. Every request must send both headers: DB-Client-Id from DB_TIMETABLES_CLIENT_ID (or DB_CLIENT_ID) and DB-Api-Key from DB_TIMETABLES_API_KEY (or DB_API_KEY). Auth is configured only when both values are present. Create an Anwendung, subscribe it to Timetables, and export the two env vars. Do not commit keys.

## Quick Start

```bash
# Health check that works without hitting the API.
db-timetables-pp-cli doctor --dry-run

# Resolve a name or EVA without spending a live call.
db-timetables-pp-cli station --pattern Frankfurt --dry-run

# Show the live-board request (plan + fchg overlay) for Frankfurt Hbf.
db-timetables-pp-cli board --eva-no 8000105 --dry-run

# Platform-change filter for the same station.
db-timetables-pp-cli platforms --eva-no 8000105 --dry-run

# Cancellations this hour.
db-timetables-pp-cli cancellations --eva-no 8000105 --dry-run

```

## Build

Go 1.26.6+. Module path is the generated `db-timetables-pp-cli`, so install from a clone.

```bash
git clone https://github.com/kingdomseed/db-timetables.git
cd db-timetables
make build          # bin/db-timetables-pp-cli
make test
./bin/db-timetables-pp-cli --help
```

`make build-mcp` builds the MCP server binary if you want that surface.

## Unique Features

These capabilities aren't available in any other tool for this API.

### Live station board
- **`board`** — See the live board for one station this hour: planned trains with delays, platform moves, and cancels overlaid.

  _Reach for this instead of calling plan and fchg separately when a traveler or agent wants the truth at a station._

  ```bash
  db-timetables-pp-cli board --eva-no 8000105 --agent --select train,planned_time,delay_minutes,platform,cancelled
  ```
- **`platforms`** — List only the trains at a station whose platform changed.

  _Use this when the question is which platforms moved, not the full board._

  ```bash
  db-timetables-pp-cli platforms --eva-no 8000105 --json
  ```
- **`cancellations`** — List cancellations at a station for the current hour.

  _Use this when the question is what was cancelled, not every delay._

  ```bash
  db-timetables-pp-cli cancellations --eva-no 8000105 --json
  ```
- **`delays`** — List late trains at a station with delay minutes.

  _Use this for late-but-running trains; cancellations are a different command._

  ```bash
  db-timetables-pp-cli delays --eva-no 8000105 --json
  ```

### Local state that compounds
- **`watch`** — Apply the last two minutes of rchg onto a cached plan+fchg snapshot.

  _Use this to poll cheaply after board/fchg; do not use it as the first fetch._

  ```bash
  db-timetables-pp-cli watch --eva-no 8000105 --json
  ```

## Cookbook

Frankfurt (Main) Hbf is EVA `8000105`. Date is `YYMMDD`, hour is `HH`.

```bash
# Find the station
db-timetables-pp-cli station --pattern "Frankfurt" --json

# Planned departures/arrivals for 16:00 on 17 Aug 2026
db-timetables-pp-cli plan 16 --eva-no 8000105 --date 260817 --json

# Live changes: delays, platform moves, cancellations
db-timetables-pp-cli fchg --eva-no 8000105 --json

# Same, but only what changed in the last two minutes
db-timetables-pp-cli rchg --eva-no 8000105 --json

# Overlay: live board this hour
db-timetables-pp-cli board --eva-no 8000105 --json
```

`--json` for scripts. `--agent` for JSON + compact + no prompts. `--dry-run` prints the request and does not call the API. `--select` keeps only the fields you name.

## Recipes

### Resolve Frankfurt Hbf

```bash
db-timetables-pp-cli station --pattern Frankfurt --agent --select stations.eva,stations.name,stations.ds100
```

Turn a name into EVA 8000105 before any board call.

### Live board this hour

```bash
db-timetables-pp-cli board --eva-no 8000105 --agent --select train,planned_time,delay_minutes,platform,cancelled
```

Overlay plan+fchg and keep only the fields an agent needs.

### Platform changes

```bash
db-timetables-pp-cli platforms --eva-no 8000105 --json
```

Only trains whose platform moved.

### Cancellations this hour

```bash
db-timetables-pp-cli cancellations --eva-no 8000105 --json
```

Cancelled stops at the station for the current hour.

### Cheap poll after a snapshot

```bash
db-timetables-pp-cli watch --eva-no 8000105 --json
```

Apply rchg onto the cached plan+fchg slice.

## Usage

Run `db-timetables-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

This CLI separates local files into four path kinds:

| Kind | Contents |
|------|----------|
| `config` | User-editable settings such as `config.toml` and saved profiles |
| `data` | Durable local data: `credentials.toml`, `data.db`, cookies, and other auth sidecars |
| `state` | Runtime state such as persisted queries, jobs, and `teach.log` |
| `cache` | Regenerable HTTP/cache files |

Each kind resolves independently. The ladder is:

1. Per-kind env var: `DB_TIMETABLES_CONFIG_DIR`, `DB_TIMETABLES_DATA_DIR`, `DB_TIMETABLES_STATE_DIR`, or `DB_TIMETABLES_CACHE_DIR`
2. `--home <dir>` for this invocation
3. `DB_TIMETABLES_HOME` for a flat relocated root
4. XDG env vars: `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`
5. Platform defaults matching existing installs

For containers and agent sandboxes, prefer a single relocated root:

```bash
export DB_TIMETABLES_HOME=/srv/db-timetables
db-timetables-pp-cli doctor
```

## Commands

### station

- **`db-timetables-pp-cli station`** — Resolve a station by pattern (name, DS100, or EVA).

### plan

- **`db-timetables-pp-cli plan <hour>`** — Planned timetable for `evaNo` on `date` (YYMMDD) and `hour` (HH). Static. Does not include delays.

### fchg

- **`db-timetables-pp-cli fchg`** — All known changes for the station from now on (delay, platform, cancel, extra stops).

### rchg

- **`db-timetables-pp-cli rchg`** — Changes that became known in the last ~2 minutes. Subset of fchg; use after an initial fchg snapshot.

### board / platforms / cancellations / delays / watch

See Unique Features. These overlay plan+fchg locally; they are not extra Marketplace endpoints.

### Self-learning loop

- **`db-timetables-pp-cli recall <query>`** — Look up cached resources for a query
- **`db-timetables-pp-cli teach`** — Record a query -> resource mapping
- **`db-timetables-pp-cli learnings list`** — Inspect taught rows

Pass `--no-learn` or set `DB_TIMETABLES_NO_LEARN=true` to disable the loop.

## Output Formats

```bash
db-timetables-pp-cli fchg --eva-no 8000105
db-timetables-pp-cli fchg --eva-no 8000105 --json
db-timetables-pp-cli fchg --eva-no 8000105 --json --select eva,m,s
db-timetables-pp-cli fchg --eva-no 8000105 --dry-run
db-timetables-pp-cli fchg --eva-no 8000105 --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** — never prompts, every input is a flag
- **Pipeable** — `--json` output to stdout, errors to stderr
- **Filterable** — `--select <field>[,<field>...]` returns only fields you need
- **Previewable** — `--dry-run` shows the request without sending
- **Read-only by default** — this CLI does not create, update, delete, or book
- **Offline-friendly** — cached plan/fchg slices can answer follow-ups
- **Agent-safe by default** — no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
db-timetables-pp-cli doctor
```

Verifies configuration, credentials, and connectivity. Both Marketplace headers must be present.

## Configuration

Run `db-timetables-pp-cli doctor` to see the resolved config, data, state, and cache directories.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `DB_TIMETABLES_CLIENT_ID` | per_call | Yes | Marketplace client id (`DB-Client-Id`). Alias: `DB_CLIENT_ID`. |
| `DB_TIMETABLES_API_KEY` | per_call | Yes | Marketplace API key (`DB-Api-Key`). Alias: `DB_API_KEY`. |

## Troubleshooting

**Authentication errors (exit code 4)**
- Run `db-timetables-pp-cli doctor` to check credentials
- Export **both** `DB_TIMETABLES_CLIENT_ID` and `DB_TIMETABLES_API_KEY`. One header is not enough.

**Not found errors (exit code 3)**
- Check the EVA number (`station --pattern FF`)
- `plan` date is `YYMMDD` and hour is `HH`

### API-specific
- **401/403 from Marketplace** — Export both DB_TIMETABLES_CLIENT_ID and DB_TIMETABLES_API_KEY. One header is not enough.
- **empty station results for a German name with umlauts** — Retry with EVA or DS100 (FF, 8000105). The official station pattern does not handle umlauts well.
- **plan looks on time but the train is late** — plan is static. Use board (plan+fchg overlay) or fchg.
- **429 / rate limit** — Marketplace is about 60 calls/min. Cache plan (static) and poll rchg after one fchg via watch.

## What this is not

The long-term travel helper wants booked-train status from mail, A→B including local transit, and prices. This repo is the official boards slice only. That is the part Marketplace actually gives you for free.

## License

Apache-2.0. Spec and board data come from Deutsche Bahn's Timetables API (CC BY on the data side; you are responsible for your own Marketplace terms). Code and data licenses are separate; see `NOTICE`.

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
