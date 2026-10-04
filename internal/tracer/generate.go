//go:build linux

// Package tracer manages eBPF program loading, attachment, and event reading
// for procscope runtime investigations.
//
// This package is Linux-only and requires kernel 5.8+ with BTF support.
package tracer

// The generated BPF object lives next to this package for embedding.
// `make build` compiles it from ../../bpf/procscope.c when missing or stale;
// Debian packaging forces regeneration from source.
