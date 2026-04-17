//go:build linux

package tracer

type procscopeEvent struct {
	Timestamp     uint64
	EventType     uint32
	Pid           uint32
	Tid           uint32
	Ppid          uint32
	CgroupId      uint64
	Uid           uint32
	Gid           uint32
	Comm          [16]byte
	ExitCode      uint32
	ChildPid      uint32
	Filename      [256]byte
	Flags         uint32
	Mode          uint32
	Path          [256]byte
	Path2         [256]byte
	Af            uint32
	Sport         uint16
	Dport         uint16
	Saddr         [16]byte
	Daddr         [16]byte
	Protocol      uint32
	Backlog       uint32
	OldUid        uint32
	NewUid        uint32
	OldGid        uint32
	NewGid        uint32
	PtraceRequest uint64
	TargetPid     uint32
	NsType        uint32
	CloneFlags    uint64
	Fstype        [64]byte
	MountFlags    uint64
	Retval        int32
	Pad           uint32
}
