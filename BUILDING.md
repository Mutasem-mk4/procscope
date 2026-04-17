# Building procscope

## Requirements

- Go 1.25 or newer
- Linux (kernel 5.8+ with BTF for runtime)
- clang with BPF target support (only when refreshing the committed eBPF object)
- llvm-strip (optional, when refreshing the committed eBPF object)

## Quick Build

```bash
# Build the binary from the committed eBPF object
make build

# The binary is at ./bin/procscope
```

## Step by Step

### 1. Clone

```bash
git clone https://github.com/Mutasem-mk4/procscope.git
cd procscope
```

### 2. Verify module integrity

```bash
go mod verify
```

This should report `all modules verified`.

### 3. Build

```bash
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/procscope ./cmd/procscope
```

Fresh checkouts already include the committed BPF object at
`internal/tracer/procscope_bpfel.o`, so normal source builds and package builds
do not need to run code generation.

### 4. Refresh the eBPF object after editing `bpf/procscope.c`

```bash
make generate
```

This recompiles `bpf/procscope.c` into the committed
`internal/tracer/procscope_bpfel.o` artifact. Use it only when the eBPF C
source changes.

If your kernel's `vmlinux.h` differs from the bundled minimal subset, refresh it first:

```bash
bpftool btf dump file /sys/kernel/btf/vmlinux format c > bpf/headers/vmlinux.h
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
Cross-compilation of the Go binary works from any platform because the source
tree already includes the committed BPF object for the supported little-endian
targets (amd64, arm64).

## Packaging

### Debian / Kali / Parrot

```bash
# Build-Depends: debhelper-compat (= 13), golang-go
dpkg-buildpackage -us -uc -b
```

### Arch / BlackArch

```bash
cd arch/
makepkg -sf
```

## Troubleshooting

### `make generate` fails with "vmlinux.h: No such file or directory"

Ensure the `bpf/headers/vmlinux.h` file exists.
You may need to regenerate it from a Linux host's kernel BTF:

```bash
bpftool btf dump file /sys/kernel/btf/vmlinux format c > bpf/headers/vmlinux.h
```

### Build succeeds but `procscope` fails at runtime

Check:
1. Kernel version: `uname -r` (must be 5.8+)
2. BTF: `ls /sys/kernel/btf/vmlinux` (must exist)
3. Privileges: run as root or with `CAP_BPF + CAP_PERFMON`

Run `procscope --skip-checks -- true` to see which tracepoints can attach.
