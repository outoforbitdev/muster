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
verify: build test
    @echo "✅ Build and tests passed"

# Install to PATH
install: build
    cp bin/muster /usr/local/bin/muster
    @echo "✅ Installed muster to /usr/local/bin/muster"
