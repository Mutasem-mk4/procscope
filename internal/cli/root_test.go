//go:build linux

package cli

import (
	"strings"
	"testing"
)

func TestInvalidOptionsFailBeforeTracing(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"ambiguous target", []string{"-p", "4294967295", "-n", "nonexistent"}, "cannot combine"},
		{"negative args", []string{"--max-args", "-1", "--", "/bin/true"}, "--max-args must be positive"},
		{"zero path", []string{"--max-path", "0", "--", "/bin/true"}, "--max-path must be positive"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRootCommand()
			cmd.SetArgs(tc.args)
			if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Execute() = %v, want %q", err, tc.want)
			}
		})
	}
}
