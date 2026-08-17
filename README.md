# Db Timetables CLI

API for passenger information for train stations operated by DB Station&Service AG

## Install

The recommended path installs both the `db-timetables-pp-cli` binary and the `pp-db-timetables` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install db-timetables
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install db-timetables --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install db-timetables --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install db-timetables --agent claude-code
npx -y @mvanhorn/printing-press-library install db-timetables --agent claude-code --agent codex
```

### Without Node

The generated install path is category-agnostic until this CLI is published. If `npx` is not available before publish, install Node or use the category-specific Go fallback from the public-library entry after publish.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/db-timetables-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install db-timetables --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-db-timetables --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-db-timetables --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install db-timetables --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/db-timetables-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.
3. Fill in `DB_TIMETABLES_CLIENT_ID` and `DB_TIMETABLES_API_KEY` when Claude Desktop prompts you.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


Install the MCP binary from this CLI's published public-library entry or pre-built release.

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "db-timetables": {
      "command": "db-timetables-pp-mcp",
      "env": {
        "DB_TIMETABLES_CLIENT_ID": "<your-client-id>",
        "DB_TIMETABLES_API_KEY": "<your-api-key>"
      }
    }
  }
}
```

</details>

## Quick Start

```bash
db-timetables-pp-cli insight --pattern BLS --json

db-timetables-pp-cli station --pattern BLS --dry-run

db-timetables-pp-cli fchg --eva-no 8000105 --dry-run

db-timetables-pp-cli rchg --eva-no 8000105 --dry-run

db-timetables-pp-cli plan 12 --eva-no 8000105 --date 220930 --dry-run

db-timetables-pp-cli health --json

db-timetables-pp-cli coverage --json

db-timetables-pp-cli stale --json

```

## Unique Features

These capabilities aren't available in any other tool for this API.
- **`insight`** — Computed guidance that turns a station pattern or EVA number into plan/fchg/rchg commands.

  ```bash
  db-timetables-pp-cli insight --pattern BLS --json
  ```
- **`health`** — GROUP BY counts and percentages over locally synced timetable resources.

  ```bash
  db-timetables-pp-cli health --json
  ```
- **`coverage`** — Coverage of expected station/fchg/rchg/plan slices in the local store.

  ```bash
  db-timetables-pp-cli coverage --json
  ```
- **`stale`** — Freshness insight for synced timetable slices versus a stale-after budget.

  ```bash
  db-timetables-pp-cli stale --stale-after 30m --json
  ```

## Recipes

### 

```bash
db-timetables-pp-cli station --pattern BLS --dry-run
```

### 

```bash
db-timetables-pp-cli plan 12 --eva-no 8000105 --date 220930 --dry-run
```

## Usage

Run `db-timetables-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

This CLI separates local files into four path kinds:

| Kind | Contents |
|------|----------|
| `config` | User-editable settings such as `config.toml` and saved profiles |
| `data` | Durable local data: `credentials.toml`, `data.db`, cookies, browser-session proof files, and other auth sidecars |
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

Under `DB_TIMETABLES_HOME=/srv/db-timetables`, the four dirs resolve to `/srv/db-timetables/config`, `/srv/db-timetables/data`, `/srv/db-timetables/state`, and `/srv/db-timetables/cache`.

MCP servers do not receive CLI flags from the host. Put relocation in the host `env` block:

```json
{
  "mcpServers": {
    "db-timetables": {
      "command": "db-timetables-pp-mcp",
      "env": {
        "DB_TIMETABLES_HOME": "/srv/db-timetables"
      }
    }
  }
}
```

Precedence matters in fleets: an ambient per-kind variable such as `DB_TIMETABLES_DATA_DIR` overrides an explicit `--home` for that kind. Use `DB_TIMETABLES_HOME` or the per-kind variables for durable fleet relocation; treat `--home` as the weaker per-invocation lever.

Relocation is one-way. Unsetting `DB_TIMETABLES_HOME` does not move files back to platform defaults, and `doctor` cannot find credentials left under a former root. Move the files manually before unsetting relocation variables.

Existing installs keep working because the platform-default rung matches the legacy layout. On the first auth write, stored secrets leave `config.toml` and are consolidated into `credentials.toml` under the data directory. Run `db-timetables-pp-cli doctor --fail-on warn` to check path and credential-location warnings in automation.

## Commands

### fchg

Manage fchg

- **`db-timetables-pp-cli fchg`** - Returns a Timetable object (see Timetable) that contains all known changes for the station given by evaNo.

The data includes all known changes from now on until ndefinitely into the future. Once changes become obsolete (because their trip departs from the station) they are removed from this resource.

Changes may include messages. On event level, they usually contain one or more of the 'changed' attributes ct, cp, cs or cpth. Changes may also include 'planned' attributes if there is no associated planned data for the change (e.g. an unplanned stop or trip).

Full changes are updated every 30s and should be cached for that period by web caches.

### plan

Manage plan

- **`db-timetables-pp-cli plan <hour>`** - Returns a Timetable object (see Timetable) that contains planned data for the specified station (evaNo) within the hourly time slice given by date (format YYMMDD) and hour (format HH). The data includes stops for all trips that arrive or depart within that slice. There is a small overlap between slices since some trips arrive in one slice and depart in another.

Planned data does never contain messages. On event level, planned data contains the 'plannned' attributes pt, pp, ps and ppth while the 'changed' attributes ct, cp, cs and cpth are absent.

Planned data is generated many hours in advance and is static, i.e. it does never change. It should be cached by web caches.public interface allows access to information about a station.

### rchg

Manage rchg

- **`db-timetables-pp-cli rchg`** - Returns a Timetable object (see Timetable) that contains all recent changes for the station given by evaNo. Recent changes are always a subset of the full changes. They may equal full changes but are typically much smaller. Data includes only those changes that became known within the last 2 minutes.

A client that updates its state in intervals of less than 2 minutes should load full changes initially and then proceed to periodically load only the recent changes in order to save bandwidth.

Recent changes are updated every 30s as well and should be cached for that period by web caches.

### station

Manage station

- **`db-timetables-pp-cli station`** - This public interface allows access to information about a station.


### Self-learning loop

This CLI caches per-question discovery so repeat queries skip the walk and structurally similar queries get answered via entity substitution. The loop also self-captures: every invocation is journaled locally, and failed-flag corrections plus fresh teaches surface as candidates on the next `recall` for confirm/reject judgment. Agents call `recall` before discovery and fire `teach &` after answering. See the `## Automatic learning` section in `SKILL.md` for the full protocol.

- **`db-timetables-pp-cli recall <query>`** - Look up cached resources for a query before running discovery
- **`db-timetables-pp-cli teach`** - Record a query -> resource mapping (silent on success, safe to background with `&`)
- **`db-timetables-pp-cli learnings list`** - Inspect taught rows
- **`db-timetables-pp-cli learnings forget <query>`** - Undo a teach
- **`db-timetables-pp-cli learnings candidates`** - List auto-captured candidates awaiting confirm/reject
- **`db-timetables-pp-cli learnings stats`** - Local loop metrics: recall hit rate, teach-to-reuse, playbook resolution, candidate counts
- **`db-timetables-pp-cli teach-pattern`** - Install a query/resource template up front
- **`db-timetables-pp-cli teach-lookup`** - Add an entity mapping (e.g. country code, team alias) for pattern substitution

Pass `--no-learn` or set `DB_TIMETABLES_NO_LEARN=true` to disable the loop for deterministic flows.

The local store's schema version stamp is one-way: once this version of `db-timetables-pp-cli` opens the database, older binaries refuse it with a version error — upgrade the binary rather than downgrading.

## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
db-timetables-pp-cli fchg --eva-no 8000105

# JSON for scripting and agents
db-timetables-pp-cli fchg --eva-no 8000105 --json
# Filter to specific fields
db-timetables-pp-cli fchg --eva-no 8000105 --json --select eva,m,s

# Dry run — show the request without sending
db-timetables-pp-cli fchg --eva-no 8000105 --dry-run

# Agent mode — JSON + compact + no prompts in one flag
db-timetables-pp-cli fchg --eva-no 8000105 --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select <field>[,<field>...]` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Read-only by default** - this CLI does not create, update, delete, publish, send, or mutate remote resources
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
db-timetables-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Run `db-timetables-pp-cli doctor` to see the resolved config, data, state, and cache directories. The platform-default config path is `~/.config/timetables-pp-cli/config.toml`; `--home`, `DB_TIMETABLES_HOME`, and per-kind env vars can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `DB_TIMETABLES_CLIENT_ID` | per_call | Yes | Marketplace client id sent as `DB-Client-Id`. Also accepts `DB_CLIENT_ID`. |
| `DB_TIMETABLES_API_KEY` | per_call | Yes | Marketplace client secret sent as `DB-Api-Key`. Also accepts `DB_API_KEY` or `TIMETABLES_CLIENT_SECRET`. |

### agentcookie (optional)

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `db-timetables-pp-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `db-timetables-pp-cli doctor` to check credentials
- Verify both environment variables are set: `DB_TIMETABLES_CLIENT_ID` and `DB_TIMETABLES_API_KEY`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
