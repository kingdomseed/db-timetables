// Copyright 2026 Jason Holt and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command. generate --force preserves implemented bodies.
// pp:data-source live

package cli

import (
	"strings"

	"db-timetables-pp-cli/internal/overlay"
	"github.com/spf13/cobra"
)

func newNovelCancellationsCmd(flags *rootFlags) *cobra.Command {
	var bf boardFlags
	cmd := &cobra.Command{
		Use:   "cancellations",
		Short: "List cancellations at a station for the selected hour.",
		Long:  "Use this command for cancelled stops this hour. Do NOT use it for delays that still run; use 'delays'.",
		Example: strings.Trim(`
  db-timetables-pp-cli cancellations --eva-no 8000105 --json
`, "\n"),
		Annotations: map[string]string{
			"mcp:read-only":       "true",
			"pp:happy-args":       "eva-no=8000105",
			"pp:typed-exit-codes": "0,2,3,4,5",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBoardCommand(cmd, flags, bf, "cancellations from plan+fchg overlay", func(b overlay.Board) []overlay.Stop {
				return overlay.Cancellations(b.Stops)
			})
		},
	}
	bindBoardFlags(cmd, &bf)
	return cmd
}
