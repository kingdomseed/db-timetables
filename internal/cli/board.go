// Copyright 2026 Jason Holt and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command. generate --force preserves implemented bodies.
// pp:data-source live

package cli

import (
	"strings"

	"db-timetables-pp-cli/internal/overlay"
	"github.com/spf13/cobra"
)

func newNovelBoardCmd(flags *rootFlags) *cobra.Command {
	var bf boardFlags
	cmd := &cobra.Command{
		Use:   "board",
		Short: "Live station board: plan plus delays, platform moves, and cancels.",
		Long:  "Use this command for the live board at one station this hour. Do NOT use it for A to B journeys; this API cannot plan trips.",
		Example: strings.Trim(`
  db-timetables-pp-cli board --eva-no 8000105 --json
  db-timetables-pp-cli board --eva-no 8000105 --agent --select stops.train,stops.planned_time,stops.delay_minutes,stops.platform,stops.cancelled
`, "\n"),
		Annotations: map[string]string{
			"mcp:read-only":       "true",
			"pp:happy-args":       "eva-no=8000105",
			"pp:typed-exit-codes": "0,2,3,4,5",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBoardCommand(cmd, flags, bf, "board plan+fchg overlay", func(b overlay.Board) []overlay.Stop {
				return b.Stops
			})
		},
	}
	bindBoardFlags(cmd, &bf)
	return cmd
}
