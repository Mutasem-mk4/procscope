//go:build linux

// Package tracer manages eBPF program loading, attachment, and event reading
// for procscope runtime investigations.
//
// This package is Linux-only and requires kernel 5.8+ with BTF support.
package tracer

// The committed BPF object lives next to this package so source builds and
// package builds do not need to invoke code generation. Refresh it with
// `make generate` after editing ../../bpf/procscope.c.
