.PHONY: build clean test test-coverage install install-plugin deps run lint snapshot check ci help

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)
GOLANGCI_LINT_VERSION ?= v1.64.8
TOOLS_DIR := $(CURDIR)/.tools
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint/$(GOLANGCI_LINT_VERSION)/golangci-lint

# Build the plugin after dependency, test, and lint checks
build: deps test lint
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
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) run ./...

# Run all checks (test + lint)
check: test lint

# Run the same checks as GitHub Actions
ci: $(GOLANGCI_LINT)
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) run ./...
	go build -o kubectl-analyze-images ./cmd/kubectl-analyze-images
	./kubectl-analyze-images --version

$(GOLANGCI_LINT):
	mkdir -p $(dir $@)
	GOBIN=$(dir $@) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

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

# Run against the current Kubernetes context
run:
	go run ./cmd/kubectl-analyze-images --namespace=default

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the plugin"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run golangci-lint"
	@echo "  check          - Run tests and linter"
	@echo "  ci             - Run the same checks as GitHub Actions"
	@echo "  snapshot       - Build snapshot release (goreleaser)"
	@echo "  install        - Install to ~/.local/bin"
	@echo "  install-plugin - Install as kubectl plugin"
	@echo "  deps           - Download dependencies"
	@echo "  run            - Run against the current Kubernetes context"
	@echo "  help           - Show this help"
