// Copyright 2026 and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"

	"db-timetables-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newHealthCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Summarize locally synced timetable resources with counts and percentages",
		Long:  "Computed local-store insight: GROUP BY resource_type with COUNT(*) and share percentages. Run sync first to populate.",
		// pp:data-source computed
		Example: "  db-timetables-pp-cli health\n  db-timetables-pp-cli health --json",
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
			db, err := store.OpenWithContext(cmd.Context(), defaultDBPath("db-timetables-pp-cli"))
			if err != nil {
				return err
			}
			defer db.Close()
			counts, err := db.CountByResourceType()
			if err != nil {
				return err
			}
			total := 0
			for _, row := range counts {
				total += row.Count
			}
			rows := make([]map[string]any, 0, len(counts))
			for _, row := range counts {
				pct := 0.0
				if total > 0 {
					pct = float64(row.Count) * 100 / float64(total)
				}
				rows = append(rows, map[string]any{
					"resource_type": row.ResourceType,
					"count":         row.Count,
					"percentage":    pct,
				})
			}
			out := map[string]any{"source": "computed", "total": total, "by_type": rows}
			if flags != nil && (flags.asJSON || flags.agent) {
				return printJSONFiltered(cmd.OutOrStdout(), out, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Local timetable stats: %d rows\n", total)
			for _, row := range rows {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s  %v  (%.1f%%)\n", row["resource_type"], row["count"], row["percentage"])
			}
			return nil
		},
	}
	return cmd
}
