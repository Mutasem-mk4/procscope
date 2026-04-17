//go:build !linux

package process

import (
	"fmt"
	"os"
)

// Launcher manages launching a command under observation.
type Launcher struct{}

func NewLauncher(_ []string) (*Launcher, error) {
	return nil, fmt.Errorf("process launcher requires Linux")
}

func (l *Launcher) Start() (uint32, error)      { return 0, fmt.Errorf("requires Linux") }
func (l *Launcher) Wait() error                 { return fmt.Errorf("requires Linux") }
func (l *Launcher) Done() <-chan struct{}        { ch := make(chan struct{}); close(ch); return ch }
func (l *Launcher) ExitCode() int               { return -1 }
func (l *Launcher) Signal(_ os.Signal) error     { return fmt.Errorf("requires Linux") }
func (l *Launcher) PID() uint32                 { return 0 }
func (l *Launcher) Args() []string              { return nil }
func (l *Launcher) CommandString() string        { return "" }

// Tree represents a snapshot of a process tree.
type Tree struct {
	Root       *TreeNode
	NodesByPID map[uint32]*TreeNode
}

type TreeNode struct {
	PID      uint32
	PPID     uint32
	Comm     string
	Cmdline  string
	Children []*TreeNode
}

func BuildTreeFromPID(_ uint32) (*Tree, error) {
	return nil, fmt.Errorf("process tree requires Linux /proc")
}

func (t *Tree) DescendantPIDs() []uint32 { return nil }
func (t *Tree) String() string           { return "(not available on this platform)" }

func FindPIDByName(_ string) ([]uint32, error) {
	return nil, fmt.Errorf("process search requires Linux /proc")
}
