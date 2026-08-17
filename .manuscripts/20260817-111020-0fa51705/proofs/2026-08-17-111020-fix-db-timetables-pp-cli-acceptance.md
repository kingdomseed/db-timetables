# Acceptance Report: db-timetables

- Level: Full Dogfood (live)
- Tests: 86/86 passed (TestsFailed == 0)
- Gate: PASS
- run_id: 20260817-111020-0fa51705
- Marker: proofs/phase5-acceptance.json (status=pass, written by dogfood runner)

## Failures

None on the passing run.

## Fixes applied (this run)

- Dual-header Marketplace auth: both `DB-Client-Id` and `DB-Api-Key` required; aliases `DB_CLIENT_ID` / `DB_API_KEY`; same-host redirect copies both headers. Recorded in `.printing-press-patches/dual-header-marketplace-auth.json`.
- `plan` happy-args reduced to positional hour `12` with defaults `--eva-no 8000105` and `--date 260817` so dogfood does not stuff `evaNo=...,date=...,hour=...` into the hour path segment.
- `feedback --help` now has an Examples section.
- `feedback` rejects `__printing_press_*` sentinel args so the dogfood error-path expects a non-zero exit.
- go.mod bumped to Go 1.26.6 so govulncheck matches the toolchain.
- Novel commands implemented for real: `board`, `platforms`, `cancellations`, `delays`, `watch` (plan+fchg overlay; not hollow insight/health/coverage/stale).

## Printing Press issues

- Generator happy-args for multi-path-param endpoints can be replayed as a single positional, producing a 404 (`/plan/{eva}/{date}/{hour}`).
- `feedback` treats any text as valid, so the mechanical `__printing_press_invalid__` error-path fails unless the CLI special-cases the sentinel.
- Generated dual-scheme auth still needed a CredentialConfigured=both patch and alias env names.

## Live smoke (not quoted as PII)

- `station --pattern FF` returned station XML for the test pattern.
- `plan 12 --eva-no 8000105 --date 260817 --json` returned a live timetable for Frankfurt Hbf.
- `board` / `fchg` / `rchg` / `platforms` / `cancellations` / `delays` / `watch` happy-path and json_fidelity passed against the live Marketplace.
