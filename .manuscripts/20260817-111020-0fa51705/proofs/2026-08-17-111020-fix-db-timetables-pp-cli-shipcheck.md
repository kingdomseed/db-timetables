# Shipcheck — db-timetables

- verify: PASS 32/32 (0 critical). Framework learnings/playbook/profile/workflow EXEC warns only.
- dogfood (structural): WARN — generated dead `--max-age` / unused helpers; not CLI-specific.
- scorecard first pass HOLD because `research.json` was not yet copied into the working tree. Copied `$API_RUN_DIR/research.json` into `$CLI_WORK_DIR/research.json`.
- Official spec reused: `/workspace/printing-press/timetables-openapi.json`.
- Grade on first scorecard: 84/100 A; live_api_verification unverified until Phase 5.
