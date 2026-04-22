//go:build linux

// Package cli implements the procscope command-line interface.
package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/Mutasem-mk4/procscope/internal/caps"
	"github.com/Mutasem-mk4/procscope/internal/events"
	"github.com/Mutasem-mk4/procscope/internal/k8s"
	"github.com/Mutasem-mk4/procscope/internal/output"
	"github.com/Mutasem-mk4/procscope/internal/process"
	"github.com/Mutasem-mk4/procscope/internal/tracer"
	"github.com/Mutasem-mk4/procscope/internal/version"
)

// Options holds all CLI flag values.
type Options struct {
	// Target specification
	PID         uint32
	ProcessName string

	// Output
	OutputDir   string
	JSONLPath   string
	SummaryPath string
	NoColor     bool
	Quiet       bool
	JSON        bool
	K8s         bool

	// Tuning
	MaxArgs    int
	MaxPathLen int

	// Privilege
	SkipChecks bool
}

// NewRootCommand creates the root cobra command.
func NewRootCommand() *cobra.Command {
	opts := &Options{}

	rootCmd := &cobra.Command{
		Use:   "procscope [flags] [-- command [args...]]",
		Short: "Process-scoped runtime investigator for Linux",
		Long: `procscope — Process-scoped runtime investigation tool.

Launch a command under observation, or attach to an existing process,
and observe its runtime behavior: process lifecycle, file activity,
network activity, privilege transitions, and more.

Designed for security research, malware triage, incident response,
and deep debugging on Linux hosts.

Requires: Linux kernel 5.8+, BTF, root or CAP_BPF+CAP_PERFMON.

Examples:
  # Trace a command
  procscope -- ./suspicious-binary

  # Attach to a running process
  procscope -p 1234

  # Save evidence bundle
  procscope --out case-001 -- ./installer.sh

  # Stream events as JSONL
  procscope --jsonl events.jsonl -- ./tool

  # Generate a Markdown report
  procscope --summary report.md -- ./script.sh`,

		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.Full(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args, opts)
		},
	}

	// Flags
	f := rootCmd.Flags()
	f.Uint32VarP(&opts.PID, "pid", "p", 0, "Attach to an existing process by PID")
	f.StringVarP(&opts.ProcessName, "name", "n", "", "Attach to a process by name (first match)")
	f.StringVarP(&opts.OutputDir, "out", "o", "", "Evidence bundle output directory")
	f.BoolVarP(&opts.JSON, "json", "j", false, "Stream events as structured JSON logs to stdout instead of timeline")
	f.StringVar(&opts.JSONLPath, "jsonl", "", "Write events as JSONL to file (use - for stdout)")
	f.StringVar(&opts.SummaryPath, "summary", "", "Write Markdown summary to file")
	f.BoolVar(&opts.NoColor, "no-color", false, "Disable colored output")
	f.BoolVarP(&opts.Quiet, "quiet", "q", false, "Suppress live timeline (only write to files)")
	f.BoolVar(&opts.K8s, "k8s", false, "Enable Kubernetes Pod metadata resolution via local API Server (requires RBAC)")
	f.IntVar(&opts.MaxArgs, "max-args", 64, "Maximum number of argv elements to capture")
	f.IntVar(&opts.MaxPathLen, "max-path", 4096, "Maximum path string length")
	f.BoolVar(&opts.SkipChecks, "skip-checks", false, "Skip privilege and kernel checks (use at own risk)")

	// Shell completions subcommand
	rootCmd.AddCommand(newCompletionCmd())

	return rootCmd
}

func run(cmd *cobra.Command, args []string, opts *Options) error {
	// If native JSON logging is enabled via -j/--json
	if opts.JSON {
		opts.Quiet = true // suppress human-readable timeline
		if opts.JSONLPath == "" {
			opts.JSONLPath = "-" // default json output to stdout
		}
	}

	// Determine mode: launch command or attach to PID
	hasCommand := cmd.ArgsLenAtDash() >= 0 || len(args) > 0
	hasPID := opts.PID != 0
	hasName := opts.ProcessName != ""

	if !hasCommand && !hasPID && !hasName {
		return fmt.Errorf("specify a command to trace (-- cmd), a PID (-p), or a process name (-n)\n\n" +
			"Run 'procscope --help' for usage examples.")
	}
	if hasCommand && (hasPID || hasName) {
		return fmt.Errorf("cannot combine command tracing with -p/--pid or -n/--name")
	}

	// Privilege check
	if !opts.SkipChecks {
		result := caps.Check()
		if !result.CanProceed() {
			_, _ = _, _ = _, _ = fmt.Fprintln(os.Stderr, result.Summary())
			return fmt.Errorf("privilege check failed — use --skip-checks to override (may cause load failures)")
		}
		if len(result.Warnings) > 0 {
			for _, w := range result.Warnings {
				_, _ = _, _ = _, _ = fmt.Fprintf(os.Stderr, "⚠ %s\n", w)
			}
		}
	}

	// Generate investigation ID
	investigationID := generateInvestigationID()

	// Resolve target PID
	var targetPID uint32
	var launcher *process.Launcher
	var commandLine string

	if hasCommand {
		// Extract command after "--"
		cmdArgs := args
		if dashIdx := cmd.ArgsLenAtDash(); dashIdx >= 0 {
			cmdArgs = args[dashIdx:]
		}
		if len(cmdArgs) == 0 {
			return fmt.Errorf("no command specified after --")
		}
		commandLine = strings.Join(cmdArgs, " ")

		var err error
		launcher, err = process.NewLauncher(cmdArgs)
		if err != nil {
			return fmt.Errorf("failed to create launcher: %w", err)
		}
	} else if hasPID {
		targetPID = opts.PID
		// Verify PID exists
		if _, err := os.Stat(fmt.Sprintf("/proc/%d", targetPID)); err != nil {
			return fmt.Errorf("process %d not found", targetPID)
		}
	} else if hasName {
		pids, err := process.FindPIDByName(opts.ProcessName)
		if err != nil {
			return err
		}
		if len(pids) == 0 {
			return fmt.Errorf("no process found with name: %s", opts.ProcessName)
		}
		targetPID = pids[0]
		if len(pids) > 1 {
			_, _ = _, _ = fmt.Fprintf(os.Stderr, "⚠ Multiple processes match '%s', attaching to PID %d\n",
				opts.ProcessName, targetPID)
		}
	}

	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		_, _ = _, _ = fmt.Fprintln(os.Stderr, "\n⏹ Stopping investigation...")
		cancel()
	}()

	// Set up correlator (with placeholder PID if launching)
	if targetPID == 0 {
		targetPID = 1 // placeholder, updated after launch
	}
	correlator := events.NewCorrelator(investigationID, targetPID, opts.MaxArgs, opts.MaxPathLen)

	// Spin up Kubernetes watcher if requested
	var watcher *k8s.Watcher
	if opts.K8s {
		_, _ = _, _ = fmt.Fprintln(os.Stderr, "🔄 Initializing Kubernetes pod metadata watcher...")
		var err error
		watcher, err = k8s.NewWatcher(ctx)
		if err != nil {
			return fmt.Errorf("kubernetes initialization failed: %w", err)
		}
		correlator.SetK8sResolver(watcher)
		_, _ = _, _ = fmt.Fprintln(os.Stderr, "✅ Kubernetes integration established")
	}

	// Initialize eBPF tracer
	mgr := tracer.NewManager(correlator)
	if err := mgr.Load(); err != nil {
		return fmt.Errorf("eBPF load failed: %w", err)
	}
	defer mgr.Close()

	if err := mgr.Attach(); err != nil {
		return fmt.Errorf("eBPF attach failed: %w", err)
	}

	// If launching, start the command now and track its PID
	startTime := time.Now()
	if launcher != nil {
		pid, err := launcher.Start()
		if err != nil {
			return fmt.Errorf("failed to start command: %w", err)
		}
		targetPID = pid

		// Replace the placeholder root PID with the actual child PID before the
		// process performs its first exec, so launch mode keeps the root event.
		correlator.SetRootPID(targetPID)

		// Track in eBPF
		if err := mgr.TrackPID(targetPID); err != nil {
			return fmt.Errorf("failed to track PID %d: %w", targetPID, err)
		}
		if err := launcher.Continue(); err != nil {
			return fmt.Errorf("failed to resume command: %w", err)
		}

		_, _ = _, _ = fmt.Fprintf(os.Stderr, "🔍 procscope investigation %s\n", investigationID)
		_, _ = _, _ = fmt.Fprintf(os.Stderr, "   Command: %s\n", commandLine)
		_, _ = _, _ = fmt.Fprintf(os.Stderr, "   PID: %d\n\n", targetPID)
	} else {
		// Attach mode: track existing PID and children
		if err := mgr.TrackPID(targetPID); err != nil {
			return fmt.Errorf("failed to track PID %d: %w", targetPID, err)
		}

		// Discover and track existing children
		tree, err := process.BuildTreeFromPID(targetPID)
		if err == nil {
			for _, pid := range tree.DescendantPIDs() {
				if pid != targetPID {
					mgr.TrackPID(pid)
					correlator.TrackPID(pid, targetPID)
				}
			}
		}

		_, _ = _, _ = fmt.Fprintf(os.Stderr, "🔍 procscope investigation %s\n", investigationID)
		_, _ = _, _ = fmt.Fprintf(os.Stderr, "   Attached to PID: %d\n", targetPID)
		_, _ = _, _ = fmt.Fprintf(os.Stderr, "   Press Ctrl+C to stop.\n\n")
	}

	// Set up output sinks
	var jsonWriter *output.JSONWriter
	if opts.JSONLPath != "" {
		var err error
		jsonWriter, err = output.NewJSONWriter(opts.JSONLPath)
		if err != nil {
			return err
		}
		defer jsonWriter.Close()
	}

	colorize := !opts.NoColor && isTerminal()
	timeline := output.NewTimeline(colorize)

	if !opts.Quiet {
		_, _ = _, _ = fmt.Fprintln(os.Stderr, timeline.Header())
	}

	// Collect all events for bundle/summary
	var allEvents []*events.Event
	var evtMu sync.Mutex

	// Start event reader in background
	var readerErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		readerErr = mgr.ReadEvents(ctx)
	}()

	// Consumer: read from correlator and render
	wg.Add(1)
	go func() {
		defer wg.Done()
		for evt := range correlator.Events() {
			// Timeline output
			if !opts.Quiet {
				_, _ = _, _ = fmt.Fprintln(os.Stderr, timeline.RenderEvent(evt))
			}

			// JSONL output
			if jsonWriter != nil {
				jsonWriter.WriteEvent(evt)
			}

			// Collect for bundle
			evtMu.Lock()
			allEvents = append(allEvents, evt)
			evtMu.Unlock()
		}
	}()

	// Wait for process exit or user interrupt
	if launcher != nil {
		select {
		case <-launcher.Done():
			// Process exited naturally
		case <-ctx.Done():
			// User cancelled — terminate the launched process
			launcher.Signal(syscall.SIGTERM)
			select {
			case <-launcher.Done():
			case <-time.After(3 * time.Second):
				launcher.Signal(syscall.SIGKILL)
			}
		}
	} else {
		// Attach mode: wait for context cancellation
		<-ctx.Done()
	}

	// Short grace period for final events
	time.Sleep(200 * time.Millisecond)

	// Stop event reading
	cancel()
	correlator.Close()
	wg.Wait()

	endTime := time.Now()

	// Print summary stats
	_, _ = _, _ = fmt.Fprintf(os.Stderr, "\n%s\n", correlator.Summary())

	if readerErr != nil && ctx.Err() == nil {
		_, _ = _, _ = fmt.Fprintf(os.Stderr, "⚠ Event reader error: %v\n", readerErr)
	}

	// Write evidence bundle
	if opts.OutputDir != "" {
		evtMu.Lock()
		evts := make([]*events.Event, len(allEvents))
		copy(evts, allEvents)
		evtMu.Unlock()

		bundle := &output.Bundle{
			Dir:         opts.OutputDir,
			Correlator:  correlator,
			Events:      evts,
			StartTime:   startTime,
			EndTime:     endTime,
			CommandLine: commandLine,
			TargetPID:   targetPID,
		}

		if err := bundle.Write(); err != nil {
			return fmt.Errorf("failed to write evidence bundle: %w", err)
		}
		_, _ = _, _ = fmt.Fprintf(os.Stderr, "📁 Evidence bundle: %s/\n", opts.OutputDir)
	}

	// Write standalone summary
	if opts.SummaryPath != "" {
		evtMu.Lock()
		evts := make([]*events.Event, len(allEvents))
		copy(evts, allEvents)
		evtMu.Unlock()

		sw := output.NewSummaryWriter(correlator, evts, commandLine, targetPID, startTime, endTime)
		if err := sw.WriteToFile(opts.SummaryPath); err != nil {
			return fmt.Errorf("failed to write summary: %w", err)
		}
		_, _ = _, _ = fmt.Fprintf(os.Stderr, "📝 Summary: %s\n", opts.SummaryPath)
	}

	// Report exit code if we launched a process
	if launcher != nil {
		exitCode := launcher.ExitCode()
		if exitCode != 0 {
			_, _ = fmt.Fprintf(os.Stderr, "⚠ Process exited with code %d\n", exitCode)
		}
		return &ExitError{Code: exitCode}
	}

	return nil
	}

	// ExitError represents an error that should result in a specific exit code.
	type ExitError struct {
	Code int
	}

	func (e *ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
	}

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for procscope.

To load completions:

Bash:
  $ source <(procscope completion bash)
  # Or install permanently:
  $ procscope completion bash > /etc/bash_completion.d/procscope

Zsh:
  $ source <(procscope completion zsh)
  $ compdef _procscope procscope

Fish:
  $ procscope completion fish | source
  $ procscope completion fish > ~/.config/fish/completions/procscope.fish`,
		ValidArgs: []string{"bash", "zsh", "fish"},
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
}

func generateInvestigationID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("ps-%d", time.Now().UnixNano()%100000)
	}
	return fmt.Sprintf("ps-%s", hex.EncodeToString(b))
}

func isTerminal() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
