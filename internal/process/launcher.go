//go:build linux

// Package process provides process tree tracking, command launching, and PID
// attachment for procscope investigations.
package process

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

// Launcher manages launching a command under observation.
type Launcher struct {
	cmd     *exec.Cmd
	done    chan struct{}
	err     error
	mu      sync.Mutex
	started bool
}

// NewLauncher creates a Launcher for the given command and arguments.
// The command is NOT started until Start() is called.
func NewLauncher(args []string) (*Launcher, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	path, err := exec.LookPath(args[0])
	if err != nil {
		return nil, fmt.Errorf("command not found: %s: %w", args[0], err)
	}

	cmd := exec.Command(path, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create a new process group so we can track all children.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return &Launcher{
		cmd:  cmd,
		done: make(chan struct{}),
	}, nil
}

// Start launches the command. Returns the PID of the new process.
func (l *Launcher) Start() (uint32, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.started {
		return 0, fmt.Errorf("command already started")
	}

	if err := l.cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start command: %w", err)
	}

	l.started = true

	// Wait for the command in a background goroutine.
	go func() {
		l.err = l.cmd.Wait()
		close(l.done)
	}()

	return uint32(l.cmd.Process.Pid), nil
}

// Wait blocks until the command exits. Returns the exit error if any.
func (l *Launcher) Wait() error {
	<-l.done
	return l.err
}

// Done returns a channel that is closed when the command exits.
func (l *Launcher) Done() <-chan struct{} {
	return l.done
}

// ExitCode returns the process exit code, or -1 if not available.
func (l *Launcher) ExitCode() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.cmd.ProcessState == nil {
		return -1
	}
	return l.cmd.ProcessState.ExitCode()
}

// Signal sends a signal to the launched process.
func (l *Launcher) Signal(sig os.Signal) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.started || l.cmd.Process == nil {
		return fmt.Errorf("process not started")
	}
	return l.cmd.Process.Signal(sig)
}

// PID returns the PID of the launched process, or 0 if not started.
func (l *Launcher) PID() uint32 {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.started || l.cmd.Process == nil {
		return 0
	}
	return uint32(l.cmd.Process.Pid)
}

// Args returns the full command line that was/will be launched.
func (l *Launcher) Args() []string {
	return l.cmd.Args
}

// CommandString returns the command line as a single string for display.
func (l *Launcher) CommandString() string {
	return strings.Join(l.cmd.Args, " ")
}
