package output

import (
	"strings"
	"testing"
	"time"

	"github.com/Mutasem-mk4/procscope/internal/events"
)

func TestTimelineRenderEvent(t *testing.T) {
	tl := NewTimeline(false) // no color

	evt := &events.Event{
		Timestamp: time.Now(),
		Type:      events.EventExec,
		PID:       1234,
		Comm:      "test-binary",
		Process:   &events.ProcessData{Filename: "/usr/bin/test-binary"},
	}

	line := tl.RenderEvent(evt)

	if !strings.Contains(line, "1234") {
		t.Error("timeline should contain PID")
	}
	if !strings.Contains(line, "test-binary") {
		t.Error("timeline should contain comm")
	}
	if !strings.Contains(line, "process.exec") {
		t.Error("timeline should contain event type")
	}
	if !strings.Contains(line, "/usr/bin/test-binary") {
		t.Error("timeline should contain filename")
	}
}

func TestTimelineHeader(t *testing.T) {
	tl := NewTimeline(false)
	hdr := tl.Header()

	for _, field := range []string{"TIME", "PID", "COMM", "EVENT", "DETAILS"} {
		if !strings.Contains(hdr, field) {
			t.Errorf("header missing field: %s", field)
		}
	}
}

func TestTimelineNotableMarker(t *testing.T) {
	tl := NewTimeline(false)

	evt := &events.Event{
		Timestamp: time.Now(),
		Type:      events.EventPrivSetUID,
		PID:       100,
		Comm:      "su",
		Privilege: &events.PrivilegeData{Operation: "setuid", OldUID: 1000, NewUID: 0},
	}

	line := tl.RenderEvent(evt)
	if !strings.Contains(line, "!") {
		t.Error("notable event should have ! marker")
	}
}

func TestTimelineNetworkDetail(t *testing.T) {
	tl := NewTimeline(false)

	evt := &events.Event{
		Timestamp: time.Now(),
		Type:      events.EventNetConnect,
		PID:       100,
		Comm:      "curl",
		Network: &events.NetworkData{
			Family:  "ipv4",
			DstAddr: "93.184.216.34",
			DstPort: 443,
		},
	}

	line := tl.RenderEvent(evt)
	if !strings.Contains(line, "93.184.216.34") {
		t.Error("timeline should contain destination address")
	}
	if !strings.Contains(line, "443") {
		t.Error("timeline should contain destination port")
	}
}

func TestTimelineCount(t *testing.T) {
	tl := NewTimeline(false)
	if tl.Count() != 0 {
		t.Error("fresh timeline should have count 0")
	}

	for i := 0; i < 5; i++ {
		tl.RenderEvent(&events.Event{
			Timestamp: time.Now(),
			Type:      events.EventExec,
			PID:       100,
			Comm:      "test",
			Process:   &events.ProcessData{Filename: "/bin/test"},
		})
	}

	if tl.Count() != 5 {
		t.Errorf("count = %d, want 5", tl.Count())
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{50 * time.Millisecond, "50ms"},
		{500 * time.Millisecond, "500ms"},
		{1500 * time.Millisecond, "1.5s"},
		{30 * time.Second, "30.0s"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %s, want %s", tt.d, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is too long for the limit", 10, "this is..."},
		{"ab", 3, "ab"},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
	}

	for _, tt := range tests {
		got := truncate(tt.s, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
		}
	}
}
