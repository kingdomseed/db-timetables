// Copyright 2026 and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"db-timetables-pp-cli/internal/store"
	"github.com/spf13/cobra"
)

var defaultSyncResources = []string{"station", "fchg", "rchg", "plan"}

func init() {
	registerNovelCommand(func(root *cobra.Command, flags *rootFlags) {
		root.AddCommand(newSyncCmd(flags))
	})
}

func newSyncCmd(flags *rootFlags) *cobra.Command {
	var evaNo string
	var pattern string
	var date string
	var hour string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Fetch station and timetable slices into the local store",
		Long:  "Live-fetch station metadata plus full/recent changes (and an hourly plan slice) and persist them with domain upserts.",
		// pp:data-source live
		Example: "  db-timetables-pp-cli sync --eva-no 8000105 --pattern BLS",
		Annotations: map[string]string{
			"mcp:read-only": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			if date == "" {
				date = time.Now().UTC().Format("060102")
			}
			if hour == "" {
				hour = time.Now().UTC().Format("15")
			}
			db, err := store.OpenWithContext(cmd.Context(), defaultDBPath("db-timetables-pp-cli"))
			if err != nil {
				return err
			}
			defer db.Close()

			headers := map[string]string{"Accept": "application/xml"}
			summaries := []map[string]any{}

			stationData, err := c.GetWithHeaders(cmd.Context(), "/station/"+pattern, nil, headers)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			ins, upd, err := db.UpsertStation([]json.RawMessage{stationData})
			if err != nil {
				return err
			}
			summaries = append(summaries, map[string]any{"resource": "station", "inserted": ins, "updated": upd})
			_ = db.SaveSyncState("station", pattern, ins+upd)

			fchgData, err := c.GetWithHeaders(cmd.Context(), "/fchg/"+evaNo, nil, headers)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			ins, upd, err = db.UpsertFchg([]json.RawMessage{fchgData})
			if err != nil {
				return err
			}
			summaries = append(summaries, map[string]any{"resource": "fchg", "inserted": ins, "updated": upd})
			_ = db.SaveSyncState("fchg", evaNo, ins+upd)

			rchgData, err := c.GetWithHeaders(cmd.Context(), "/rchg/"+evaNo, nil, headers)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			ins, upd, err = db.UpsertRchg([]json.RawMessage{rchgData})
			if err != nil {
				return err
			}
			summaries = append(summaries, map[string]any{"resource": "rchg", "inserted": ins, "updated": upd})
			_ = db.SaveSyncState("rchg", evaNo, ins+upd)

			planData, err := c.GetWithHeaders(cmd.Context(), "/plan/"+evaNo+"/"+date+"/"+hour, nil, headers)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			ins, upd, err = db.UpsertPlan([]json.RawMessage{planData})
			if err != nil {
				return err
			}
			summaries = append(summaries, map[string]any{"resource": "plan", "inserted": ins, "updated": upd})
			_ = db.SaveSyncState("plan", evaNo+"/"+date+"/"+hour, ins+upd)

			out := map[string]any{"synced": summaries, "eva_no": evaNo, "pattern": pattern, "resources": defaultSyncResources}
			if flags.asJSON || flags.agent {
				return printJSONFiltered(cmd.OutOrStdout(), out, flags)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Synced station %s / EVA %s\n", pattern, evaNo)
			for _, row := range summaries {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s inserted=%v updated=%v\n", row["resource"], row["inserted"], row["updated"])
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&evaNo, "eva-no", "8000105", "Station EVA-number")
	cmd.Flags().StringVar(&pattern, "pattern", "BLS", "Station name prefix, EVA number, or DS100 code")
	cmd.Flags().StringVar(&date, "date", "", "Plan date in YYMMDD (default: today UTC)")
	cmd.Flags().StringVar(&hour, "hour", "", "Plan hour in HH (default: current UTC hour)")
	return cmd
}
