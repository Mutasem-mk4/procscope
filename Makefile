# procscope — Makefile
# Requires: Go 1.25+
# Optional for refreshing the committed BPF object: clang, llvm-strip, bpftool

BINARY      := procscope
MODULE      := github.com/Mutasem-mk4/procscope
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GO          ?= go
CLANG       ?= clang
STRIP       ?= llvm-strip
BPFTOOL     ?= bpftool

GOFLAGS     ?=
CGO_ENABLED := 0
LDFLAGS     := -s -w \
  -X '$(MODULE)/internal/version.Version=$(VERSION)' \
  -X '$(MODULE)/internal/version.Commit=$(COMMIT)'
BPF_SRC     := bpf/procscope.c
BPF_OBJ     := internal/tracer/procscope_bpfel.o

# eBPF compilation flags
BPF_CFLAGS  := -O2 -g -Wall -Werror

# Architecture detection
ARCH        := $(shell uname -m)
ifeq ($(ARCH),x86_64)
  GOARCH    := amd64
endif
ifeq ($(ARCH),aarch64)
  GOARCH    := arm64
endif

.PHONY: all build generate test lint clean install uninstall \
        vmlinux fixtures deb arch-pkg help

all: build ## Build the procscope binary

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# --- eBPF ---

vmlinux: ## Generate vmlinux.h from running kernel BTF
	$(BPFTOOL) btf dump file /sys/kernel/btf/vmlinux format c > bpf/headers/vmlinux.h

generate: ## Refresh the committed eBPF object (requires clang with BPF target support)
	$(CLANG) $(BPF_CFLAGS) -target bpfel -I./bpf/headers -c $(BPF_SRC) -o $(BPF_OBJ)
	@which $(STRIP) >/dev/null 2>&1 && $(STRIP) -g $(BPF_OBJ) || true

# --- Build ---

build: ## Build the procscope binary
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=$(GOARCH) \
	  $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/procscope

build-all: ## Cross-compile for amd64 and arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	  $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-amd64 ./cmd/procscope
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
	  $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-arm64 ./cmd/procscope

# --- Test ---

test: ## Run all tests
	$(GO) test -v -race -count=1 ./...

test-unit: ## Run unit tests only
	$(GO) test -v -race -count=1 -short ./...

test-integration: build ## Run Linux smoke test (requires root, Linux)
	./bin/$(BINARY) -- /bin/true

test-cover: ## Run tests with coverage
	$(GO) test -coverprofile=coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

fixtures: ## Build test fixture binaries
	$(MAKE) -C test/fixtures

# --- Quality ---

lint: ## Run linters
	$(GO) vet ./...
	@which staticcheck >/dev/null 2>&1 && staticcheck ./... || \
	  echo "staticcheck not installed; skipping"
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || \
	  echo "golangci-lint not installed; skipping"

fmt: ## Format Go source
	$(GO) fmt ./...
	@which goimports >/dev/null 2>&1 && goimports -w . || true

vuln: ## Check for known vulnerabilities
	@which govulncheck >/dev/null 2>&1 && govulncheck ./... || \
	  echo "govulncheck not installed; install with: go install golang.org/x/vuln/cmd/govulncheck@latest"

# --- Install ---

PREFIX ?= /usr/local

install: build ## Install to PREFIX (default /usr/local)
	install -Dm755 bin/$(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	install -Dm644 man/procscope.1 $(DESTDIR)$(PREFIX)/share/man/man1/procscope.1
	install -Dm644 completions/procscope.bash \
	  $(DESTDIR)$(PREFIX)/share/bash-completion/completions/procscope
	install -Dm644 completions/procscope.zsh \
	  $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_procscope
	install -Dm644 completions/procscope.fish \
	  $(DESTDIR)$(PREFIX)/share/fish/vendor_completions.d/procscope.fish

uninstall: ## Uninstall from PREFIX
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	rm -f $(DESTDIR)$(PREFIX)/share/man/man1/procscope.1
	rm -f $(DESTDIR)$(PREFIX)/share/bash-completion/completions/procscope
	rm -f $(DESTDIR)$(PREFIX)/share/zsh/site-functions/_procscope
	rm -f $(DESTDIR)$(PREFIX)/share/fish/vendor_completions.d/procscope.fish

# --- Packaging ---

deb: ## Build Debian package (requires debuild)
	dpkg-buildpackage -us -uc -b

arch-pkg: ## Build Arch package (requires makepkg)
	cd arch && makepkg -sf

# --- Clean ---

clean: ## Remove build artifacts
	rm -rf bin/ dist/ coverage.out coverage.html
	$(MAKE) -C test/fixtures clean 2>/dev/null || true
