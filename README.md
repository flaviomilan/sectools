# sectools

[![Go](https://github.com/flaviomilan/sectools/actions/workflows/go.yml/badge.svg)](https://github.com/flaviomilan/sectools/actions/workflows/go.yml)
[![Rust](https://github.com/flaviomilan/sectools/actions/workflows/rust.yml/badge.svg)](https://github.com/flaviomilan/sectools/actions/workflows/rust.yml)
[![Security](https://github.com/flaviomilan/sectools/actions/workflows/security.yml/badge.svg)](https://github.com/flaviomilan/sectools/actions/workflows/security.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A curated monorepo of security tools built with **Go** and **Rust**.  
Each tool is independently versioned and released as a standalone binary.

---

## Tools

| Tool | Language | Description |
|------|----------|-------------|
| **banner-grabber** | Go | TCP banner grabbing — probes open ports and captures service banners |
| **port-knocking-scanner** | Go | Detects port-knocking sequences using raw packet capture (gopacket/pcap) |
| **sectools-common** | Rust | Shared library with network utilities (IP validation, port parsing, banner grab) |

## Project Structure

```
sectools/
├── tools/
│   ├── banner-grabber/          # Go CLI tool
│   └── port-knocking-scanner/   # Go CLI tool
├── libs/
│   ├── netutil/                 # Shared Go library
│   └── sectools-common/         # Shared Rust library
├── .github/
│   └── workflows/
│       ├── go.yml               # Go lint, test, build
│       ├── rust.yml             # Rust lint, test, build
│       ├── security.yml         # govulncheck, cargo-audit, Trivy, CodeQL
│       └── release.yml          # Per-tool release on tag push
├── go.mod                       # Go module
├── Cargo.toml                   # Rust workspace
├── Makefile
└── ...
```

## Installation

### From source (Go tools)

```bash
go install github.com/flaviomilan/sectools/tools/banner-grabber@latest
go install github.com/flaviomilan/sectools/tools/port-knocking-scanner@latest
```

### Pre-built binaries

Download from [Releases](https://github.com/flaviomilan/sectools/releases).  
Each tool has its own release page with binaries for Linux, macOS, and Windows.

```bash
# Example: install banner-grabber on Linux amd64
curl -Lo banner-grabber \
  https://github.com/flaviomilan/sectools/releases/download/banner-grabber%2Fv1.0.0/banner-grabber-linux-amd64
chmod +x banner-grabber
sudo mv banner-grabber /usr/local/bin/
```

### Build locally

```bash
make build        # Build all (Go + Rust)
make build-go     # Build Go tools only → bin/
make build-rust   # Build Rust crates only
```

## Usage

### banner-grabber

```bash
banner-grabber -host 192.168.1.1 -ports 22,80,443 -timeout 5s
banner-grabber -host 10.0.0.1 -ports 1-1024 -send "HEAD / HTTP/1.0\r\n\r\n" -output results.txt
banner-grabber -version
```

### port-knocking-scanner

> Requires root / `CAP_NET_RAW` for raw packet capture.

```bash
sudo port-knocking-scanner -target 192.168.1.1 -ports 7000,8000,9000
sudo port-knocking-scanner -target 10.0.0.1 -ports 7000,8000,9000 -timeout 10s
port-knocking-scanner -version
```

## Development

### Prerequisites

- **Go** ≥ 1.24
- **Rust** ≥ 1.75 (2021 edition)
- **libpcap-dev** (for port-knocking-scanner)
- **golangci-lint** (for Go linting)

### Common tasks

```bash
make help          # Show all available targets
make lint          # Lint Go + Rust
make test          # Test Go + Rust
make build         # Build everything
make clean         # Remove artifacts
```

## Release Process

Each tool is versioned and released independently using the tag pattern:

```
<tool-name>/v<semver>
```

### Creating a release

```bash
# Tag a specific tool with a version
make release-tag TOOL=banner-grabber VERSION=v1.0.0

# Push the tag to trigger the release pipeline
git push origin banner-grabber/v1.0.0
```

The release workflow will:

1. Detect which tool to release from the tag prefix
2. Build cross-platform binaries (linux/darwin/windows × amd64/arm64)
3. Generate SHA-256 checksums
4. Create a GitHub Release with changelog, install instructions, and assets

### Version history

Tags follow the convention `<tool>/v<major>.<minor>.<patch>`:

| Tag example | Effect |
|-------------|--------|
| `banner-grabber/v1.0.0` | Releases banner-grabber v1.0.0 |
| `port-knocking-scanner/v0.2.0` | Releases port-knocking-scanner v0.2.0 |

Each tool's version is fully independent — releasing one tool does **not** affect others.

## CI / CD

| Workflow | Trigger | What it does |
|----------|---------|--------------|
| **Go** | Push/PR touching `tools/`, `libs/netutil/`, `go.mod` | golangci-lint → tests (race + coverage) → build |
| **Rust** | Push/PR touching `libs/sectools-common/`, `Cargo.toml` | clippy + fmt → tests → release build |
| **Security** | Push/PR to main + weekly cron | govulncheck, cargo-audit, Trivy, CodeQL |
| **Release** | Tag `<tool>/v*` | Cross-compile, checksum, GitHub Release |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE) — see [LICENSE](LICENSE) for details.
