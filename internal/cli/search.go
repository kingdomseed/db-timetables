// Copyright 2026 and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"

	"db-timetables-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newSearchCmd(flags *rootFlags) *cobra.Command {
	var limit int
	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Full-text search locally synced station and timetable slices",
		Long:  "Search the local store with domain SearchStation/SearchFchg/SearchRchg/SearchPlan helpers. Faster than repeating live API calls.",
		// pp:data-source local
		Example: "  db-timetables-pp-cli search BLS\n  db-timetables-pp-cli search Frankfurt --json",
		Annotations: map[string]string{
			"mcp:read-only": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags != nil && flags.dryRun {
				fmt.Fprintln(cmd.ErrOrStderr(), "(dry run - no request sent)")
				if flags.asJSON || flags.agent {
					return printJSONFiltered(cmd.OutOrStdout(), map[string]any{"dry_run": true}, flags)
				}
				return nil
			}
			query := "BLS"
			if len(args) > 0 && args[0] != "" {
				query = args[0]
			}
			if limit <= 0 {
				limit = 20
			}
			db, err := store.OpenWithContext(cmd.Context(), defaultDBPath("db-timetables-pp-cli"))
			if err != nil {
				return err
			}
			defer db.Close()
			stationHits, err := db.SearchStation(query, limit)
			if err != nil {
				return err
			}
			fchgHits, err := db.SearchFchg(query, limit)
			if err != nil {
				return err
			}
			rchgHits, err := db.SearchRchg(query, limit)
			if err != nil {
				return err
			}
			planHits, err := db.SearchPlan(query, limit)
			if err != nil {
				return err
			}
			out := map[string]any{
				"query":   query,
				"station": len(stationHits),
				"fchg":    len(fchgHits),
				"rchg":    len(rchgHits),
				"plan":    len(planHits),
				"hits": map[string]any{
					"station": stationHits,
					"fchg":    fchgHits,
					"rchg":    rchgHits,
					"plan":    planHits,
				},
			}
			if flags != nil && (flags.asJSON || flags.agent) {
				return printJSONFiltered(cmd.OutOrStdout(), out, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Search %q: station=%d fchg=%d rchg=%d plan=%d\n", query, len(stationHits), len(fchgHits), len(rchgHits), len(planHits))
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum hits per resource type")
	return cmd
}
