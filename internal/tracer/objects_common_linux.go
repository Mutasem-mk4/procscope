//go:build linux

package tracer

import "github.com/cilium/ebpf"

type procscopeObjects struct {
	procscopePrograms
	procscopeMaps
}

type procscopePrograms struct {
	HandleExec    *ebpf.Program `ebpf:"handle_exec"`
	HandleFork    *ebpf.Program `ebpf:"handle_fork"`
	HandleExit    *ebpf.Program `ebpf:"handle_exit"`
	HandleOpenat  *ebpf.Program `ebpf:"handle_openat"`
	HandleRename  *ebpf.Program `ebpf:"handle_rename"`
	HandleUnlink  *ebpf.Program `ebpf:"handle_unlink"`
	HandleChmod   *ebpf.Program `ebpf:"handle_chmod"`
	HandleChown   *ebpf.Program `ebpf:"handle_chown"`
	HandleConnect *ebpf.Program `ebpf:"handle_connect"`
	HandleAccept  *ebpf.Program `ebpf:"handle_accept"`
	HandleBind    *ebpf.Program `ebpf:"handle_bind"`
	HandleListen  *ebpf.Program `ebpf:"handle_listen"`
	HandleSetuid  *ebpf.Program `ebpf:"handle_setuid"`
	HandleSetgid  *ebpf.Program `ebpf:"handle_setgid"`
	HandlePtrace  *ebpf.Program `ebpf:"handle_ptrace"`
	HandleSetns   *ebpf.Program `ebpf:"handle_setns"`
	HandleUnshare *ebpf.Program `ebpf:"handle_unshare"`
	HandleMount   *ebpf.Program `ebpf:"handle_mount"`
}

type procscopeMaps struct {
	Events      *ebpf.Map `ebpf:"events"`
	TrackedPids *ebpf.Map `ebpf:"tracked_pids"`
}

func (o *procscopeObjects) Close() {
	o.procscopePrograms.Close()
	o.procscopeMaps.Close()
}

func (p *procscopePrograms) Close() {
	closeProgram(p.HandleExec)
	closeProgram(p.HandleFork)
	closeProgram(p.HandleExit)
	closeProgram(p.HandleOpenat)
	closeProgram(p.HandleRename)
	closeProgram(p.HandleUnlink)
	closeProgram(p.HandleChmod)
	closeProgram(p.HandleChown)
	closeProgram(p.HandleConnect)
	closeProgram(p.HandleAccept)
	closeProgram(p.HandleBind)
	closeProgram(p.HandleListen)
	closeProgram(p.HandleSetuid)
	closeProgram(p.HandleSetgid)
	closeProgram(p.HandlePtrace)
	closeProgram(p.HandleSetns)
	closeProgram(p.HandleUnshare)
	closeProgram(p.HandleMount)
}

func (m *procscopeMaps) Close() {
	closeMap(m.Events)
	closeMap(m.TrackedPids)
}

func closeProgram(prog *ebpf.Program) {
	if prog != nil {
		_ = prog.Close()
	}
}

func closeMap(m *ebpf.Map) {
	if m != nil {
		_ = m.Close()
	}
}
