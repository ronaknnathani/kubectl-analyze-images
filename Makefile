.PHONY: build clean test install

# Build the plugin
build:
	go build -o kubectl-analyze-images cmd/kubectl-analyze-images/main.go

# Clean build artifacts
clean:
	rm -f kubectl-analyze-images

# Run tests
test:
	go test ./...

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
	go mod download
	go mod tidy

# Run with mock data (for testing without cluster)
run-mock:
	go run cmd/kubectl-analyze-images/main.go analyze --namespace=default

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the plugin"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  install       - Install to ~/.local/bin"
	@echo "  install-plugin - Install as kubectl plugin"
	@echo "  deps          - Download dependencies"
	@echo "  run-mock      - Run with mock data"
