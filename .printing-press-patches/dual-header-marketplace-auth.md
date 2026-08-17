# Dual-header Deutsche Bahn Marketplace auth

A reprint can overwrite generated auth wiring. Keep this intent:

- Send both required headers on every request: `DB-Client-ID` and `DB-Api-Key`.
- Read client id from `DB_TIMETABLES_CLIENT_ID` (also `DB_CLIENT_ID`).
- Read API key from `DB_TIMETABLES_API_KEY` (also `DB_API_KEY`, then generated `TIMETABLES_CLIENT_SECRET`).
- Treat auth as configured only when both values are present.
- Command examples must use real path values (`8000105`, `BLS`, hour `12`, date `YYMMDD`), never `example-value`.
