#!/usr/bin/env just --justfile

# Default recipe displays help
default:
    @just --list

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

# Run linter
lint:
    golangci-lint run ./...

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
