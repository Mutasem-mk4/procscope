package process

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// containerIDRegex extracts exactly 64-character hex strings, common to Docker/Containerd/CRI-O ids.
var containerIDRegex = regexp.MustCompile(`([a-f0-9]{64})`)

// GetContainerID parses /proc/[pid]/cgroup to extract the container ID.
// Returns an empty string if the process is not in a recognized container.
func GetContainerID(pid uint32) string {
	path := filepath.Join("/proc", fmt.Sprintf("%d", pid), "cgroup")
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Docker, containerd, and cri-o all embed the 64-char ID in the cgroup path.
		// e.g. 0::/kubepods.slice/.../cri-containerd-123456789abcdef...scope
		matches := containerIDRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}
