// Copyright 2026 and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"

	"db-timetables-pp-cli/internal/store"
)

func collectCacheReport(ctx context.Context, staleAfterSpec string) map[string]any {
	report := map[string]any{}
	dbPath := defaultDBPath("db-timetables-pp-cli")
	report["db_path"] = dbPath

	fi, err := os.Stat(dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			report["status"] = "unknown"
			report["hint"] = "Database not created yet; run 'db-timetables-pp-cli sync' to hydrate."
			return report
		}
		report["status"] = "error"
		report["error"] = err.Error()
		return report
	}
	report["db_bytes"] = fi.Size()

	s, err := store.OpenWithContext(ctx, dbPath)
	if err != nil {
		report["status"] = "error"
		report["error"] = err.Error()
		return report
	}
	defer s.Close()

	if v, verr := s.SchemaVersion(); verr == nil {
		report["schema_version"] = v
	}

	staleAfter := 30 * time.Minute
	if staleAfterSpec != "" {
		if d, derr := time.ParseDuration(staleAfterSpec); derr == nil {
			staleAfter = d
		}
	}

	rows, qerr := s.DB().Query(`SELECT resource_type, COALESCE(total_count, 0), last_synced_at FROM sync_state ORDER BY resource_type`)
	if qerr != nil {
		report["status"] = "unknown"
		report["hint"] = "No sync state recorded; run 'db-timetables-pp-cli sync' to populate."
		return report
	}
	defer rows.Close()

	var resources []map[string]any
	fresh := true
	haveAny := false
	oldest := time.Duration(0)
	for rows.Next() {
		var rtype string
		var count int64
		var lastSynced sql.NullTime
		if err := rows.Scan(&rtype, &count, &lastSynced); err != nil {
			continue
		}
		r := map[string]any{"type": rtype, "rows": count}
		if lastSynced.Valid {
			haveAny = true
			r["last_synced_at"] = lastSynced.Time.UTC().Format(time.RFC3339)
			age := time.Since(lastSynced.Time)
			r["staleness"] = age.Round(time.Minute).String()
			if age > staleAfter {
				fresh = false
			}
			if age > oldest {
				oldest = age
			}
		} else {
			r["staleness"] = "never"
			fresh = false
		}
		resources = append(resources, r)
	}
	report["resources"] = resources
	report["stale_after"] = staleAfter.String()

	switch {
	case !haveAny && len(resources) == 0:
		report["status"] = "empty"
		report["hint"] = "No sync recorded; run 'db-timetables-pp-cli sync' to hydrate API-backed resources."
	case fresh:
		report["status"] = "fresh"
	default:
		report["status"] = "stale"
		report["oldest_age"] = oldest.Round(time.Minute).String()
		report["hint"] = "Some resources are older than stale_after; run 'db-timetables-pp-cli sync' to refresh."
	}
	return report
}

func renderCacheReport(w io.Writer, rep map[string]any) {
	status, _ := rep["status"].(string)
	indicator := green("OK")
	switch status {
	case "stale":
		indicator = yellow("WARN")
	case "error":
		indicator = red("FAIL")
	case "unknown", "empty":
		indicator = yellow("INFO")
	}
	fmt.Fprintf(w, "  %s Cache: %s\n", indicator, status)
	if v, ok := rep["db_path"]; ok {
		fmt.Fprintf(w, "    db_path: %v\n", v)
	}
	if v, ok := rep["schema_version"]; ok {
		fmt.Fprintf(w, "    schema_version: %v\n", v)
	}
	if hint, ok := rep["hint"]; ok {
		fmt.Fprintf(w, "    hint: %v\n", hint)
	}
}
