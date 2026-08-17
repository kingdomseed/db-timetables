// Copyright 2026 and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"
	"time"

	"db-timetables-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

func newStaleCmd(flags *rootFlags) *cobra.Command {
	var maxAge time.Duration
	cmd := &cobra.Command{
		Use:   "stale",
		Short: "List local timetable slices older than a freshness budget",
		Long:  "Computed freshness insight: compare sync_state last_synced_at against --stale-after and report stale resource types.",
		// pp:data-source computed
		Example: "  db-timetables-pp-cli stale\n  db-timetables-pp-cli stale --stale-after 30m --json",
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
			if maxAge <= 0 {
				maxAge = 30 * time.Minute
			}
			db, err := store.OpenWithContext(cmd.Context(), defaultDBPath("db-timetables-pp-cli"))
			if err != nil {
				return err
			}
			defer db.Close()
			now := time.Now().UTC()
			items := []map[string]any{}
			staleCount := 0
			for _, name := range defaultSyncResources {
				_, lastSynced, count, err := db.GetSyncState(name)
				if err != nil {
					return err
				}
				age := time.Duration(0)
				isStale := lastSynced.IsZero()
				if !lastSynced.IsZero() {
					age = now.Sub(lastSynced)
					isStale = age > maxAge
				}
				if isStale {
					staleCount++
				}
				items = append(items, map[string]any{
					"resource_type": name,
					"count":         count,
					"last_synced":   lastSynced,
					"age":           age.String(),
					"stale":         isStale,
				})
			}
			pct := 0.0
			if len(defaultSyncResources) > 0 {
				pct = float64(staleCount) * 100 / float64(len(defaultSyncResources))
			}
			out := map[string]any{
				"source":        "computed",
				"max_age":       maxAge.String(),
				"stale_count":   staleCount,
				"stale_percent": pct,
				"items":         items,
			}
			if flags != nil && (flags.asJSON || flags.agent) {
				return printJSONFiltered(cmd.OutOrStdout(), out, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Stale resources: %d/%d (%.0f%%) older than %s\n", staleCount, len(defaultSyncResources), pct, maxAge)
			for _, item := range items {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s stale=%v age=%v count=%v\n", item["resource_type"], item["stale"], item["age"], item["count"])
			}
			return nil
		},
	}
	cmd.Flags().DurationVar(&maxAge, "stale-after", 30*time.Minute, "Maximum acceptable age of synced resources")
	return cmd
}
