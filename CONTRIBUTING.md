# Contributing to sectools

Thank you for your interest in contributing! This guide will help you get started.

## Project Structure

```
sectools/
├── tools/                  # CLI tools (one folder per tool)
│   ├── banner-grabber/
│   └── port-knocking-scanner/
├── libs/
│   ├── netutil/            # Shared Go library
│   └── sectools-common/   # Shared Rust library
├── Cargo.toml              # Rust workspace manifest
├── .github/workflows/      # CI/CD pipelines
├── go.mod                  # Go module (root)
└── Makefile
```

## Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Go | ≥ 1.24 | Build & test Go tools |
| Rust | ≥ 1.75 | Build & test Rust crates |
| libpcap-dev | any | Required by port-knocking-scanner (CGO) |
| golangci-lint | latest | Go linting |

## Getting Started

```bash
git clone https://github.com/flaviomilan/sectools.git
cd sectools
make build
make test
```

## Development Workflow

1. **Fork and branch** — create a feature branch from `main`.
2. **Make changes** — follow the conventions below.
3. **Test locally** — run `make lint && make test`.
4. **Commit** — use [Conventional Commits](https://www.conventionalcommits.org/) format.
5. **Open a PR** — target `main`, fill out the PR template.

## Adding a New Go Tool

1. Create `tools/<tool-name>/main.go` with a `main` package.
2. Add `var version = "dev"` so the release pipeline can inject the version via `-ldflags`.
3. Reuse shared utilities from `libs/netutil/` when possible.
4. Add tests alongside your code.
5. The CI and release workflows pick up new tools automatically.

```bash
# Verify it builds
go build ./tools/<tool-name>

# Run Go tests
make test-go
```

## Adding a New Rust Crate

1. Create `libs/<crate-name>/` with a `Cargo.toml` that inherits workspace settings.
2. Add the crate to `members` in the root `Cargo.toml`.
3. Add tests in `src/lib.rs` or a `tests/` directory.

```bash
# Verify it builds
cargo build -p <crate-name>

# Run Rust tests
make test-rust
```

## Code Style

### Go

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines.
- All code must pass `golangci-lint` (see `.golangci.yml` for enabled linters).
- Export functions and types that are shared; keep tool-specific logic private.

### Rust

- Follow standard Rust idioms and `clippy` recommendations.
- Code must pass `cargo fmt --check` and `cargo clippy -- -D warnings`.
- Use the workspace `Cargo.toml` for shared dependency versions.

## Running Tests

```bash
make test          # All tests (Go + Rust)
make test-go       # Go tests with race detector
make test-rust     # Rust tests
```

## Linting

```bash
make lint          # All linters
make lint-go       # golangci-lint
make lint-rust     # clippy + rustfmt
```

## Release Process

Each tool is versioned independently. To release a tool:

```bash
# Create and push a tag
make release-tag TOOL=banner-grabber VERSION=v1.2.0
git push origin banner-grabber/v1.2.0
```

The release workflow handles cross-compilation, checksums, and GitHub Release creation automatically.

**Tag convention:** `<tool-name>/v<major>.<minor>.<patch>`

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(banner-grabber): add JSON output format
fix(netutil): handle IPv6 addresses correctly
docs: update README installation section
ci: add arm64 build target
```

## Reporting Issues

- Use the [Bug Report](https://github.com/flaviomilan/sectools/issues/new?template=bug_report.md) template for bugs.
- Use the [Feature Request](https://github.com/flaviomilan/sectools/issues/new?template=feature_request.md) template for ideas.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).
