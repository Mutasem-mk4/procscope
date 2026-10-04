//go:build linux

package process

import (
	"bytes"
	"testing"
)

func TestLauncherStartSuspendedAndContinue(t *testing.T) {
	var output bytes.Buffer
	launcher, err := NewLauncher([]string{"sh", "-c", "printf target-output; exit 7"}, &output)
	if err != nil {
		t.Fatalf("NewLauncher() error = %v", err)
	}

	pid, err := launcher.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if pid == 0 {
		t.Fatal("Start() returned PID 0")
	}

	state, err := processState(pid)
	if err != nil {
		t.Fatalf("processState() error = %v", err)
	}
	if state != 'T' {
		t.Fatalf("state = %q, want stopped ('T')", state)
	}

	if err := launcher.Continue(); err != nil {
		t.Fatalf("Continue() error = %v", err)
	}
	if err := launcher.Wait(); err == nil {
		t.Fatal("Wait() error = nil, want non-nil exit status for code 7")
	}
	if exitCode := launcher.ExitCode(); exitCode != 7 {
		t.Fatalf("ExitCode() = %d, want 7", exitCode)
	}
	if output.String() != "target-output" {
		t.Fatalf("target output = %q, want target-output", output.String())
	}
}
