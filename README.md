# y2psync

Cross-platform peer-to-peer YouTube playlist and subscription sync tool. Built with Go, cross-compiled via Zig.

## Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- [Zig](https://ziglang.org/download/) 0.16+

## Building

The build system uses Zig to orchestrate `go build` with the correct environment variables and optionally provides `zig cc` as a cross-compiler for non-native targets.

### Native build

```bash
zig build
```

Output goes to `build/<os>-<arch>/y2psync` (or `y2psync.exe` on Windows).

### Cross-compilation

Use `-Dos` to select the target OS and optionally `-Darch` to select the architecture:

| Command | Target |
|---------|--------|
| `zig build -Dos=linux` | Linux x86_64 |
| `zig build -Dos=linux -Darch=arm64` | Linux ARM64 |
| `zig build -Dos=windows` | Windows x86_64 |
| `zig build -Dos=mac` | macOS ARM64 (Apple Silicon) |
| `zig build -Dos=mac -Darch=amd64` | macOS x86_64 (Intel) |

You can also use the standard Zig target triple directly:

```bash
zig build -Dtarget=x86_64-linux
zig build -Dtarget=aarch64-linux
zig build -Dtarget=x86_64-windows
zig build -Dtarget=aarch64-macos
zig build -Dtarget=x86_64-macos
```

### Options

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `-Dos` | `linux`, `windows`, `mac`, `darwin`, `macos` | native | Target OS |
| `-Darch` | `amd64`, `x86_64`, `arm64`, `aarch64` | `amd64` (linux/windows), `arm64` (mac) | Target architecture |
| `-Dtarget` | Zig target triple | native | Full target specification (overrides `-Dos`/`-Darch`) |

### Steps

| Command | Description |
|---------|-------------|
| `zig build` | Build for native or specified target |
| `zig build clean` | Remove `build/` directory |

### Output

All builds are placed in `build/<goos>-<goarch>/`:

```
build/linux-amd64/y2psync
build/linux-arm64/y2psync
build/windows-amd64/y2psync.exe
build/darwin-amd64/y2psync
build/darwin-arm64/y2psync
```
