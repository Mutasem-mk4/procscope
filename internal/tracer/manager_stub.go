//go:build !linux

// Package tracer manages eBPF program loading, attachment, and event reading
// for procscope runtime investigations.
//
// This stub is used on non-Linux platforms for compilation and testing.
// Actual eBPF functionality requires Linux 5.8+ with BTF.
package tracer

import (
	"context"
	"fmt"
	"runtime"

	"github.com/Mutasem-mk4/procscope/internal/events"
)

// Manager is the non-Linux stub. All methods return errors indicating
// that eBPF is not available on this platform.
type Manager struct {
	correlator *events.Correlator
}

// NewManager creates a stub manager.
func NewManager(correlator *events.Correlator) *Manager {
	return &Manager{correlator: correlator}
}

func (m *Manager) Load() error {
	return fmt.Errorf("eBPF tracer requires Linux (current: %s/%s)", runtime.GOOS, runtime.GOARCH)
}

func (m *Manager) TrackPID(_ uint32) error {
	return fmt.Errorf("eBPF tracer requires Linux")
}

func (m *Manager) UntrackPID(_ uint32) error {
	return fmt.Errorf("eBPF tracer requires Linux")
}

func (m *Manager) Attach() error {
	return fmt.Errorf("eBPF tracer requires Linux")
}

func (m *Manager) ReadEvents(_ context.Context) error {
	return fmt.Errorf("eBPF tracer requires Linux")
}

func (m *Manager) Close() {}
