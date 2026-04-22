package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Mutasem-mk4/procscope/internal/events"
)

// Bundle creates an evidence bundle directory containing:
//   - metadata.json    — investigation metadata
//   - events.jsonl     — complete event stream
//   - process-tree.txt — human-readable process tree
//   - files.json       — file activity summary
//   - network.json     — network activity summary
//   - notable.json     — notably security-relevant events
//   - summary.md       — Markdown executive summary
type Bundle struct {
	Dir         string
	Correlator  *events.Correlator
	Events      []*events.Event
	StartTime   time.Time
	EndTime     time.Time
	CommandLine string
	TargetPID   uint32
}

// BundleMetadata is written to metadata.json.
type BundleMetadata struct {
	SchemaVersion   string    `json:"schema_version"`
	InvestigationID string    `json:"investigation_id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Duration        string    `json:"duration"`
	CommandLine     string    `json:"command_line,omitempty"`
	TargetPID       uint32    `json:"target_pid"`
	TotalEvents     int       `json:"total_events"`
	ProcessCount    int       `json:"process_count"`
	EventCounts     map[string]uint64 `json:"event_counts"`
}

// Write creates the evidence bundle directory and writes all files.
func (b *Bundle) Write() error {
	if err := os.MkdirAll(b.Dir, 0750); err != nil {
		return fmt.Errorf("failed to create bundle directory: %w", err)
	}

	// 1. metadata.json
	if err := b.writeMetadata(); err != nil {
		return fmt.Errorf("metadata: %w", err)
	}

	// 2. events.jsonl
	if err := b.writeEventsJSONL(); err != nil {
		return fmt.Errorf("events: %w", err)
	}

	// 3. process-tree.txt
	if err := b.writeProcessTree(); err != nil {
		return fmt.Errorf("process tree: %w", err)
	}

	// 4. files.json
	if err := b.writeFileSummary(); err != nil {
		return fmt.Errorf("file summary: %w", err)
	}

	// 5. network.json
	if err := b.writeNetworkSummary(); err != nil {
		return fmt.Errorf("network summary: %w", err)
	}

	// 6. notable.json
	if err := b.writeNotableEvents(); err != nil {
		return fmt.Errorf("notable events: %w", err)
	}

	// 7. summary.md
	if err := b.writeSummary(); err != nil {
		return fmt.Errorf("summary: %w", err)
	}

	return nil
}

func (b *Bundle) writeMetadata() error {
	stats := b.Correlator.Stats()
	eventCounts := make(map[string]uint64)
	for k, v := range stats {
		eventCounts[string(k)] = v
	}

	meta := BundleMetadata{
		SchemaVersion:   events.SchemaVersion,
		InvestigationID: b.Correlator.InvestigationID(),
		StartTime:       b.StartTime,
		EndTime:         b.EndTime,
		Duration:        b.EndTime.Sub(b.StartTime).String(),
		CommandLine:     b.CommandLine,
		TargetPID:       b.TargetPID,
		TotalEvents:     len(b.Events),
		ProcessCount:    len(b.Correlator.ProcessTree()),
		EventCounts:     eventCounts,
	}

	return writeJSON(filepath.Join(b.Dir, "metadata.json"), meta)
}

func (b *Bundle) writeEventsJSONL() error {
	f, err := os.Create(filepath.Join(b.Dir, "events.jsonl"))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	for _, evt := range b.Events {
		data, err := json.Marshal(evt)
		if err != nil {
			continue
		}
		_, _ = _, _ = f.Write(data)
		_, _ = _, _ = f.Write([]byte("\n"))
	}
	return nil
}

func (b *Bundle) writeProcessTree() error {
	procs := b.Correlator.ProcessTree()

	f, err := os.Create(filepath.Join(b.Dir, "process-tree.txt"))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, _ = _, _ = fmt.Fprintf(f, "Process Tree — Investigation %s\n", b.Correlator.InvestigationID())
	_, _ = _, _ = fmt.Fprintf(f, "Root PID: %d\n", b.TargetPID)
	_, _ = _, _ = fmt.Fprintf(f, "═══════════════════════════════════════════════════════\n\n")

	// Sort by PID for deterministic output
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].PID < procs[j].PID
	})

	for _, p := range procs {
		indent := ""
		for i := 0; i < p.Depth; i++ {
			indent += "  "
		}
		status := "active"
		if p.Exited {
			status = fmt.Sprintf("exited(%d)", p.ExitCode)
		}
		_, _ = _, _ = fmt.Fprintf(f, "%s[%d] %s — ppid=%d %s\n", indent, p.PID, p.Comm, p.PPID, status)
		if len(p.Args) > 0 {
			_, _ = _, _ = fmt.Fprintf(f, "%s  args: %v\n", indent, p.Args)
		}
	}
	return nil
}

func (b *Bundle) writeFileSummary() error {
	type fileSummary struct {
		Path       string   `json:"path"`
		Operations []string `json:"operations"`
		AccessMode string   `json:"access_mode,omitempty"`
	}

	fileMap := make(map[string]*fileSummary)
	for _, evt := range b.Events {
		if evt.File == nil {
			continue
		}
		key := evt.File.Path
		if _, ok := fileMap[key]; !ok {
			fileMap[key] = &fileSummary{
				Path:       evt.File.Path,
				AccessMode: string(evt.File.AccessMode),
			}
		}
		fileMap[key].Operations = append(fileMap[key].Operations, string(evt.Type))
	}

	files := make([]*fileSummary, 0, len(fileMap))
	for _, f := range fileMap {
		files = append(files, f)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	return writeJSON(filepath.Join(b.Dir, "files.json"), files)
}

func (b *Bundle) writeNetworkSummary() error {
	type netSummary struct {
		Event    string `json:"event"`
		Family   string `json:"family"`
		Protocol string `json:"protocol"`
		SrcAddr  string `json:"src_addr,omitempty"`
		SrcPort  uint16 `json:"src_port,omitempty"`
		DstAddr  string `json:"dst_addr,omitempty"`
		DstPort  uint16 `json:"dst_port,omitempty"`
		PID      uint32 `json:"pid"`
		Comm     string `json:"comm"`
	}

	var nets []*netSummary
	for _, evt := range b.Events {
		if evt.Network == nil {
			continue
		}
		nets = append(nets, &netSummary{
			Event:    string(evt.Type),
			Family:   evt.Network.Family,
			Protocol: evt.Network.Protocol,
			SrcAddr:  evt.Network.SrcAddr,
			SrcPort:  evt.Network.SrcPort,
			DstAddr:  evt.Network.DstAddr,
			DstPort:  evt.Network.DstPort,
			PID:      evt.PID,
			Comm:     evt.Comm,
		})
	}

	return writeJSON(filepath.Join(b.Dir, "network.json"), nets)
}

func (b *Bundle) writeNotableEvents() error {
	var notable []*events.Event
	for _, evt := range b.Events {
		if evt.Type.IsNotable() {
			notable = append(notable, evt)
		}
	}
	return writeJSON(filepath.Join(b.Dir, "notable.json"), notable)
}

func (b *Bundle) writeSummary() error {
	summaryWriter := NewSummaryWriter(b.Correlator, b.Events, b.CommandLine, b.TargetPID, b.StartTime, b.EndTime)
	return summaryWriter.WriteToFile(filepath.Join(b.Dir, "summary.md"))
}

func writeJSON(path string, v interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
