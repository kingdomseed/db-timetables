// Copyright 2026 and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"

	"db-timetables-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newCoverageCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coverage",
		Short: "Show which timetable resource types are present in the local store",
		Long:  "Computed coverage insight: expected station/fchg/rchg/plan slices versus COUNT(*) rows already synced.",
		// pp:data-source computed
		Example: "  db-timetables-pp-cli coverage\n  db-timetables-pp-cli coverage --json",
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
			have := map[string]int{}
			for _, row := range counts {
				have[row.ResourceType] = row.Count
			}
			expected := []string{"station", "fchg", "rchg", "plan"}
			present := 0
			gaps := []string{}
			items := []map[string]any{}
			for _, name := range expected {
				n := have[name]
				if n > 0 {
					present++
				} else {
					gaps = append(gaps, name)
				}
				items = append(items, map[string]any{"resource_type": name, "count": n, "present": n > 0})
			}
			pct := 0.0
			if len(expected) > 0 {
				pct = float64(present) * 100 / float64(len(expected))
			}
			out := map[string]any{
				"source":           "computed",
				"expected":         expected,
				"present":          present,
				"coverage_percent": pct,
				"gaps":             gaps,
				"items":            items,
			}
			if flags != nil && (flags.asJSON || flags.agent) {
				return printJSONFiltered(cmd.OutOrStdout(), out, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Coverage: %d/%d (%.0f%%)\n", present, len(expected), pct)
			for _, item := range items {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s present=%v count=%v\n", item["resource_type"], item["present"], item["count"])
			}
			return nil
		},
	}
	return cmd
}
