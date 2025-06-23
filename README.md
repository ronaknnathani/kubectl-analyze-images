# kubectl-analyze-images

A kubectl plugin to analyze container images from pods in Kubernetes clusters.

## Phase 1 Features

- Analyze pods from a single Kubernetes cluster
- Extract unique container images
- Generate image size reports
- Support for namespace and label selectors
- Table and JSON output formats

## Installation

### Prerequisites

- Go 1.21 or later
- kubectl configured with access to a Kubernetes cluster

### Build and Install

```bash
# Clone the repository
git clone <repository-url>
cd kubectl-analyze-images

# Download dependencies
make deps

# Build the plugin
make build

# Install as kubectl plugin
make install-plugin
```

### Alternative Installation

```bash
# Install to local bin directory
make install

# Add to PATH if needed
export PATH=$PATH:~/.local/bin
```

## Usage

### Basic Usage

```bash
# Analyze all pods in default namespace
kubectl analyze-images analyze

# Analyze pods in specific namespace
kubectl analyze-images analyze --namespace=production

# Analyze pods with label selector
kubectl analyze-images analyze --namespace=production --selector="app=web"

# Output in JSON format
kubectl analyze-images analyze --output=json
```

### Command Options

- `--namespace, -n`: Target namespace (default: all namespaces)
- `--selector, -l`: Label selector for pods
- `--output, -o`: Output format (table, json) (default: table)

### Example Output

```
Analyzing pods in namespace: production

Image Analysis Summary
=====================
Total Images:	15
Total Size:	2.5 GB
Unique Size:	2.5 GB

Top 25 Images by Size
=====================
Image                           Size        Registry    Tag
----                            ----        --------    ---
nginx:1.21                      133.0 MB    docker.io   1.21
redis:6.2                       110.0 MB    docker.io   6.2
postgres:13                     232.0 MB    docker.io   13
```

## Development

### Project Structure

```
kubectl-analyze-images/
├── cmd/kubectl-analyze-images/  # Main entry point
├── internal/
│   ├── analyzer/               # Pod and image analysis
│   ├── cluster/               # Kubernetes cluster client
│   ├── registry/              # OCI registry client
│   └── reporter/              # Report generation
└── pkg/types/                 # Common types
```

### Building

```bash
# Build the plugin
make build

# Run tests
make test

# Clean build artifacts
make clean
```

### Testing

```bash
# Run with mock data (no cluster required)
make run-mock
```

## Current Limitations (Phase 1)

- Only supports single cluster analysis
- Mock image size data (not real registry queries)
- Basic error handling
- Limited output formats
- No caching system

## Roadmap

### Phase 2: File-Based Pod Analysis
- Support for reading PodList JSON files
- Multiple file input support

### Phase 3: Advanced Image Analysis
- Real OCI registry queries
- Layer analysis and deduplication
- Enhanced reporting

### Phase 4: Reporting and Visualization
- Charts and graphs
- Multiple output formats
- Export capabilities

### Phase 5: Multi-Cluster Support
- Multiple cluster analysis
- Context regex support

### Phase 6: Caching System
- Disk-based image metadata caching
- Cache management commands

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

[Add your license here]
