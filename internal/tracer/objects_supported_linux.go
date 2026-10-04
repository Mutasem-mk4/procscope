//go:build linux && (amd64 || arm64)

package tracer

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/cilium/ebpf"
)

//go:embed procscope_bpfel.o
var procscopeBPFEL []byte

func loadProcscope() (*ebpf.CollectionSpec, error) {
	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(procscopeBPFEL))
	if err != nil {
		return nil, fmt.Errorf("load embedded BPF object: %w", err)
	}
	return spec, nil
}
