# Building procscope

## Requirements

- Go 1.22 or newer
- Linux (kernel 5.8+ with BTF for runtime)
- clang (for compiling eBPF C programs)
- llvm-strip (optional, for stripping eBPF objects)

## Quick Build

```bash
# 1. Generate eBPF Go bindings (requires Linux + clang)
make generate

# 2. Build the binary
make build

# The binary is at ./bin/procscope
```

## Step by Step

### 1. Clone

```bash
git clone https://github.com/procscope/procscope.git
cd procscope
```

### 2. Verify module integrity

```bash
go mod verify
```

This should report `all modules verified`.

### 3. Generate eBPF bindings

```bash
cd internal/tracer
go generate ./...
cd ../..
```

This runs `bpf2go` which:
- Compiles `bpf/procscope.c` into BPF ELF objects (`.o`)
- Generates Go source files (`.go`) that embed the objects
- Output files are architecture-specific (bpfel for little-endian, bpfeb for big-endian)

If your kernel's `vmlinux.h` differs from the bundled minimal subset, regenerate it:

```bash
bpftool btf dump file /sys/kernel/btf/vmlinux format c > bpf/headers/vmlinux.h
```

### 4. Build

```bash
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/procscope ./cmd/procscope
```

Or simply:

```bash
make build
```

### 5. Test (unit tests, no root needed)

```bash
go test -short ./...
```

### 6. Test (integration, requires root + eBPF)

```bash
sudo go test -v -count=1 -run Integration ./test/integration/...
```

## Cross-Compilation

```bash
# amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o bin/procscope-linux-amd64 ./cmd/procscope

# arm64
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o bin/procscope-linux-arm64 ./cmd/procscope
```

Note: eBPF generation (`go generate`) must run on a Linux host.
Cross-compilation of the Go binary works from any platform after generation.

## Packaging

### Debian / Kali / Parrot

```bash
# Build-Depends: debhelper-compat (= 13), golang-go, clang, llvm
dpkg-buildpackage -us -uc -b
```

### Arch / BlackArch

```bash
cd arch/
makepkg -sf
```

## Troubleshooting

### `go generate` fails with "vmlinux.h: No such file or directory"

Ensure you are running on Linux and the `bpf/headers/vmlinux.h` file exists.
You may need to regenerate it from your kernel's BTF:

```bash
bpftool btf dump file /sys/kernel/btf/vmlinux format c > bpf/headers/vmlinux.h
```

### Build succeeds but `procscope` fails at runtime

Check:
1. Kernel version: `uname -r` (must be 5.8+)
2. BTF: `ls /sys/kernel/btf/vmlinux` (must exist)
3. Privileges: run as root or with `CAP_BPF + CAP_PERFMON`

Run `procscope --skip-checks -- true` to see which tracepoints can attach.
