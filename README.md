# y2psync

Cross-platform peer-to-peer YouTube playlist and subscription sync tool. Built with Go, cross-compiled via Zig.

## Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- [Zig](https://ziglang.org/download/) 0.16+

## Building

The build system uses Zig to orchestrate `go build` with `zig cc` as the C cross-compiler. All builds use `zig cc` targeting musl (Linux) or gnu (Windows) to produce **fully static binaries** — no external libc dependencies. macOS binaries link Apple's system frameworks and are not fully static.

> **WSL2 note**: If the project is on a Windows filesystem (under `/mnt/`), Zig's cache locking fails. Either move the repo to the Linux filesystem (`~/projects/`) or set:
> ```bash
> export ZIG_CACHE_DIR=$HOME/.cache/zig
> ```

### Commands

| Command | Output |
|---------|--------|
| `zig build` | Native target (auto-detected) |
| `zig build -Dos=linux` | `build/linux-amd64/y2psync` (static musl) |
| `zig build -Dos=linux -Darch=arm64` | `build/linux-arm64/y2psync` (static musl) |
| `zig build -Dos=windows` | `build/windows-amd64/y2psync.exe` (static) |
| `zig build -Dos=mac` | `build/darwin-arm64/y2psync` |
| `zig build -Dos=mac -Darch=amd64` | `build/darwin-amd64/y2psync` |
| `zig build clean` | Remove `build/` directory |

### Options

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `-Dos` | `linux`, `windows`, `mac`, `darwin`, `macos` | native | Target OS |
| `-Darch` | `amd64`, `x86_64`, `arm64`, `aarch64` | `amd64` (linux/windows), `arm64` (mac) | Target architecture |

### Linking

| Platform | libc | Linkage |
|----------|------|---------|
| Linux | musl (zig cc) | Fully static |
| Windows | mingw (zig cc) | Fully static |
| macOS | System SDK (zig cc) | Dynamic (Apple frameworks) |
