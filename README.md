# db-timetables

A CLI for Deutsche Bahn's official **Timetables** API (DB API Marketplace).

It answers station-board questions: which trains are planned at a station this hour, what changed (delay, platform, cancel), and what changed in the last two minutes. Built for agents and scripts. Printed with [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press).

This is not a journey planner. It will not search A→B, quote fares, or book tickets. There is no official self-serve API for those. It also does not scrape bahn.de.

## What you get

Four Marketplace endpoints, plus a local store so an agent can sync a station and ask follow-ups offline:

| Command | What it hits | Use it for |
| --- | --- | --- |
| `station` | `GET /station/{pattern}` | Resolve a name or EVA number (`FF`, `Frankfurt`, `8000105`) |
| `plan` | `GET /plan/{evaNo}/{date}/{hour}` | The planned board for one hour (date is `YYMMDD`) |
| `fchg` | `GET /fchg/{evaNo}` | All known changes from now on: delay, platform, cancel, extra stops |
| `rchg` | `GET /rchg/{evaNo}` | Changes that became known in the last ~2 minutes |

`plan` is static. Live truth is `fchg` (or `rchg` if you already have a full snapshot and are polling). Combine them: look the station up, pull the hour's plan, then overlay full changes.

There are also generated local commands (`insight`, `health`, `coverage`, `stale`, `sync`, `doctor`, `auth`). They operate on the SQLite cache, not on extra DB APIs.

## Auth

You need your own Marketplace Anwendung. Subscribe it to **Timetables**. The CLI sends both headers on every call:

- `DB-Client-Id` ← `DB_TIMETABLES_CLIENT_ID` (or `DB_CLIENT_ID`)
- `DB-Api-Key` ← `DB_TIMETABLES_API_KEY` (or `DB_API_KEY`)

```bash
export DB_TIMETABLES_CLIENT_ID=...
export DB_TIMETABLES_API_KEY=...
db-timetables-pp-cli doctor
```

Or `auth set-token` / `auth set-api-key` to store them locally. Do not commit keys. The API is free for personal use; each user brings their own Anwendung.

## Build

Go 1.26.5+. Module path is still the generated `db-timetables-pp-cli`, so install from a clone, not `go install github.com/...` yet.

```bash
git clone https://github.com/kingdomseed/db-timetables.git
cd db-timetables
make build          # bin/db-timetables-pp-cli
make test
./bin/db-timetables-pp-cli --help
```

`make build-mcp` builds the MCP server binary if you want that surface.

## Examples

Frankfurt (Main) Hbf is EVA `8000105`.

```bash
# Find the station
db-timetables-pp-cli station --pattern "Frankfurt" --json

# Planned departures/arrivals for 16:00 on 17 Aug 2026
db-timetables-pp-cli plan 16 --eva-no 8000105 --date 260817 --json

# Live changes: delays, platform moves, cancellations
db-timetables-pp-cli fchg --eva-no 8000105 --json

# Same, but only what changed in the last two minutes
db-timetables-pp-cli rchg --eva-no 8000105 --json
```

`--json` for scripts. `--agent` for JSON + compact + no prompts. `--dry-run` prints the request and does not call the API. `--select` keeps only the fields you name.

Exit codes: `0` ok, `2` usage, `3` not found, `4` auth, `5` API, `7` rate limit, `10` config.

## What this is not (yet)

The long-term travel helper wants booked-train status from mail, A→B including local transit, and prices. This repo is the official boards slice only. That is the part Marketplace actually gives you for free.

## License

Apache-2.0. Spec and board data come from Deutsche Bahn's Timetables API (CC BY on the data side; you are responsible for your own Marketplace terms). Code and data licenses are separate; see `NOTICE`.
