# Build log — db-timetables

- generate wrote the working tree under the managed run dir (govulncheck failed only because the generated go.mod pinned 1.26.5; host Go is 1.26.6).
- Bumped `go.mod` to 1.26.6.
- Implemented traveler novels: `board`, `platforms`, `cancellations`, `delays`, `watch` (plan+fchg overlay; watch also applies rchg).
- Dual-header patch: both Marketplace headers required; aliases `DB_CLIENT_ID` / `DB_API_KEY`; no unused `TIMETABLES_CLIENT_ID`.
- `go test -count=1 ./...` PASS after those edits.
