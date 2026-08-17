// Copyright 2026 Jason Holt and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command. generate --force preserves implemented bodies.
// pp:data-source live

package cli

import (
	"strings"

	"db-timetables-pp-cli/internal/overlay"
	"github.com/spf13/cobra"
)

func newNovelDelaysCmd(flags *rootFlags) *cobra.Command {
	var bf boardFlags
	cmd := &cobra.Command{
		Use:   "delays",
		Short: "List late trains at a station with delay minutes.",
		Long:  "Use this command for late trains. Do NOT use it for cancellations; use 'cancellations'.",
		Example: strings.Trim(`
  db-timetables-pp-cli delays --eva-no 8000105 --json
`, "\n"),
		Annotations: map[string]string{
			"mcp:read-only":       "true",
			"pp:happy-args":       "eva-no=8000105",
			"pp:typed-exit-codes": "0,2,3,4,5",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBoardCommand(cmd, flags, bf, "delays from plan+fchg overlay", func(b overlay.Board) []overlay.Stop {
				return overlay.Delays(b.Stops)
			})
		},
	}
	bindBoardFlags(cmd, &bf)
	return cmd
}
