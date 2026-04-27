//go:build !linux

package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// NewRootCommand creates the root cobra command.
// On non-Linux platforms, all commands return an error.
func NewRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "procscope",
		Short: "Process-scoped runtime investigator for Linux",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("procscope requires Linux (current: %s/%s)", runtime.GOOS, runtime.GOARCH)
		},
	}
}

// ExitError represents an error that should result in a specific exit code.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}
