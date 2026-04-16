package events

import (
	"fmt"
	"sync"
	"time"
)

// Correlator maintains a process tree rooted at the investigation target
// and correlates incoming events to tracked processes.
//
// Thread-safe for concurrent use by multiple eBPF event readers.
type Correlator struct {
	mu sync.RWMutex

	// investigationID is the unique ID for this investigation session.
	investigationID string

	// rootPID is the root process being investigated.
	rootPID uint32

	// tracked maps PIDs to their process info.
	tracked map[uint32]*TrackedProcess

	// events is the channel where correlated events are sent.
	events chan *Event

	// stats tracks event counts by type.
	stats map[EventType]uint64

	// startTime is the investigation start time.
	startTime time.Time

	// maxArgs is the maximum number of argv elements to retain.
	maxArgs int

	// maxPathLen is the maximum path string length to retain.
	maxPathLen int
}

// TrackedProcess records metadata about a tracked process in the tree.
type TrackedProcess struct {
	PID       uint32
	PPID      uint32
	Comm      string
	Args      []string
	StartTime time.Time
	ExitTime  time.Time
	ExitCode  int32
	Exited    bool
	Children  []uint32
	Depth     int // depth in process tree from root
}

// NewCorrelator creates a new event correlator for an investigation.
func NewCorrelator(investigationID string, rootPID uint32, maxArgs, maxPathLen int) *Correlator {
	if maxArgs <= 0 {
		maxArgs = 64
	}
	if maxPathLen <= 0 {
		maxPathLen = 4096
	}
	c := &Correlator{
		investigationID: investigationID,
		rootPID:         rootPID,
		tracked:         make(map[uint32]*TrackedProcess),
		events:          make(chan *Event, 4096),
		stats:           make(map[EventType]uint64),
		startTime:       time.Now(),
		maxArgs:         maxArgs,
		maxPathLen:      maxPathLen,
	}
	// Register root process
	c.tracked[rootPID] = &TrackedProcess{
		PID:       rootPID,
		StartTime: c.startTime,
	}
	return c
}

// Events returns the channel of correlated events for consumers.
func (c *Correlator) Events() <-chan *Event {
	return c.events
}

// IsTracked returns true if the given PID is part of the investigation.
func (c *Correlator) IsTracked(pid uint32) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.tracked[pid]
	return ok
}

// TrackPID adds a PID to the tracked set (e.g., when attaching to existing children).
func (c *Correlator) TrackPID(pid, ppid uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.tracked[pid]; exists {
		return
	}
	depth := 0
	if parent, ok := c.tracked[ppid]; ok {
		depth = parent.Depth + 1
		parent.Children = append(parent.Children, pid)
	}
	c.tracked[pid] = &TrackedProcess{
		PID:       pid,
		PPID:      ppid,
		StartTime: time.Now(),
		Depth:     depth,
	}
}

// Submit processes a raw event: checks if it belongs to a tracked process,
// enriches it with investigation context, and forwards to the events channel.
//
// Returns true if the event was accepted (PID was tracked).
func (c *Correlator) Submit(evt *Event) bool {
	if evt == nil {
		return false
	}

	c.mu.Lock()

	// Check if this PID is tracked
	proc, tracked := c.tracked[evt.PID]
	if !tracked {
		c.mu.Unlock()
		return false
	}

	// Enrich event with investigation context
	evt.InvestigationID = c.investigationID
	evt.SchemaVersion = SchemaVersion

	// Handle fork: auto-track child PIDs
	if evt.Type == EventFork && evt.Process != nil && evt.Process.ChildPID != 0 {
		childPID := evt.Process.ChildPID
		if _, exists := c.tracked[childPID]; !exists {
			c.tracked[childPID] = &TrackedProcess{
				PID:       childPID,
				PPID:      evt.PID,
				StartTime: evt.Timestamp,
				Depth:     proc.Depth + 1,
			}
			proc.Children = append(proc.Children, childPID)
		}
	}

	// Handle exec: update process metadata
	if evt.Type == EventExec && evt.Process != nil {
		proc.Comm = evt.Comm
		proc.Args = c.truncateArgs(evt.Process.Args)
		if evt.Process.Filename != "" {
			// Truncate path if needed
			if len(evt.Process.Filename) > c.maxPathLen {
				evt.Process.Filename = evt.Process.Filename[:c.maxPathLen] + "..."
			}
		}
	}

	// Handle exit: record exit info
	if evt.Type == EventExit {
		proc.Exited = true
		proc.ExitTime = evt.Timestamp
		if evt.Process != nil {
			proc.ExitCode = evt.Process.ExitCode
		}
	}

	// Truncate file paths
	if evt.File != nil {
		if len(evt.File.Path) > c.maxPathLen {
			evt.File.Path = evt.File.Path[:c.maxPathLen] + "..."
		}
		if len(evt.File.NewPath) > c.maxPathLen {
			evt.File.NewPath = evt.File.NewPath[:c.maxPathLen] + "..."
		}
	}

	// Update stats
	c.stats[evt.Type]++

	c.mu.Unlock()

	// Non-blocking send — if consumer is too slow, drop oldest events
	select {
	case c.events <- evt:
	default:
		// Channel full — this is a bounded-loss design choice for safety.
		// In practice, the 4096-event buffer should be sufficient for
		// typical process-scoped investigations.
	}

	return true
}

// Close closes the events channel. Call after all event sources have stopped.
func (c *Correlator) Close() {
	close(c.events)
}

// Stats returns a snapshot of event counts by type.
func (c *Correlator) Stats() map[EventType]uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[EventType]uint64, len(c.stats))
	for k, v := range c.stats {
		result[k] = v
	}
	return result
}

// ProcessTree returns a snapshot of all tracked processes.
func (c *Correlator) ProcessTree() []*TrackedProcess {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*TrackedProcess, 0, len(c.tracked))
	for _, p := range c.tracked {
		cp := *p // copy
		cp.Children = append([]uint32(nil), p.Children...)
		cp.Args = append([]string(nil), p.Args...)
		result = append(result, &cp)
	}
	return result
}

// RootProcess returns the root tracked process.
func (c *Correlator) RootProcess() *TrackedProcess {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if p, ok := c.tracked[c.rootPID]; ok {
		cp := *p
		return &cp
	}
	return nil
}

// Duration returns the elapsed investigation duration.
func (c *Correlator) Duration() time.Duration {
	return time.Since(c.startTime)
}

// InvestigationID returns the investigation ID.
func (c *Correlator) InvestigationID() string {
	return c.investigationID
}

// Summary returns a human-readable summary of the investigation state.
func (c *Correlator) Summary() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := uint64(0)
	for _, v := range c.stats {
		total += v
	}

	active := 0
	exited := 0
	for _, p := range c.tracked {
		if p.Exited {
			exited++
		} else {
			active++
		}
	}

	return fmt.Sprintf(
		"Investigation %s | Duration: %s | Events: %d | Processes: %d active, %d exited",
		c.investigationID,
		c.Duration().Round(time.Millisecond),
		total,
		active,
		exited,
	)
}

// truncateArgs limits the number and length of command-line arguments.
func (c *Correlator) truncateArgs(args []string) []string {
	if len(args) > c.maxArgs {
		args = args[:c.maxArgs]
	}
	for i, arg := range args {
		if len(arg) > c.maxPathLen {
			args[i] = arg[:c.maxPathLen] + "..."
		}
	}
	return args
}
