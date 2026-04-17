// procscope — Process-scoped runtime investigator for Linux.
//
// Usage:
//
//	procscope [flags] [-- command [args...]]
//	procscope -p PID
//	procscope --out case-001 -- ./suspicious-binary
//
// See 'procscope --help' for full usage.
package main

import (
	"fmt"
	"os"

	"github.com/Mutasem-mk4/procscope/internal/cli"
)

func main() {
	rootCmd := cli.NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
