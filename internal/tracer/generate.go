//go:build linux

// Package tracer manages eBPF program loading, attachment, and event reading
// for procscope runtime investigations.
//
// This package is Linux-only and requires kernel 5.8+ with BTF support.
package tracer

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall -Werror" -target bpfel,bpfeb -type event procscope ../../bpf/procscope.c -- -I../../bpf/headers
