//go:build linux && !(amd64 || arm64)

package tracer

import (
	"fmt"
	"runtime"

	"github.com/cilium/ebpf"
)

func loadProcscope() (*ebpf.CollectionSpec, error) {
	return nil, fmt.Errorf("unsupported Linux architecture %s: committed BPF object is currently shipped for amd64 and arm64 only", runtime.GOARCH)
}
