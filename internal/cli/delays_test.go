// Copyright 2026 Jason Holt and contributors. Licensed under Apache-2.0. See LICENSE.
// cli-printing-press: novel-scaffold-test
// Novel command scaffold tests. Keep the wiring smoke test and add behavior cases as needed.

package cli

import (
	"bytes"
	"strings"
	"testing"
)

// TestNovelDelaysHelpWires smoke-tests that the delays command
// resolves at runtime and renders useful --help output. Catches wiring
// regressions (missing AddCommand, panicking RunE on --help, etc.) before
// review. Keep this smoke test when adding behavior-specific cases.
func TestNovelDelaysHelpWires(t *testing.T) {
	cmd := RootCmd()
	cmd.SetArgs([]string{"delays", "--help"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("delays --help error = %v (novel command not wired correctly?)", err)
	}
	help := out.String()
	for _, want := range []string{"Usage:", "delays"} {
		if !strings.Contains(help, want) {
			t.Fatalf("delays --help missing %q in output:\n%s", want, help)
		}
	}
}
