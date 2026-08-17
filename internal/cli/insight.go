// Copyright 2026 and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newInsightCmd(flags *rootFlags) *cobra.Command {
	var pattern string
	var evaNo string
	var date string
	var hour string

	cmd := &cobra.Command{
		Use:   "insight",
		Short: "Explain how to turn a station into plan, full-change, and recent-change queries",
		Long:  "Computed guidance for Deutsche Bahn Timetables: resolve a station pattern to an EVA number, then name the plan/fchg/rchg commands that fetch that station.",
		// pp:data-source computed
		Example: "  db-timetables-pp-cli insight --pattern BLS\n  db-timetables-pp-cli insight --eva-no 8000105 --hour 12",
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
			if strings.TrimSpace(pattern) == "" && strings.TrimSpace(evaNo) == "" {
				pattern = "BLS"
			}
			if strings.TrimSpace(date) == "" {
				date = time.Now().UTC().Format("060102")
			}
			if strings.TrimSpace(hour) == "" {
				hour = time.Now().UTC().Format("15")
			}
			station := strings.TrimSpace(evaNo)
			if station == "" {
				station = strings.TrimSpace(pattern)
			}
			out := map[string]any{
				"source":  "computed",
				"station": station,
				"next": []map[string]string{
					{
						"goal":    "Resolve station names, EVA numbers, or DS100 codes",
						"command": "db-timetables-pp-cli station --pattern " + strings.TrimSpace(firstNonEmpty(pattern, station)),
					},
					{
						"goal":    "Planned arrivals and departures for one hourly slice",
						"command": fmt.Sprintf("db-timetables-pp-cli plan %s --eva-no %s --date %s", hour, firstNonEmpty(evaNo, "8000105"), date),
					},
					{
						"goal":    "All known changes from now on",
						"command": "db-timetables-pp-cli fchg --eva-no " + firstNonEmpty(evaNo, "8000105"),
					},
					{
						"goal":    "Changes that became known in the last two minutes",
						"command": "db-timetables-pp-cli rchg --eva-no " + firstNonEmpty(evaNo, "8000105"),
					},
				},
				"auth": []string{"DB_TIMETABLES_CLIENT_ID", "DB_TIMETABLES_API_KEY"},
				"notes": []string{
					"Date is YYMMDD and hour is HH (00-23).",
					"Both Marketplace headers are required: DB-Client-Id and DB-Api-Key.",
					"8000105 is Frankfurt (Main) Hbf; BLS is Berlin Hbf (DS100).",
				},
			}
			if flags.asJSON || flags.agent {
				return printJSONFiltered(cmd.OutOrStdout(), out, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Station: %s\n\n", station)
			fmt.Fprintln(cmd.OutOrStdout(), "Next commands:")
			for _, step := range out["next"].([]map[string]string) {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n    %s\n", step["goal"], step["command"])
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&pattern, "pattern", "", "Station name prefix, EVA number, DS100/RL100 code, or wildcard")
	cmd.Flags().StringVar(&evaNo, "eva-no", "", "Station EVA-number")
	cmd.Flags().StringVar(&date, "date", "", "Plan date in YYMMDD (default: today UTC)")
	cmd.Flags().StringVar(&hour, "hour", "", "Plan hour in HH (default: current UTC hour)")
	return cmd
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
