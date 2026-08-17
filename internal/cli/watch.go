// Copyright 2026 Jason Holt and contributors. Licensed under Apache-2.0. See LICENSE.
// Novel command. generate --force preserves implemented bodies.
// pp:data-source live

package cli

import (
	"strings"

	"db-timetables-pp-cli/internal/overlay"
	"github.com/spf13/cobra"
)

func newNovelWatchCmd(flags *rootFlags) *cobra.Command {
	var bf boardFlags
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Apply the last two minutes of rchg onto a live plan+fchg snapshot.",
		Long:  "Use this command to apply recent changes onto a snapshot. Do NOT use it as the first fetch; use 'board' or 'fchg' first.",
		Example: strings.Trim(`
  db-timetables-pp-cli watch --eva-no 8000105 --json
`, "\n"),
		Annotations: map[string]string{
			"mcp:read-only":       "true",
			"pp:happy-args":       "eva-no=8000105",
			"pp:typed-exit-codes": "0,2,3,4,5",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBoardCommand(cmd, flags, bf, "watch", func(b overlay.Board) []overlay.Stop {
				return b.Stops
			})
		},
	}
	bindBoardFlags(cmd, &bf)
	return cmd
}
