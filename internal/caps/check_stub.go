//go:build !linux

package caps

// CheckResult contains the privilege check results.
type CheckResult struct {
	IsRoot        bool
	Capabilities  map[string]bool
	BTFAvailable  bool
	KernelVersion string
	Errors        []string
	Warnings      []string
}

// Check returns a stub result on non-Linux platforms.
func Check() *CheckResult {
	return &CheckResult{
		Errors: []string{"procscope requires Linux"},
	}
}

// CanProceed returns false on non-Linux.
func (r *CheckResult) CanProceed() bool {
	return false
}

// Summary returns a stub message.
func (r *CheckResult) Summary() string {
	return "procscope requires Linux with kernel 5.8+ and BTF support.\n"
}
