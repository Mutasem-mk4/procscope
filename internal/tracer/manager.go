//go:build linux

package tracer

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"

	"github.com/Mutasem-mk4/procscope/internal/events"
)

// eBPF event type enum values — must match bpf/procscope.c
const (
	bpfEventExec       = 1
	bpfEventFork       = 2
	bpfEventExit       = 3
	bpfEventFileOpen   = 10
	bpfEventFileRename = 12
	bpfEventFileUnlink = 13
	bpfEventFileChmod  = 14
	bpfEventFileChown  = 15
	bpfEventNetConnect = 20
	bpfEventNetAccept  = 21
	bpfEventNetBind    = 22
	bpfEventNetListen  = 23
	bpfEventPrivSetUID = 30
	bpfEventPrivSetGID = 31
	bpfEventPrivPtrace = 32
	bpfEventNSSetns    = 40
	bpfEventNSUnshare  = 41
	bpfEventMount      = 50
)

// Address families
const (
	afINET  = 2
	afINET6 = 10
)

// Manager loads eBPF programs, attaches probes, and reads events from the ring buffer.
type Manager struct {
	objs       procscopeObjects
	links      []link.Link
	reader     *ringbuf.Reader
	correlator *events.Correlator
	bootTime   time.Time
}

// NewManager creates a new eBPF tracer manager.
func NewManager(correlator *events.Correlator) *Manager {
	return &Manager{
		correlator: correlator,
		bootTime:   estimateBootTime(),
	}
}

// Load loads the eBPF programs and maps into the kernel.
func (m *Manager) Load() error {
	spec, err := loadProcscope()
	if err != nil {
		return fmt.Errorf("failed to load eBPF spec: %w\n\nThis usually means:\n"+
			"  - Kernel < 5.8 (ring buffer not supported)\n"+
			"  - BTF not available (/sys/kernel/btf/vmlinux missing)\n"+
			"  - Insufficient privileges (need CAP_BPF + CAP_PERFMON, or root)\n"+
			"  - Embedded eBPF object missing or stale (run 'make generate' after editing bpf/procscope.c)", err)
	}

	if err := spec.LoadAndAssign(&m.objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "", // no pinning
		},
	}); err != nil {
		return fmt.Errorf("failed to load eBPF objects: %w", err)
	}

	return nil
}

// TrackPID adds a PID to the eBPF tracked_pids map.
func (m *Manager) TrackPID(pid uint32) error {
	val := uint8(1)
	if err := m.objs.TrackedPids.Put(pid, val); err != nil {
		return fmt.Errorf("failed to track PID %d: %w", pid, err)
	}
	return nil
}

// UntrackPID removes a PID from the eBPF tracked_pids map.
func (m *Manager) UntrackPID(pid uint32) error {
	if err := m.objs.TrackedPids.Delete(pid); err != nil {
		return fmt.Errorf("failed to untrack PID %d: %w", pid, err)
	}
	return nil
}

// Attach attaches all eBPF programs to their tracepoints.
func (m *Manager) Attach() error {
	type probe struct {
		group string
		name  string
		prog  *ebpf.Program
	}

	probes := []probe{
		// Process lifecycle
		{"sched", "sched_process_exec", m.objs.HandleExec},
		{"sched", "sched_process_fork", m.objs.HandleFork},
		{"sched", "sched_process_exit", m.objs.HandleExit},
		// File activity
		{"syscalls", "sys_enter_openat", m.objs.HandleOpenat},
		{"syscalls", "sys_enter_renameat2", m.objs.HandleRename},
		{"syscalls", "sys_enter_unlinkat", m.objs.HandleUnlink},
		{"syscalls", "sys_enter_fchmodat", m.objs.HandleChmod},
		{"syscalls", "sys_enter_fchownat", m.objs.HandleChown},
		// Network activity
		{"syscalls", "sys_enter_connect", m.objs.HandleConnect},
		{"syscalls", "sys_enter_accept4", m.objs.HandleAccept},
		{"syscalls", "sys_enter_bind", m.objs.HandleBind},
		{"syscalls", "sys_enter_listen", m.objs.HandleListen},
		// Privilege
		{"syscalls", "sys_enter_setuid", m.objs.HandleSetuid},
		{"syscalls", "sys_enter_setgid", m.objs.HandleSetgid},
		{"syscalls", "sys_enter_ptrace", m.objs.HandlePtrace},
		// Namespace
		{"syscalls", "sys_enter_setns", m.objs.HandleSetns},
		{"syscalls", "sys_enter_unshare", m.objs.HandleUnshare},
		// Mount
		{"syscalls", "sys_enter_mount", m.objs.HandleMount},
	}

	for _, p := range probes {
		if p.prog == nil {
			continue // program may not exist in all builds
		}
		tp, err := link.Tracepoint(p.group, p.name, p.prog, nil)
		if err != nil {
			// Non-fatal: log and continue. Some tracepoints may not be
			// available on all kernels.
			_, _ = _, _ = fmt.Fprintf(os.Stderr, "  warning: tracepoint %s/%s: %v (skipping)\n", p.group, p.name, err)
			continue
		}
		m.links = append(m.links, tp)
	}

	if len(m.links) == 0 {
		return fmt.Errorf("no tracepoints could be attached — check kernel version and privileges")
	}

	return nil
}

// ReadEvents reads events from the ring buffer and submits them to the correlator.
// Blocks until ctx is cancelled.
func (m *Manager) ReadEvents(ctx context.Context) error {
	rd, err := ringbuf.NewReader(m.objs.Events)
	if err != nil {
		return fmt.Errorf("failed to create ring buffer reader: %w", err)
	}
	m.reader = rd

	// Close reader when context is done (interrupts blocking read)
	go func() {
		<-ctx.Done()
		rd.Close()
	}()

	for {
		record, err := rd.Read()
		if err != nil {
			if ctx.Err() != nil {
				return nil // normal shutdown
			}
			return fmt.Errorf("ring buffer read error: %w", err)
		}

		evt, err := m.parseEvent(record.RawSample)
		if err != nil {
			continue // skip malformed events
		}

		m.correlator.Submit(evt)
	}
}

// Close releases all eBPF resources.
func (m *Manager) Close() {
	if m.reader != nil {
		_ = m.reader.Close()
	}
	for _, l := range m.links {
		_ = l.Close()
	}
	m.objs.Close()
}

// parseEvent converts a raw eBPF ring buffer record to a procscope Event.
func (m *Manager) parseEvent(raw []byte) (*events.Event, error) {
	if len(raw) < int(unsafe.Sizeof(procscopeEvent{})) {
		return nil, fmt.Errorf("event too small: %d bytes", len(raw))
	}

	var bpfEvt procscopeEvent
	if err := binary.Read(bytes.NewReader(raw), binary.LittleEndian, &bpfEvt); err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	evt := &events.Event{
		Timestamp: m.bootTime.Add(time.Duration(bpfEvt.Timestamp)),
		MonoNanos: bpfEvt.Timestamp,
		PID:       bpfEvt.Pid,
		TID:       bpfEvt.Tid,
		PPID:      bpfEvt.Ppid,
		UID:       bpfEvt.Uid,
		GID:       bpfEvt.Gid,
		Comm:      nullTermString(bpfEvt.Comm[:]),
		CgroupID:  bpfEvt.CgroupId,
	}

	switch bpfEvt.EventType {
	case bpfEventExec:
		evt.Type = events.EventExec
		evt.Confidence = events.ConfidenceExact
		evt.Process = &events.ProcessData{
			Filename: nullTermString(bpfEvt.Filename[:]),
		}
	case bpfEventFork:
		evt.Type = events.EventFork
		evt.Confidence = events.ConfidenceExact
		evt.Process = &events.ProcessData{
			ChildPID: bpfEvt.ChildPid,
		}
	case bpfEventExit:
		evt.Type = events.EventExit
		evt.Confidence = events.ConfidenceExact
		evt.Process = &events.ProcessData{
			ExitCode: int32(bpfEvt.ExitCode),
		}
	case bpfEventFileOpen:
		evt.Type = events.EventFileOpen
		evt.Confidence = events.ConfidenceBestEffort
		evt.File = &events.FileData{
			Path:       nullTermString(bpfEvt.Path[:]),
			Flags:      bpfEvt.Flags,
			Mode:       bpfEvt.Mode,
			AccessMode: classifyFileAccess(bpfEvt.Flags),
		}
	case bpfEventFileRename:
		evt.Type = events.EventFileRename
		evt.Confidence = events.ConfidenceBestEffort
		evt.File = &events.FileData{
			Path:       nullTermString(bpfEvt.Path[:]),
			NewPath:    nullTermString(bpfEvt.Path2[:]),
			AccessMode: events.AccessWrite,
		}
	case bpfEventFileUnlink:
		evt.Type = events.EventFileUnlink
		evt.Confidence = events.ConfidenceBestEffort
		evt.File = &events.FileData{
			Path:       nullTermString(bpfEvt.Path[:]),
			AccessMode: events.AccessWrite,
		}
	case bpfEventFileChmod:
		evt.Type = events.EventFileChmod
		evt.Confidence = events.ConfidenceBestEffort
		evt.File = &events.FileData{
			Path: nullTermString(bpfEvt.Path[:]),
			Mode: bpfEvt.Mode,
		}
	case bpfEventFileChown:
		evt.Type = events.EventFileChown
		evt.Confidence = events.ConfidenceBestEffort
		evt.File = &events.FileData{
			Path: nullTermString(bpfEvt.Path[:]),
		}
		evt.Privilege = &events.PrivilegeData{
			Operation: "chown",
			NewUID:    bpfEvt.NewUid,
			NewGID:    bpfEvt.NewGid,
		}
	case bpfEventNetConnect:
		evt.Type = events.EventNetConnect
		evt.Confidence = events.ConfidenceBestEffort
		evt.Network = parseNetEvent(&bpfEvt, false)
	case bpfEventNetAccept:
		evt.Type = events.EventNetAccept
		evt.Confidence = events.ConfidenceBestEffort
		evt.Network = &events.NetworkData{Protocol: "unknown"}
	case bpfEventNetBind:
		evt.Type = events.EventNetBind
		evt.Confidence = events.ConfidenceBestEffort
		evt.Network = parseNetEvent(&bpfEvt, true)
	case bpfEventNetListen:
		evt.Type = events.EventNetListen
		evt.Confidence = events.ConfidenceBestEffort
		evt.Network = &events.NetworkData{
			Backlog: bpfEvt.Backlog,
		}
	case bpfEventPrivSetUID:
		evt.Type = events.EventPrivSetUID
		evt.Confidence = events.ConfidenceExact
		evt.Privilege = &events.PrivilegeData{
			Operation: "setuid",
			OldUID:    bpfEvt.OldUid,
			NewUID:    bpfEvt.NewUid,
		}
	case bpfEventPrivSetGID:
		evt.Type = events.EventPrivSetGID
		evt.Confidence = events.ConfidenceExact
		evt.Privilege = &events.PrivilegeData{
			Operation: "setgid",
			OldGID:    bpfEvt.OldGid,
			NewGID:    bpfEvt.NewGid,
		}
	case bpfEventPrivPtrace:
		evt.Type = events.EventPrivPtrace
		evt.Confidence = events.ConfidenceBestEffort
		evt.Privilege = &events.PrivilegeData{
			Operation: "ptrace",
			TargetPID: bpfEvt.TargetPid,
			PtraceReq: bpfEvt.PtraceRequest,
		}
	case bpfEventNSSetns:
		evt.Type = events.EventNSSetns
		evt.Confidence = events.ConfidenceBestEffort
		evt.Namespace = &events.NamespaceData{
			Operation: "setns",
			NSType:    bpfEvt.NsType,
		}
	case bpfEventNSUnshare:
		evt.Type = events.EventNSUnshare
		evt.Confidence = events.ConfidenceBestEffort
		evt.Namespace = &events.NamespaceData{
			Operation:  "unshare",
			CloneFlags: bpfEvt.CloneFlags,
		}
	case bpfEventMount:
		evt.Type = events.EventMount
		evt.Confidence = events.ConfidenceBestEffort
		evt.Mount = &events.MountData{
			Source:     nullTermString(bpfEvt.Path[:]),
			Target:     nullTermString(bpfEvt.Path2[:]),
			FSType:     nullTermString(bpfEvt.Fstype[:]),
			MountFlags: bpfEvt.MountFlags,
		}
	default:
		return nil, fmt.Errorf("unknown event type: %d", bpfEvt.EventType)
	}

	return evt, nil
}

// parseNetEvent extracts network address info from a BPF event.
func parseNetEvent(bpfEvt *procscopeEvent, isLocal bool) *events.NetworkData {
	nd := &events.NetworkData{}

	switch bpfEvt.Af {
	case afINET:
		nd.Family = "ipv4"
		addr := net.IP(bpfEvt.Daddr[:4])
		port := bpfEvt.Dport
		if isLocal {
			addr = net.IP(bpfEvt.Saddr[:4])
			port = bpfEvt.Sport
		}
		if isLocal {
			nd.SrcAddr = addr.String()
			nd.SrcPort = port
		} else {
			nd.DstAddr = addr.String()
			nd.DstPort = port
		}
	case afINET6:
		nd.Family = "ipv6"
		addr := net.IP(bpfEvt.Daddr[:16])
		port := bpfEvt.Dport
		if isLocal {
			addr = net.IP(bpfEvt.Saddr[:16])
			port = bpfEvt.Sport
		}
		if isLocal {
			nd.SrcAddr = addr.String()
			nd.SrcPort = port
		} else {
			nd.DstAddr = addr.String()
			nd.DstPort = port
		}
	default:
		nd.Family = "other"
	}

	nd.Protocol = "tcp" // best guess; socket type not available at syscall enter
	return nd
}

// O_WRONLY and O_RDWR flags for classifying file access.
const (
	oRDONLY = 0x0
	oWRONLY = 0x1
	oRDWR   = 0x2
	oCREAT  = 0x40
	oTRUNC  = 0x200
	oAPPEND = 0x400
)

// classifyFileAccess determines if a file open is read-like or write-like.
func classifyFileAccess(flags uint32) events.AccessMode {
	accessBits := flags & 0x3
	switch accessBits {
	case oRDONLY:
		if flags&(oCREAT|oTRUNC) != 0 {
			return events.AccessWrite
		}
		return events.AccessRead
	case oWRONLY:
		return events.AccessWrite
	case oRDWR:
		return events.AccessReadWrite
	default:
		return events.AccessUnknown
	}
}

// nullTermString converts a null-terminated byte slice to a Go string.
func nullTermString(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}

// estimateBootTime approximates kernel boot time for converting ktime_get_ns
// to wall clock time. This is best-effort.
func estimateBootTime() time.Time {
	now := time.Now()
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return now // fallback
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return now
	}

	var uptime float64
	if _, err := fmt.Sscanf(fields[0], "%f", &uptime); err != nil {
		return now
	}

	return now.Add(-time.Duration(uptime * float64(time.Second)))
}
