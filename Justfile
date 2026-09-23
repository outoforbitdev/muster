#!/usr/bin/env just --justfile

# Default recipe displays help
default:
    @just --list

# Bootstrap: one-time repository setup
bootstrap: install
    @echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    @echo "✅ Bootstrap complete"

# Install: install project dependencies
install:
    go mod download

# Build the CLI binary
build:
    go build -o bin/muster ./cmd/muster

# Run all tests
test:
    go test ./... -v

# Run tests with coverage
test-coverage:
    go test ./... -v -coverprofile=coverage.out

# Format code
fmt:
    go fmt ./...

# Lint: check code style and formatting
lint:
    golangci-lint run ./...
    gofmt -l .

# Lint-write: run linters and automatically fix issues
lint-write:
    golangci-lint run --fix ./...
    go fmt ./...

# Gate: agentic verification step (tests, lint, build)
gate: build test lint
    @echo "✅ Gate passed"

# Clean build artifacts
clean:
    rm -f bin/muster
    rm -f coverage.out

# Build and run a quick verification
verify: build test fmt lint
    @echo "✅ Build and tests passed"

# Install to GOPATH/bin
install-muster:
    go install ./cmd/muster
    @echo "✅ Installed muster to $(go env GOPATH)/bin/muster"

# Uninstall from GOPATH/bin
uninstall-muster:
    rm -f $(go env GOPATH)/bin/muster
    @echo "✅ Uninstalled muster from $(go env GOPATH)/bin/muster"
