//go:build linux

// Package process provides process tree tracking, command launching, and PID
// attachment for procscope investigations.
package process

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Launcher manages launching a command under observation.
type Launcher struct {
	cmd        *exec.Cmd
	done       chan struct{}
	err        error
	mu         sync.Mutex
	started    bool
	commandArg []string
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

	// Start the child in a stopped shell wrapper so the caller can add the PID
	// to the eBPF tracked set before the target binary performs its exec.
	wrapperArgs := []string{
		"-c",
		"kill -STOP $$; exec \"$0\" \"$@\"",
		path,
	}
	wrapperArgs = append(wrapperArgs, args[1:]...)

	cmd := exec.Command("/bin/sh", wrapperArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create a new process group so we can track all children.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return &Launcher{
		cmd:        cmd,
		done:       make(chan struct{}),
		commandArg: append([]string(nil), args...),
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

	pid := uint32(l.cmd.Process.Pid)
	if err := waitForStop(pid, 2*time.Second); err != nil {
		return 0, err
	}

	return pid, nil
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

// Continue resumes a launcher started in the suspended pre-exec state.
func (l *Launcher) Continue() error {
	return l.Signal(syscall.SIGCONT)
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
	return append([]string(nil), l.commandArg...)
}

// CommandString returns the command line as a single string for display.
func (l *Launcher) CommandString() string {
	return strings.Join(l.commandArg, " ")
}

func waitForStop(pid uint32, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state, err := processState(pid)
		if err == nil && state == 'T' {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("process %d did not enter the suspended pre-exec state", pid)
}

func processState(pid uint32) (byte, error) {
	data, err := os.ReadFile("/proc/" + strconv.FormatUint(uint64(pid), 10) + "/status")
	if err != nil {
		return 0, err
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "State:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				break
			}
			return fields[1][0], nil
		}
	}

	return 0, fmt.Errorf("could not determine process %d state", pid)
}
