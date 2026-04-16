// Package output provides event rendering for procscope investigations.
//
// It includes live terminal timeline, JSON/JSONL output, evidence bundles,
// and Markdown summary generation.
package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/procscope/procscope/internal/events"
)

// Timeline renders events as a compact live terminal timeline.
type Timeline struct {
	startTime time.Time
	colorize  bool
	count     uint64
}

// NewTimeline creates a new timeline renderer.
func NewTimeline(colorize bool) *Timeline {
	return &Timeline{
		startTime: time.Now(),
		colorize:  colorize,
	}
}

// RenderEvent renders a single event to a compact timeline line.
// Format: [+0.123s] PID COMM TYPE details
func (t *Timeline) RenderEvent(evt *events.Event) string {
	t.count++
	elapsed := evt.Timestamp.Sub(t.startTime)
	if elapsed < 0 {
		elapsed = 0
	}

	// Color codes (ANSI)
	var typeColor, resetColor string
	if t.colorize {
		resetColor = "\033[0m"
		switch evt.Type.CategoryString() {
		case "process":
			typeColor = "\033[1;36m" // bold cyan
		case "file":
			typeColor = "\033[0;33m" // yellow
		case "network":
			typeColor = "\033[0;35m" // magenta
		case "dns":
			typeColor = "\033[0;34m" // blue
		case "privilege":
			typeColor = "\033[1;31m" // bold red
		case "namespace":
			typeColor = "\033[1;31m" // bold red
		case "mount":
			typeColor = "\033[1;33m" // bold yellow
		default:
			typeColor = "\033[0m"
		}
	}

	// Notable event marker
	marker := " "
	if evt.Type.IsNotable() {
		if t.colorize {
			marker = "\033[1;31m!\033[0m"
		} else {
			marker = "!"
		}
	}

	detail := t.eventDetail(evt)

	return fmt.Sprintf("%s[+%8s]%s %s%-5d %-15s %s%-18s%s %s",
		typeColor, formatDuration(elapsed), resetColor,
		marker,
		evt.PID,
		truncate(evt.Comm, 15),
		typeColor, string(evt.Type), resetColor,
		detail,
	)
}

// Header returns the timeline header line.
func (t *Timeline) Header() string {
	hdr := fmt.Sprintf("%-12s %-6s %-15s %-18s %s",
		"TIME", "PID", "COMM", "EVENT", "DETAILS")
	if t.colorize {
		return "\033[1;37m" + hdr + "\033[0m"
	}
	return hdr
}

// Count returns the number of events rendered.
func (t *Timeline) Count() uint64 {
	return t.count
}

// eventDetail extracts a concise detail string for each event type.
func (t *Timeline) eventDetail(evt *events.Event) string {
	switch {
	case evt.Process != nil:
		return t.processDetail(evt)
	case evt.File != nil:
		return t.fileDetail(evt)
	case evt.Network != nil:
		return t.networkDetail(evt)
	case evt.DNS != nil:
		return t.dnsDetail(evt)
	case evt.Privilege != nil:
		return t.privilegeDetail(evt)
	case evt.Namespace != nil:
		return t.namespaceDetail(evt)
	case evt.Mount != nil:
		return t.mountDetail(evt)
	default:
		return ""
	}
}

func (t *Timeline) processDetail(evt *events.Event) string {
	p := evt.Process
	switch evt.Type {
	case events.EventExec:
		return truncate(p.Filename, 80)
	case events.EventFork:
		return fmt.Sprintf("child_pid=%d", p.ChildPID)
	case events.EventExit:
		if p.Signal != 0 {
			return fmt.Sprintf("exit_code=%d signal=%d", p.ExitCode, p.Signal)
		}
		return fmt.Sprintf("exit_code=%d", p.ExitCode)
	default:
		return ""
	}
}

func (t *Timeline) fileDetail(evt *events.Event) string {
	f := evt.File
	detail := truncate(f.Path, 60)
	if f.NewPath != "" {
		detail += " → " + truncate(f.NewPath, 40)
	}
	if f.AccessMode != "" && f.AccessMode != events.AccessUnknown {
		detail += fmt.Sprintf(" [%s]", f.AccessMode)
	}
	return detail
}

func (t *Timeline) networkDetail(evt *events.Event) string {
	n := evt.Network
	switch evt.Type {
	case events.EventNetConnect:
		if n.DstAddr != "" {
			return fmt.Sprintf("%s → %s:%d", n.Family, n.DstAddr, n.DstPort)
		}
		return n.Family
	case events.EventNetBind:
		if n.SrcAddr != "" {
			return fmt.Sprintf("%s bind %s:%d", n.Family, n.SrcAddr, n.SrcPort)
		}
		return n.Family
	case events.EventNetListen:
		return fmt.Sprintf("backlog=%d", n.Backlog)
	case events.EventNetAccept:
		return n.Family
	default:
		return ""
	}
}

func (t *Timeline) dnsDetail(evt *events.Event) string {
	d := evt.DNS
	if d.QueryName != "" {
		return fmt.Sprintf("%s (%s)", d.QueryName, d.QueryType)
	}
	return "(best-effort)"
}

func (t *Timeline) privilegeDetail(evt *events.Event) string {
	p := evt.Privilege
	switch p.Operation {
	case "setuid":
		return fmt.Sprintf("uid %d → %d", p.OldUID, p.NewUID)
	case "setgid":
		return fmt.Sprintf("gid %d → %d", p.OldGID, p.NewGID)
	case "ptrace":
		return fmt.Sprintf("target_pid=%d req=%d", p.TargetPID, p.PtraceReq)
	case "chown":
		return fmt.Sprintf("uid→%d gid→%d", p.NewUID, p.NewGID)
	default:
		return p.Operation
	}
}

func (t *Timeline) namespaceDetail(evt *events.Event) string {
	n := evt.Namespace
	return fmt.Sprintf("%s flags=0x%x", n.Operation, n.CloneFlags|uint64(n.NSType))
}

func (t *Timeline) mountDetail(evt *events.Event) string {
	m := evt.Mount
	parts := make([]string, 0, 3)
	if m.Source != "" {
		parts = append(parts, m.Source)
	}
	if m.Target != "" {
		parts = append(parts, "→ "+m.Target)
	}
	if m.FSType != "" {
		parts = append(parts, "("+m.FSType+")")
	}
	return strings.Join(parts, " ")
}

// --- Helpers ---

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.0fm%.0fs", d.Minutes(), d.Seconds()-float64(int(d.Minutes()))*60)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
