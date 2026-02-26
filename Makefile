.DEFAULT_GOAL := help

# ─── Go ──────────────────────────────────────────────────────────

.PHONY: lint-go test-go build-go install-go

lint-go: ## Run Go linters
	golangci-lint run ./...

test-go: ## Run Go tests with race detector
	go test -v -race -coverprofile=coverage-go.out ./...

build-go: ## Build all Go tools
	@for tool in tools/*/; do \
		name=$$(basename "$$tool"); \
		echo "=== Building $$name ==="; \
		CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=dev" -o bin/$$name ./$$tool; \
	done

install-go: ## Install all Go tools locally
	@for tool in tools/*/; do \
		echo "=== Installing $$(basename $$tool) ==="; \
		go install ./$$tool; \
	done

# ─── Rust ────────────────────────────────────────────────────────

.PHONY: lint-rust test-rust build-rust

lint-rust: ## Run Rust linters
	cargo clippy --all-targets --all-features -- -D warnings
	cargo fmt --all -- --check

test-rust: ## Run Rust tests
	cargo test --all

build-rust: ## Build all Rust crates in release mode
	cargo build --release

# ─── Aggregate ───────────────────────────────────────────────────

.PHONY: lint test build clean

lint: lint-go lint-rust ## Run all linters

test: test-go test-rust ## Run all tests

build: build-go build-rust ## Build everything

clean: ## Remove build artifacts
	rm -rf bin/ coverage-go.out
	cargo clean

# ─── Git hooks ───────────────────────────────────────────────────

.PHONY: hooks

hooks: ## Install git hooks (.githooks → .git/hooks)
	git config core.hooksPath .githooks
	@echo "Git hooks activated from .githooks/"

# ─── Release helpers ─────────────────────────────────────────────

.PHONY: release-tag

release-tag: ## Create a per-tool release tag (usage: make release-tag TOOL=banner-grabber VERSION=v1.0.0)
ifndef TOOL
	$(error TOOL is required — e.g. make release-tag TOOL=banner-grabber VERSION=v1.0.0)
endif
ifndef VERSION
	$(error VERSION is required — e.g. make release-tag TOOL=banner-grabber VERSION=v1.0.0)
endif
	git tag -a "$(TOOL)/$(VERSION)" -m "Release $(TOOL) $(VERSION)"
	@echo "Tag created: $(TOOL)/$(VERSION)"
	@echo "Push with:   git push origin $(TOOL)/$(VERSION)"

# ─── Help ────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
