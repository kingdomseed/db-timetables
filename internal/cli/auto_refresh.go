// Copyright 2026 and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"db-timetables-pp-cli/internal/cliutil"
	"db-timetables-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

var readCommandResources = map[string][]string{
	"db-timetables-pp-cli station":  {"station"},
	"db-timetables-pp-cli fchg":     {"fchg"},
	"db-timetables-pp-cli rchg":     {"rchg"},
	"db-timetables-pp-cli plan":     {"plan"},
	"db-timetables-pp-cli search":   {"station", "fchg", "rchg", "plan"},
	"db-timetables-pp-cli health":   {"station", "fchg", "rchg", "plan"},
	"db-timetables-pp-cli coverage": {"station", "fchg", "rchg", "plan"},
	"db-timetables-pp-cli stale":    {"station", "fchg", "rchg", "plan"},
}

func cachePolicy() cliutil.Policy {
	return cliutil.Policy{
		StaleAfter:  30 * time.Minute,
		EnvOptOut:   "DB_TIMETABLES_NO_AUTO_REFRESH",
		ShareEnabled: false,
	}
}

// autoRefreshIfStale inspects local sync_state and prints a stderr hint when
// data is stale. It never dials the Deutsche Bahn API.
func autoRefreshIfStale(cmd *cobra.Command, flags *rootFlags) {
	if cmd == nil || flags == nil || flags.dryRun || flags.dataSource == "live" {
		return
	}
	resources := readCommandResources[cmd.CommandPath()]
	if len(resources) == 0 {
		return
	}
	dbPath := defaultDBPath("db-timetables-pp-cli")
	if _, err := os.Stat(dbPath); err != nil {
		return
	}
	db, err := store.OpenWithContext(cmd.Context(), dbPath)
	if err != nil {
		return
	}
	defer db.Close()
	started := time.Now()
	decision, err := cliutil.EnsureFresh(cmd.Context(), db.DB(), resources, cachePolicy())
	meta := cliutil.FreshnessMeta{
		Decision:  decision.String(),
		Ran:       false,
		Resources: resources,
		ElapsedMS: time.Since(started).Milliseconds(),
		Source:    flags.dataSource,
	}
	if err != nil {
		meta.Error = err.Error()
		meta.Reason = "error"
	} else {
		switch decision {
		case cliutil.DecisionFresh:
			meta.Reason = "fresh"
		case cliutil.DecisionNoStore:
			meta.Reason = "no-store"
		default:
			meta.Reason = "stale"
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: local %s data is stale; run `db-timetables-pp-cli sync` (no automatic live refresh)\n", strings.Join(resources, ","))
		}
	}
	flags.freshnessMeta = map[string]any{
		"decision":  meta.Decision,
		"ran":       meta.Ran,
		"reason":    meta.Reason,
		"resources": meta.Resources,
	}
}
