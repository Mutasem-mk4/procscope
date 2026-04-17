//go:build linux

// Package caps provides runtime capability and privilege detection for procscope.
//
// procscope requires specific Linux capabilities to load eBPF programs and
// attach tracepoint probes. This package detects the current privilege level
// and provides actionable guidance when permissions are insufficient.
package caps

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

// RequiredCapabilities documents the capabilities procscope needs.
//
// Minimum required (non-root):
//   - CAP_BPF          — load eBPF programs (kernel 5.8+)
//   - CAP_PERFMON      — attach perf/tracepoint probes (kernel 5.8+)
//   - CAP_SYS_RESOURCE — increase RLIMIT_MEMLOCK for eBPF maps
//
// Full functionality (root or equivalent):
//   - CAP_SYS_ADMIN    — legacy fallback for kernels without CAP_BPF
//   - CAP_SYS_PTRACE   — read /proc/[pid]/... of other users' processes
var RequiredCapabilities = []string{
	"CAP_BPF",
	"CAP_PERFMON",
	"CAP_SYS_RESOURCE",
}

// CheckResult contains the privilege check results.
type CheckResult struct {
	IsRoot        bool
	Capabilities  map[string]bool
	BTFAvailable  bool
	KernelVersion string
	Errors        []string
	Warnings      []string
}

// Check performs runtime privilege and capability detection.
func Check() *CheckResult {
	result := &CheckResult{
		Capabilities: make(map[string]bool),
	}

	// Check if running as root
	result.IsRoot = os.Geteuid() == 0

	// Check kernel version
	var uname unix.Utsname
	if err := unix.Uname(&uname); err == nil {
		result.KernelVersion = utsStr(uname.Release[:])
	}

	// Check BTF availability
	if _, err := os.Stat("/sys/kernel/btf/vmlinux"); err == nil {
		result.BTFAvailable = true
	} else {
		result.BTFAvailable = false
		result.Errors = append(result.Errors,
			"BTF not available at /sys/kernel/btf/vmlinux — CO-RE eBPF programs will not load. "+
				"Ensure CONFIG_DEBUG_INFO_BTF=y in your kernel config.")
	}

	// Check capabilities (only meaningful when not root)
	if !result.IsRoot {
		// Try to read capabilities from /proc/self/status
		data, err := os.ReadFile("/proc/self/status")
		if err == nil {
			capEff := parseCapEffective(string(data))
			result.Capabilities["CAP_BPF"] = hasCapBit(capEff, 39)          // CAP_BPF = 39
			result.Capabilities["CAP_PERFMON"] = hasCapBit(capEff, 38)      // CAP_PERFMON = 38
			result.Capabilities["CAP_SYS_RESOURCE"] = hasCapBit(capEff, 24) // CAP_SYS_RESOURCE = 24
			result.Capabilities["CAP_SYS_ADMIN"] = hasCapBit(capEff, 21)    // CAP_SYS_ADMIN = 21
			result.Capabilities["CAP_SYS_PTRACE"] = hasCapBit(capEff, 19)   // CAP_SYS_PTRACE = 19
		}

		// Check if we have the minimum capabilities
		hasBPF := result.Capabilities["CAP_BPF"] || result.Capabilities["CAP_SYS_ADMIN"]
		hasPerfmon := result.Capabilities["CAP_PERFMON"] || result.Capabilities["CAP_SYS_ADMIN"]

		if !hasBPF {
			result.Errors = append(result.Errors,
				"Missing CAP_BPF (or CAP_SYS_ADMIN). Cannot load eBPF programs. "+
					"Run as root or grant capabilities: sudo setcap cap_bpf,cap_perfmon,cap_sys_resource+ep $(which procscope)")
		}
		if !hasPerfmon {
			result.Errors = append(result.Errors,
				"Missing CAP_PERFMON (or CAP_SYS_ADMIN). Cannot attach tracepoint probes.")
		}
		if !result.Capabilities["CAP_SYS_PTRACE"] {
			result.Warnings = append(result.Warnings,
				"Missing CAP_SYS_PTRACE — attaching to other users' processes may fail.")
		}
	}

	// Check RLIMIT_MEMLOCK
	var rlim unix.Rlimit
	if err := unix.Getrlimit(unix.RLIMIT_MEMLOCK, &rlim); err == nil {
		if rlim.Cur < 64*1024*1024 && !result.IsRoot {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("RLIMIT_MEMLOCK is %d bytes (recommended: 64MB+). "+
					"eBPF map allocation may fail. Increase with: ulimit -l unlimited",
					rlim.Cur))
		}
	}

	// Kernel version check (need 5.8+ for ring buffer + CAP_BPF)
	if result.KernelVersion != "" {
		major, minor := parseKernelVersion(result.KernelVersion)
		if major < 5 || (major == 5 && minor < 8) {
			result.Errors = append(result.Errors,
				fmt.Sprintf("Kernel %s is below minimum 5.8. "+
					"procscope requires kernel 5.8+ for BPF ring buffer and CAP_BPF support.",
					result.KernelVersion))
		}
	}

	return result
}

// CanProceed returns true if there are no blocking errors.
func (r *CheckResult) CanProceed() bool {
	return len(r.Errors) == 0
}

// Summary returns a human-readable privilege summary.
func (r *CheckResult) Summary() string {
	var sb strings.Builder

	sb.WriteString("procscope privilege check:\n")
	sb.WriteString(fmt.Sprintf("  Kernel:  %s\n", r.KernelVersion))
	sb.WriteString(fmt.Sprintf("  Root:    %v\n", r.IsRoot))
	sb.WriteString(fmt.Sprintf("  BTF:     %v\n", r.BTFAvailable))

	if !r.IsRoot {
		sb.WriteString("  Capabilities:\n")
		for cap, has := range r.Capabilities {
			marker := "x"
			if has {
				marker = "ok"
			}
			sb.WriteString(fmt.Sprintf("    [%s] %s\n", marker, cap))
		}
	}

	if len(r.Warnings) > 0 {
		sb.WriteString("\n  Warnings:\n")
		for _, w := range r.Warnings {
			sb.WriteString(fmt.Sprintf("    - %s\n", w))
		}
	}

	if len(r.Errors) > 0 {
		sb.WriteString("\n  Errors:\n")
		for _, e := range r.Errors {
			sb.WriteString(fmt.Sprintf("    - %s\n", e))
		}
	}

	return sb.String()
}

// parseCapEffective extracts CapEff hex string from /proc/self/status.
func parseCapEffective(status string) uint64 {
	for _, line := range strings.Split(status, "\n") {
		if strings.HasPrefix(line, "CapEff:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				var val uint64
				fmt.Sscanf(fields[1], "%x", &val)
				return val
			}
		}
	}
	return 0
}

// hasCapBit checks if a specific capability bit is set.
func hasCapBit(caps uint64, bit int) bool {
	return caps&(1<<uint(bit)) != 0
}

// parseKernelVersion extracts major.minor from a kernel release string.
func parseKernelVersion(release string) (int, int) {
	var major, minor int
	fmt.Sscanf(release, "%d.%d", &major, &minor)
	return major, minor
}

// utsStr converts a Utsname field to a Go string, stopping at the first null
// byte. The x/sys unix package exposes these fields as either []byte or []int8
// depending on the target platform.
func utsStr[T ~byte | ~int8](raw []T) string {
	n := 0
	for ; n < len(raw); n++ {
		if raw[n] == 0 {
			break
		}
	}

	buf := make([]byte, n)
	for i := 0; i < n; i++ {
		buf[i] = byte(raw[i])
	}
	return string(buf)
}
