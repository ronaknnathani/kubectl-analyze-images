.PHONY: build clean test test-coverage install install-plugin deps run-mock lint snapshot check help

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

# Build the plugin with version info
build: deps
	go build -ldflags "$(LDFLAGS)" -o kubectl-analyze-images ./cmd/kubectl-analyze-images

# Clean build artifacts
clean:
	rm -f kubectl-analyze-images
	rm -f coverage.out
	rm -rf dist/

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Run linter
lint:
	golangci-lint run ./...

# Run all checks (test + lint)
check: test lint

# Local release test (no publish)
snapshot:
	goreleaser release --snapshot --clean

# Install the plugin
install: build
	mkdir -p ~/.local/bin
	cp kubectl-analyze-images ~/.local/bin/
	chmod +x ~/.local/bin/kubectl-analyze-images

# Install as kubectl plugin
install-plugin: build
	mkdir -p ~/.kube/plugins/analyze-images
	cp kubectl-analyze-images ~/.kube/plugins/analyze-images/
	chmod +x ~/.kube/plugins/analyze-images/kubectl-analyze-images

# Download dependencies
deps:
	go mod tidy

# Run with mock data (for testing without cluster)
run-mock:
	go run cmd/kubectl-analyze-images/main.go analyze --namespace=default

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the plugin with version info"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run golangci-lint"
	@echo "  check          - Run tests and linter"
	@echo "  snapshot       - Build snapshot release (goreleaser)"
	@echo "  install        - Install to ~/.local/bin"
	@echo "  install-plugin - Install as kubectl plugin"
	@echo "  deps           - Download dependencies"
	@echo "  help           - Show this help"
