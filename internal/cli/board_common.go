// Copyright 2026 Jason Holt and contributors. Licensed under Apache-2.0. See LICENSE.
// Shared live-board fetch/print for traveler novel commands.

package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"db-timetables-pp-cli/internal/overlay"
	"github.com/spf13/cobra"
)

type boardFlags struct {
	evaNo string
	date  string
	hour  string
}

func bindBoardFlags(cmd *cobra.Command, bf *boardFlags) {
	cmd.Flags().StringVar(&bf.evaNo, "eva-no", "", "Station EVA number (for example 8000105)")
	cmd.Flags().StringVar(&bf.date, "date", "", "Date YYMMDD (defaults to today in Europe/Berlin)")
	cmd.Flags().StringVar(&bf.hour, "hour", "", "Hour HH (defaults to the current hour in Europe/Berlin)")
}

func resolveBoardSlice(bf boardFlags) (eva, date, hour string, err error) {
	eva = strings.TrimSpace(bf.evaNo)
	if eva == "" {
		return "", "", "", usageErr(fmt.Errorf("--eva-no is required"))
	}
	date = strings.TrimSpace(bf.date)
	hour = strings.TrimSpace(bf.hour)
	defDate, defHour := overlay.NowSlice(time.Now())
	if date == "" {
		date = defDate
	}
	if hour == "" {
		hour = defHour
	}
	if len(date) != 6 {
		return "", "", "", usageErr(fmt.Errorf("--date must be YYMMDD"))
	}
	if len(hour) == 1 {
		hour = "0" + hour
	}
	if len(hour) != 2 {
		return "", "", "", usageErr(fmt.Errorf("--hour must be HH"))
	}
	return eva, date, hour, nil
}

func fetchOverlayBoard(ctx context.Context, flags *rootFlags, eva, date, hour string) (overlay.Board, error) {
	if err := validateDataSourceStrategy(flags, "live"); err != nil {
		return overlay.Board{}, err
	}
	c, err := flags.newClient()
	if err != nil {
		return overlay.Board{}, err
	}
	headers := map[string]string{"Accept": "application/xml"}
	planPath := "/plan/" + eva + "/" + date + "/" + hour
	planRaw, err := c.GetWithHeaders(ctx, planPath, nil, headers)
	if err != nil {
		return overlay.Board{}, classifyAPIError(err, flags)
	}
	fchgRaw, err := c.GetWithHeaders(ctx, "/fchg/"+eva, nil, headers)
	if err != nil {
		return overlay.Board{}, classifyAPIError(err, flags)
	}
	station, planned, err := overlay.ParseTimetable(planRaw)
	if err != nil {
		return overlay.Board{}, err
	}
	_, changes, err := overlay.ParseTimetable(fchgRaw)
	if err != nil {
		return overlay.Board{}, err
	}
	stops := overlay.Overlay(planned, changes)
	if station == "" {
		for _, s := range stops {
			if s.Station != "" {
				station = s.Station
				break
			}
		}
	}
	return overlay.Board{EVA: eva, Date: date, Hour: hour, Station: station, Stops: stops}, nil
}

func fetchWatchBoard(ctx context.Context, flags *rootFlags, eva, date, hour string) (overlay.Board, error) {
	board, err := fetchOverlayBoard(ctx, flags, eva, date, hour)
	if err != nil {
		return overlay.Board{}, err
	}
	c, err := flags.newClient()
	if err != nil {
		return overlay.Board{}, err
	}
	headers := map[string]string{"Accept": "application/xml"}
	rchgRaw, err := c.GetWithHeaders(ctx, "/rchg/"+eva, nil, headers)
	if err != nil {
		return overlay.Board{}, classifyAPIError(err, flags)
	}
	_, recent, err := overlay.ParseTimetable(rchgRaw)
	if err != nil {
		return overlay.Board{}, err
	}
	board.Stops = overlay.RecentOnly(overlay.MarkRecent(board.Stops, recent))
	return board, nil
}

func printBoardResult(cmd *cobra.Command, flags *rootFlags, board overlay.Board, rows []overlay.Stop) error {
	payload := map[string]any{
		"eva":     board.EVA,
		"date":    board.Date,
		"hour":    board.Hour,
		"station": board.Station,
		"stops":   rows,
	}
	if !wantsHumanTable(cmd.OutOrStdout(), flags) {
		return printJSONFiltered(cmd.OutOrStdout(), payload, flags)
	}
	if len(rows) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No matching trains.")
		return nil
	}
	items := make([]map[string]any, 0, len(rows))
	for _, s := range rows {
		delay := ""
		if s.DelayMinutes != nil {
			delay = fmt.Sprintf("%d", *s.DelayMinutes)
		}
		items = append(items, map[string]any{
			"train":     s.Train,
			"planned":   s.PlannedTime,
			"delay":     delay,
			"platform":  s.Platform,
			"cancelled": s.Cancelled,
			"changed":   s.PlatformChanged,
			"dest":      s.Destination,
		})
	}
	return printAutoTable(cmd.OutOrStdout(), items)
}

func runBoardCommand(cmd *cobra.Command, flags *rootFlags, bf boardFlags, action string, filter func(overlay.Board) []overlay.Stop) error {
	if dryRunOK(flags) {
		return writeDryRun(cmd.OutOrStdout(), flags, action)
	}
	eva, date, hour, err := resolveBoardSlice(bf)
	if err != nil {
		_ = cmd.Usage()
		return err
	}
	ctx, cancel := boundCtx(cmd.Context(), flags)
	defer cancel()
	var board overlay.Board
	if action == "watch" {
		board, err = fetchWatchBoard(ctx, flags, eva, date, hour)
	} else {
		board, err = fetchOverlayBoard(ctx, flags, eva, date, hour)
	}
	if err != nil {
		return err
	}
	return printBoardResult(cmd, flags, board, filter(board))
}
