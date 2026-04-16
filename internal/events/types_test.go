package events

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEventTypeCategoryString(t *testing.T) {
	tests := []struct {
		eventType EventType
		want      string
	}{
		{EventExec, "process"},
		{EventFork, "process"},
		{EventExit, "process"},
		{EventFileOpen, "file"},
		{EventFileCreate, "file"},
		{EventFileRename, "file"},
		{EventFileUnlink, "file"},
		{EventFileChmod, "file"},
		{EventFileChown, "file"},
		{EventNetConnect, "network"},
		{EventNetAccept, "network"},
		{EventNetBind, "network"},
		{EventNetListen, "network"},
		{EventDNSQuery, "dns"},
		{EventPrivSetUID, "privilege"},
		{EventPrivSetGID, "privilege"},
		{EventPrivPtrace, "privilege"},
		{EventNSSetns, "namespace"},
		{EventNSUnshare, "namespace"},
		{EventMount, "mount"},
		{EventType("unknown.thing"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			got := tt.eventType.CategoryString()
			if got != tt.want {
				t.Errorf("CategoryString(%s) = %s, want %s", tt.eventType, got, tt.want)
			}
		})
	}
}

func TestEventTypeIsNotable(t *testing.T) {
	notable := []EventType{EventPrivSetUID, EventPrivSetGID, EventPrivPtrace, EventNSSetns, EventNSUnshare, EventMount}
	notNotable := []EventType{EventExec, EventFork, EventExit, EventFileOpen, EventNetConnect}

	for _, et := range notable {
		if !et.IsNotable() {
			t.Errorf("expected %s to be notable", et)
		}
	}
	for _, et := range notNotable {
		if et.IsNotable() {
			t.Errorf("expected %s to NOT be notable", et)
		}
	}
}

func TestEventMarshalJSON(t *testing.T) {
	evt := &Event{
		InvestigationID: "ps-test1234",
		Timestamp:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Type:            EventExec,
		Confidence:      ConfidenceExact,
		PID:             1234,
		TID:             1234,
		PPID:            1,
		Comm:            "test",
		Process: &ProcessData{
			Filename: "/usr/bin/test",
			Args:     []string{"test", "--flag"},
		},
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Verify schema version is injected
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if sv, ok := parsed["schema_version"].(string); !ok || sv != SchemaVersion {
		t.Errorf("schema_version = %v, want %s", parsed["schema_version"], SchemaVersion)
	}

	if et, ok := parsed["type"].(string); !ok || et != "process.exec" {
		t.Errorf("type = %v, want process.exec", parsed["type"])
	}

	// Verify roundtrip
	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("roundtrip unmarshal failed: %v", err)
	}
	if decoded.PID != 1234 {
		t.Errorf("roundtrip PID = %d, want 1234", decoded.PID)
	}
	if decoded.Process == nil || decoded.Process.Filename != "/usr/bin/test" {
		t.Error("roundtrip process data lost")
	}
}

func TestEventMarshalJSON_OmitsEmptyOptionals(t *testing.T) {
	evt := &Event{
		InvestigationID: "ps-test",
		Timestamp:       time.Now(),
		Type:            EventExec,
		PID:             1,
		Comm:            "test",
		Process:         &ProcessData{Filename: "/bin/test"},
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	jsonStr := string(data)
	// File, Network, DNS, etc. should not appear
	for _, field := range []string{"file", "network", "dns", "privilege", "namespace", "mount"} {
		if contains(jsonStr, `"`+field+`"`) {
			t.Errorf("JSON contains %q field but it should be omitted", field)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
