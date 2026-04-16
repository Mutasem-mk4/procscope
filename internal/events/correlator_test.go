package events

import (
	"testing"
	"time"
)

func TestCorrelatorBasic(t *testing.T) {
	c := NewCorrelator("test-001", 100, 64, 4096)
	defer c.Close()

	if c.InvestigationID() != "test-001" {
		t.Errorf("InvestigationID = %s, want test-001", c.InvestigationID())
	}

	if !c.IsTracked(100) {
		t.Error("root PID 100 should be tracked")
	}

	if c.IsTracked(999) {
		t.Error("PID 999 should not be tracked")
	}
}

func TestCorrelatorSubmit(t *testing.T) {
	c := NewCorrelator("test-002", 100, 64, 4096)

	// Submit an event for a tracked PID
	evt := &Event{
		Timestamp: time.Now(),
		Type:      EventExec,
		PID:       100,
		Comm:      "test",
		Process:   &ProcessData{Filename: "/usr/bin/test"},
	}

	accepted := c.Submit(evt)
	if !accepted {
		t.Error("event for tracked PID should be accepted")
	}

	// Verify event is enriched
	select {
	case received := <-c.Events():
		if received.InvestigationID != "test-002" {
			t.Errorf("InvestigationID = %s, want test-002", received.InvestigationID)
		}
		if received.SchemaVersion != SchemaVersion {
			t.Errorf("SchemaVersion = %s, want %s", received.SchemaVersion, SchemaVersion)
		}
	default:
		t.Error("expected event on channel")
	}

	c.Close()
}

func TestCorrelatorRejectsUntracked(t *testing.T) {
	c := NewCorrelator("test-003", 100, 64, 4096)
	defer c.Close()

	evt := &Event{
		Timestamp: time.Now(),
		Type:      EventExec,
		PID:       999, // not tracked
		Comm:      "other",
	}

	accepted := c.Submit(evt)
	if accepted {
		t.Error("event for untracked PID should be rejected")
	}
}

func TestCorrelatorForkAutoTracking(t *testing.T) {
	c := NewCorrelator("test-004", 100, 64, 4096)

	// Fork event from PID 100 creating PID 200
	evt := &Event{
		Timestamp: time.Now(),
		Type:      EventFork,
		PID:       100,
		Comm:      "parent",
		Process:   &ProcessData{ChildPID: 200},
	}

	c.Submit(evt)
	// Drain the event
	<-c.Events()

	// PID 200 should now be tracked
	if !c.IsTracked(200) {
		t.Error("child PID 200 should be auto-tracked after fork")
	}

	// Events from PID 200 should be accepted
	childEvt := &Event{
		Timestamp: time.Now(),
		Type:      EventExec,
		PID:       200,
		Comm:      "child",
		Process:   &ProcessData{Filename: "/usr/bin/child"},
	}

	accepted := c.Submit(childEvt)
	if !accepted {
		t.Error("event from auto-tracked child should be accepted")
	}

	c.Close()
}

func TestCorrelatorExitTracking(t *testing.T) {
	c := NewCorrelator("test-005", 100, 64, 4096)

	// Exec first
	c.Submit(&Event{
		Timestamp: time.Now(),
		Type:      EventExec,
		PID:       100,
		Comm:      "test",
		Process:   &ProcessData{Filename: "/bin/test"},
	})
	<-c.Events()

	// Then exit
	c.Submit(&Event{
		Timestamp: time.Now(),
		Type:      EventExit,
		PID:       100,
		Comm:      "test",
		Process:   &ProcessData{ExitCode: 42},
	})
	<-c.Events()

	root := c.RootProcess()
	if root == nil {
		t.Fatal("root process should exist")
	}
	if !root.Exited {
		t.Error("root process should be marked as exited")
	}
	if root.ExitCode != 42 {
		t.Errorf("exit code = %d, want 42", root.ExitCode)
	}

	c.Close()
}

func TestCorrelatorStats(t *testing.T) {
	c := NewCorrelator("test-006", 100, 64, 4096)

	for i := 0; i < 5; i++ {
		c.Submit(&Event{
			Timestamp: time.Now(),
			Type:      EventFileOpen,
			PID:       100,
			File:      &FileData{Path: "/tmp/test"},
		})
	}
	// Drain events
	for i := 0; i < 5; i++ {
		<-c.Events()
	}

	stats := c.Stats()
	if stats[EventFileOpen] != 5 {
		t.Errorf("file.open count = %d, want 5", stats[EventFileOpen])
	}

	c.Close()
}

func TestCorrelatorTruncatesLongArgs(t *testing.T) {
	c := NewCorrelator("test-007", 100, 3, 10)

	longArg := "a-very-long-argument-that-exceeds-the-limit"
	evt := &Event{
		Timestamp: time.Now(),
		Type:      EventExec,
		PID:       100,
		Comm:      "test",
		Process: &ProcessData{
			Args: []string{"arg1", "arg2", "arg3", "arg4", "arg5"},
		},
	}
	_ = longArg

	c.Submit(evt)
	<-c.Events()

	root := c.RootProcess()
	if root == nil {
		t.Fatal("root process should exist")
	}
	// MaxArgs is 3, so only first 3 should be retained
	if len(root.Args) != 3 {
		t.Errorf("args count = %d, want 3 (maxArgs)", len(root.Args))
	}

	c.Close()
}

func TestCorrelatorProcessTree(t *testing.T) {
	c := NewCorrelator("test-008", 1, 64, 4096)

	// Fork chain: 1 → 2 → 3
	c.Submit(&Event{Timestamp: time.Now(), Type: EventFork, PID: 1, Process: &ProcessData{ChildPID: 2}})
	<-c.Events()

	c.Submit(&Event{Timestamp: time.Now(), Type: EventFork, PID: 2, Process: &ProcessData{ChildPID: 3}})
	<-c.Events()

	tree := c.ProcessTree()
	if len(tree) != 3 {
		t.Errorf("tree size = %d, want 3", len(tree))
	}

	// Verify depths
	depthMap := make(map[uint32]int)
	for _, p := range tree {
		depthMap[p.PID] = p.Depth
	}

	if depthMap[1] != 0 {
		t.Errorf("PID 1 depth = %d, want 0", depthMap[1])
	}
	if depthMap[2] != 1 {
		t.Errorf("PID 2 depth = %d, want 1", depthMap[2])
	}
	if depthMap[3] != 2 {
		t.Errorf("PID 3 depth = %d, want 2", depthMap[3])
	}

	c.Close()
}

func TestCorrelatorNilEvent(t *testing.T) {
	c := NewCorrelator("test-009", 100, 64, 4096)
	defer c.Close()

	if c.Submit(nil) {
		t.Error("nil event should not be accepted")
	}
}
