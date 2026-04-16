//go:build linux

package process

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Tree represents a snapshot of a process tree rooted at a given PID.
type Tree struct {
	Root     *TreeNode
	NodesByPID map[uint32]*TreeNode
}

// TreeNode is a single process in the tree.
type TreeNode struct {
	PID      uint32
	PPID     uint32
	Comm     string
	Cmdline  string
	Children []*TreeNode
}

// BuildTreeFromPID reads /proc to build a process tree rooted at the given PID.
// This is used when attaching to an existing process to discover its children.
//
// Best-effort: processes that exit during enumeration will be silently skipped.
func BuildTreeFromPID(rootPID uint32) (*Tree, error) {
	// Verify root PID exists
	if _, err := os.Stat(fmt.Sprintf("/proc/%d", rootPID)); err != nil {
		return nil, fmt.Errorf("process %d not found: %w", rootPID, err)
	}

	// Read all processes from /proc
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc: %w", err)
	}

	type procInfo struct {
		pid     uint32
		ppid    uint32
		comm    string
		cmdline string
	}

	procs := make(map[uint32]*procInfo)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid64, err := strconv.ParseUint(entry.Name(), 10, 32)
		if err != nil {
			continue // not a PID directory
		}
		pid := uint32(pid64)

		info := &procInfo{pid: pid}

		// Read PPID from /proc/[pid]/stat
		statPath := filepath.Join("/proc", entry.Name(), "stat")
		if ppid, err := readPPID(statPath); err == nil {
			info.ppid = ppid
		}

		// Read comm
		commPath := filepath.Join("/proc", entry.Name(), "comm")
		if data, err := os.ReadFile(commPath); err == nil {
			info.comm = strings.TrimSpace(string(data))
		}

		// Read cmdline
		cmdlinePath := filepath.Join("/proc", entry.Name(), "cmdline")
		if data, err := os.ReadFile(cmdlinePath); err == nil {
			// cmdline is null-separated
			info.cmdline = strings.ReplaceAll(strings.TrimRight(string(data), "\x00"), "\x00", " ")
		}

		procs[pid] = info
	}

	// Check root exists
	rootInfo, ok := procs[rootPID]
	if !ok {
		return nil, fmt.Errorf("root process %d disappeared during scan", rootPID)
	}

	// Build tree nodes
	nodes := make(map[uint32]*TreeNode)
	for pid, info := range procs {
		nodes[pid] = &TreeNode{
			PID:     pid,
			PPID:    info.ppid,
			Comm:    info.comm,
			Cmdline: info.cmdline,
		}
	}

	// Link children to parents and find all descendants of rootPID
	descendants := make(map[uint32]bool)
	descendants[rootPID] = true

	// Multi-pass to discover full tree (handles deep nesting)
	changed := true
	for changed {
		changed = false
		for pid, info := range procs {
			if descendants[pid] {
				continue
			}
			if descendants[info.ppid] {
				descendants[pid] = true
				changed = true
			}
		}
	}

	// Build the tree structure
	for pid := range descendants {
		node := nodes[pid]
		if node == nil {
			continue
		}
		if parentNode, ok := nodes[node.PPID]; ok && descendants[node.PPID] && pid != rootPID {
			parentNode.Children = append(parentNode.Children, node)
		}
	}

	tree := &Tree{
		Root:       nodes[rootPID],
		NodesByPID: make(map[uint32]*TreeNode),
	}
	for pid := range descendants {
		if n, ok := nodes[pid]; ok {
			tree.NodesByPID[pid] = n
		}
	}

	_ = rootInfo // used for existence check

	return tree, nil
}

// DescendantPIDs returns all PIDs in the tree (including root).
func (t *Tree) DescendantPIDs() []uint32 {
	pids := make([]uint32, 0, len(t.NodesByPID))
	for pid := range t.NodesByPID {
		pids = append(pids, pid)
	}
	return pids
}

// String returns a human-readable process tree.
func (t *Tree) String() string {
	if t.Root == nil {
		return "(empty tree)"
	}
	var sb strings.Builder
	printNode(&sb, t.Root, "", true)
	return sb.String()
}

func printNode(sb *strings.Builder, node *TreeNode, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if prefix == "" {
		connector = ""
	}

	display := node.Comm
	if node.Cmdline != "" && node.Cmdline != node.Comm {
		display = node.Cmdline
	}
	fmt.Fprintf(sb, "%s%s[%d] %s\n", prefix, connector, node.PID, display)

	childPrefix := prefix
	if prefix != "" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i, child := range node.Children {
		printNode(sb, child, childPrefix, i == len(node.Children)-1)
	}
}

// readPPID parses the PPID from /proc/[pid]/stat.
// Format: pid (comm) state ppid ...
func readPPID(statPath string) (uint32, error) {
	f, err := os.Open(statPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return 0, fmt.Errorf("empty stat file")
	}
	line := scanner.Text()

	// Find the closing paren of comm field (comm can contain spaces/parens)
	closeParen := strings.LastIndex(line, ")")
	if closeParen < 0 || closeParen+2 >= len(line) {
		return 0, fmt.Errorf("malformed stat line")
	}

	// Fields after ") " are: state ppid pgrp ...
	rest := line[closeParen+2:]
	fields := strings.Fields(rest)
	if len(fields) < 2 {
		return 0, fmt.Errorf("insufficient fields in stat")
	}

	ppid, err := strconv.ParseUint(fields[1], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid ppid: %w", err)
	}

	return uint32(ppid), nil
}

// FindPIDByName searches /proc for processes matching the given name.
// Returns all matching PIDs. Best-effort.
func FindPIDByName(name string) ([]uint32, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc: %w", err)
	}

	var pids []uint32
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid64, err := strconv.ParseUint(entry.Name(), 10, 32)
		if err != nil {
			continue
		}

		commPath := filepath.Join("/proc", entry.Name(), "comm")
		data, err := os.ReadFile(commPath)
		if err != nil {
			continue
		}

		comm := strings.TrimSpace(string(data))
		if comm == name {
			pids = append(pids, uint32(pid64))
		}
	}

	return pids, nil
}
