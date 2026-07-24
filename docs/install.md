# Installation Guide

This guide covers the various ways to install `procscope`.

> **Note:** Running `procscope` usually requires `sudo` (eBPF capabilities).

## 1. Homebrew (Recommended)

```bash
brew tap Mutasem-mk4/kharma
brew install procscope
```

## 2. Go Install (Requires Go 1.26+)

```bash
go install github.com/Mutasem-mk4/procscope/cmd/procscope@latest
```

## 3. Direct Download

Download the release asset that matches your architecture from:

- [Latest Releases](https://github.com/Mutasem-mk4/procscope/releases/latest)

Available assets include:
- Debian package (`.deb`)
- Linux tarballs for `amd64` and `arm64`

## 4. Build from Source

Ensure you have Go 1.26+ and `make` installed.

```bash
git clone https://github.com/Mutasem-mk4/procscope.git
cd procscope
make build
sudo install -m755 bin/procscope /usr/local/bin/procscope
```

## 5. Native Package Managers

`procscope` is becoming available in official repositories.

**BlackArch Linux:**
```bash
sudo pacman -S procscope
```

**Kali Linux & Parrot OS:**
```bash
# Pending official inclusion; use the .deb package from releases in the meantime.
sudo dpkg -i procscope_*.deb
```
