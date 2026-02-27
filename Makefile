.DEFAULT_GOAL := help

# ─── Go ──────────────────────────────────────────────────────────

.PHONY: lint-go test-go build-go install-go

lint-go: ## Run Go linters
	golangci-lint run ./...

test-go: ## Run Go tests with race detector
	go test -v -race -coverprofile=coverage-go.out ./...

build-go: ## Build all Go tools (version from VERSION file)
	@for tool in tools/*/; do \
		name=$$(basename "$$tool"); \
		[ -f "$$tool/Cargo.toml" ] && continue; \
		ver="dev"; \
		if [ -f "$$tool/VERSION" ]; then ver=$$(cat "$$tool/VERSION" | tr -d '\n'); fi; \
		echo "=== Building $$name v$$ver ==="; \
		CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=v$$ver" -o bin/$$name ./$$tool; \
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

.PHONY: release-tag versions

versions: ## Show current version of each tool
	@for vf in tools/*/VERSION libs/*/VERSION; do \
		[ -f "$$vf" ] || continue; \
		dir=$$(dirname "$$vf"); \
		name=$$(basename "$$dir"); \
		ver=$$(cat "$$vf" | tr -d '\n'); \
		printf "  %-28s %s\n" "$$name" "v$$ver"; \
	done

release-tag: ## Tag a tool for release (usage: make release-tag TOOL=banner-grabber)
ifndef TOOL
	$(error TOOL is required — e.g. make release-tag TOOL=banner-grabber)
endif
	@VERSION_FILE=""; \
	if [ -f "tools/$(TOOL)/VERSION" ]; then VERSION_FILE="tools/$(TOOL)/VERSION"; \
	elif [ -f "libs/$(TOOL)/VERSION" ]; then VERSION_FILE="libs/$(TOOL)/VERSION"; \
	else echo "ERROR: No VERSION file found for $(TOOL)"; exit 1; fi; \
	VER=$$(cat "$$VERSION_FILE" | tr -d '\n'); \
	echo "Creating tag $(TOOL)/v$$VER from $$VERSION_FILE"; \
	git tag -a "$(TOOL)/v$$VER" -m "Release $(TOOL) v$$VER"; \
	echo "Tag created: $(TOOL)/v$$VER"; \
	echo "Push with:   git push origin $(TOOL)/v$$VER"

# ─── Help ────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
