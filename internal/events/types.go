// Package events defines the procscope event schema for runtime observations.
//
// Schema version: 1.0.0
//
// All events share a common envelope (Event) with type-specific data structs.
// The schema is explicitly versioned to support forward-compatible consumers.
//
// Visibility guarantees:
//   - Process lifecycle (exec/fork/exit): exact when eBPF probes are active.
//   - File activity: best-effort — not all file I/O paths are traced.
//   - Network activity: best-effort — covers common socket syscalls.
//   - DNS: best-effort — only userspace UDP DNS to port 53 is partially visible.
//   - Privilege transitions: best-effort — covers setuid/setgid/ptrace syscalls.
//   - Namespace changes: best-effort — covers setns/unshare syscalls.
//   - Mount operations: best-effort — covers mount syscall.
package events

import (
	"encoding/json"
	"time"
)

// SchemaVersion is the current event schema version.
// Consumers should check this field for forward compatibility.
const SchemaVersion = "1.0.0"

// EventType classifies the observed activity.
type EventType string

// Process lifecycle events.
const (
	EventExec EventType = "process.exec"
	EventFork EventType = "process.fork"
	EventExit EventType = "process.exit"
)

// File activity events.
const (
	EventFileOpen   EventType = "file.open"
	EventFileCreate EventType = "file.create"
	EventFileRename EventType = "file.rename"
	EventFileUnlink EventType = "file.unlink"
	EventFileChmod  EventType = "file.chmod"
	EventFileChown  EventType = "file.chown"
)

// Network activity events.
const (
	EventNetConnect EventType = "net.connect"
	EventNetAccept  EventType = "net.accept"
	EventNetBind    EventType = "net.bind"
	EventNetListen  EventType = "net.listen"
)

// DNS events (best-effort).
const (
	EventDNSQuery EventType = "dns.query"
)

// Privilege-relevant events.
const (
	EventPrivSetUID  EventType = "priv.setuid"
	EventPrivSetGID  EventType = "priv.setgid"
	EventPrivPtrace  EventType = "priv.ptrace"
)

// Namespace-relevant events.
const (
	EventNSSetns   EventType = "ns.setns"
	EventNSUnshare EventType = "ns.unshare"
)

// Mount events.
const (
	EventMount EventType = "mount.mount"
)

// Confidence indicates the reliability of a specific observation.
type Confidence string

const (
	ConfidenceExact      Confidence = "exact"
	ConfidenceBestEffort Confidence = "best-effort"
	ConfidenceInferred   Confidence = "inferred"
)

// Event is the top-level event envelope.
// Exactly one of the type-specific fields (Process, File, Network, DNS,
// Privilege, Namespace, Mount) will be populated, corresponding to EventType.
type Event struct {
	// Schema metadata
	SchemaVersion string `json:"schema_version"`

	// Investigation context
	InvestigationID string `json:"investigation_id"`

	// Timing
	Timestamp  time.Time `json:"timestamp"`
	MonoNanos  uint64    `json:"mono_nanos"` // monotonic clock for ordering

	// Event classification
	Type       EventType  `json:"type"`
	Confidence Confidence `json:"confidence"`

	// Process identity
	PID  uint32 `json:"pid"`
	TID  uint32 `json:"tid"`
	PPID uint32 `json:"ppid"`
	Comm string `json:"comm"`
	UID  uint32 `json:"uid"`
	GID  uint32 `json:"gid"`

	// Container context (best-effort, empty if not in container)
	CgroupID    uint64 `json:"cgroup_id,omitempty"`
	ContainerID string `json:"container_id,omitempty"`

	// Type-specific data — exactly one populated per event
	Process   *ProcessData   `json:"process,omitempty"`
	File      *FileData      `json:"file,omitempty"`
	Network   *NetworkData   `json:"network,omitempty"`
	DNS       *DNSData       `json:"dns,omitempty"`
	Privilege *PrivilegeData `json:"privilege,omitempty"`
	Namespace *NamespaceData `json:"namespace,omitempty"`
	Mount     *MountData     `json:"mount,omitempty"`
}

// ProcessData carries process lifecycle details.
type ProcessData struct {
	Filename string   `json:"filename,omitempty"` // exec path
	Args     []string `json:"args,omitempty"`     // argv (bounded)
	ExitCode int32    `json:"exit_code,omitempty"`
	ChildPID uint32   `json:"child_pid,omitempty"` // for fork events
	Signal   int32    `json:"signal,omitempty"`     // if killed by signal
}

// AccessMode describes whether a file operation is read-like or write-like.
type AccessMode string

const (
	AccessRead      AccessMode = "read"
	AccessWrite     AccessMode = "write"
	AccessReadWrite AccessMode = "read-write"
	AccessUnknown   AccessMode = "unknown"
)

// FileData carries file activity details.
type FileData struct {
	Path       string     `json:"path"`
	NewPath    string     `json:"new_path,omitempty"` // for rename
	Flags      uint32     `json:"flags,omitempty"`
	Mode       uint32     `json:"mode,omitempty"`
	AccessMode AccessMode `json:"access_mode,omitempty"`
	ReturnCode int32      `json:"return_code,omitempty"`
}

// NetworkData carries network activity details.
type NetworkData struct {
	Family   string `json:"family"`            // "ipv4", "ipv6", "unix", "other"
	Protocol string `json:"protocol"`          // "tcp", "udp", "other"
	SrcAddr  string `json:"src_addr,omitempty"`
	SrcPort  uint16 `json:"src_port,omitempty"`
	DstAddr  string `json:"dst_addr,omitempty"`
	DstPort  uint16 `json:"dst_port,omitempty"`
	Backlog  uint32 `json:"backlog,omitempty"` // for listen
	ReturnCode int32 `json:"return_code,omitempty"`
}

// DNSData carries best-effort DNS query observations.
//
// Limitation: Only UDP DNS queries to port 53 may be partially visible.
// Queries over TLS (DoT/DoH), queries from statically linked resolvers,
// or queries via non-standard paths are NOT captured.
type DNSData struct {
	QueryName string `json:"query_name,omitempty"`
	QueryType string `json:"query_type,omitempty"` // "A", "AAAA", etc.
}

// PrivilegeData carries privilege transition details.
type PrivilegeData struct {
	Operation    string `json:"operation"`           // "setuid", "setgid", "ptrace"
	OldUID       uint32 `json:"old_uid,omitempty"`
	NewUID       uint32 `json:"new_uid,omitempty"`
	OldGID       uint32 `json:"old_gid,omitempty"`
	NewGID       uint32 `json:"new_gid,omitempty"`
	TargetPID    uint32 `json:"target_pid,omitempty"`  // ptrace target
	PtraceReq    uint64 `json:"ptrace_request,omitempty"`
	ReturnCode   int32  `json:"return_code,omitempty"`
}

// NamespaceData carries namespace change details.
type NamespaceData struct {
	Operation  string `json:"operation"`            // "setns", "unshare"
	NSType     uint32 `json:"ns_type,omitempty"`    // CLONE_NEW* flags
	CloneFlags uint64 `json:"clone_flags,omitempty"`
	ReturnCode int32  `json:"return_code,omitempty"`
}

// MountData carries mount operation details.
type MountData struct {
	Source     string `json:"source,omitempty"`
	Target     string `json:"target,omitempty"`
	FSType     string `json:"fs_type,omitempty"`
	MountFlags uint64 `json:"mount_flags,omitempty"`
	ReturnCode int32  `json:"return_code,omitempty"`
}

// MarshalJSON implements custom JSON marshaling with schema version injection.
func (e *Event) MarshalJSON() ([]byte, error) {
	e.SchemaVersion = SchemaVersion
	type Alias Event
	return json.Marshal((*Alias)(e))
}

// CategoryString returns a human-readable category for the event type.
func (t EventType) CategoryString() string {
	switch {
	case t == EventExec || t == EventFork || t == EventExit:
		return "process"
	case t == EventFileOpen || t == EventFileCreate ||
		t == EventFileRename || t == EventFileUnlink ||
		t == EventFileChmod || t == EventFileChown:
		return "file"
	case t == EventNetConnect || t == EventNetAccept ||
		t == EventNetBind || t == EventNetListen:
		return "network"
	case t == EventDNSQuery:
		return "dns"
	case t == EventPrivSetUID || t == EventPrivSetGID || t == EventPrivPtrace:
		return "privilege"
	case t == EventNSSetns || t == EventNSUnshare:
		return "namespace"
	case t == EventMount:
		return "mount"
	default:
		return "unknown"
	}
}

// IsNotable returns true if this event type is typically significant in
// security investigations (privilege changes, namespace ops, etc.).
func (t EventType) IsNotable() bool {
	switch t {
	case EventPrivSetUID, EventPrivSetGID, EventPrivPtrace,
		EventNSSetns, EventNSUnshare, EventMount:
		return true
	default:
		return false
	}
}
