# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

eth-call — a Go project for Ethereum call operations.
Module: `github.com/rootwarp/eth-call`

## Build & Development Commands

```bash
make build          # Build binary to bin/eth-call
make run            # Build and run
make test           # Run all tests with race detector
make test-coverage  # Tests with coverage report (coverage.html)
make lint           # Run golangci-lint
make fmt            # Format with gofmt + goimports
make mod-tidy       # Tidy go modules
make clean          # Remove build artifacts
```

Run a single test:
```bash
go test -v -run TestName ./path/to/package/...
```

## Architecture

Full service layout:
- `cmd/eth-call/` — entrypoint (keep minimal, delegate to internal)
- `internal/eth-call/` — private application logic
- `pkg/` — packages intended for external import
- `api/` — API definitions (protobuf, OpenAPI, etc.)
- `configs/` — configuration files and templates
- `deployments/` — deployment configurations
- `test/data/` — test fixtures and data
