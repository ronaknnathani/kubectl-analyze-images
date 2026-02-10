# kubectl-analyze-images

A kubectl plugin that analyzes container image sizes across Kubernetes clusters using node status data. No registry credentials required.

## Features

- Analyze image sizes from node status (no external registry queries needed)
- Histogram visualization of image size distribution
- Filter by namespace and label selector
- Table and JSON output formats
- Top N images by size report
- Performance metrics (query time, analysis time)
- Color-coded output with `--no-color` option
- Multi-cluster support via `--context`

## Installation

### Via krew (recommended)

```bash
kubectl krew install analyze-images
```

### From GitHub releases

Download the latest release for your platform:

```bash
# Linux/macOS (amd64)
curl -LO "https://github.com/ronaknnathani/kubectl-analyze-images/releases/latest/download/kubectl-analyze-images_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/').tar.gz"
tar xzf kubectl-analyze-images_*.tar.gz
sudo mv kubectl-analyze-images /usr/local/bin/
```

For other platforms, download the appropriate archive from the
[releases page](https://github.com/ronaknnathani/kubectl-analyze-images/releases).

### From source

```bash
git clone https://github.com/ronaknnathani/kubectl-analyze-images.git
cd kubectl-analyze-images
make build
make install
```

## Usage

```bash
# Analyze all images in the cluster
kubectl analyze-images

# Analyze images in a specific namespace
kubectl analyze-images --namespace=production

# Filter by label selector
kubectl analyze-images --namespace=production --selector="app=web"

# JSON output for scripting
kubectl analyze-images --output=json

# Use a specific kubectl context
kubectl analyze-images --context=prod-cluster

# Show top 50 images (default is 25)
kubectl analyze-images --top-images=50

# Disable colored output
kubectl analyze-images --no-color
```

## Flags Reference

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--namespace` | `-n` | (all namespaces) | Target namespace |
| `--selector` | `-l` | | Label selector for pods |
| `--output` | `-o` | `table` | Output format: table, json |
| `--context` | | (current context) | Kubernetes context to use |
| `--no-color` | | `false` | Disable colored output |
| `--top-images` | | `25` | Number of top images to show |
| `--version` | | | Show version information |

## Example Output

```
Analyzing images in namespace: All

Found 45 unique images from 12 nodes (query time: 1.2s)
Completed analyzing 45 images (time: 1.5s)

Performance Summary
==================
Node Query Time     1.2s
Image Analysis Time 1.5s
Total Time         1.5s
Images Processed   45

Image Analysis Summary
=====================
Total Images    45
Total Size      2.5 GB

Image Size Distribution
=======================
  100MB-200MB : ████████████████████████████████████████ (15 images, 33%)
  200MB-300MB : ████████████████████████████ (12 images, 27%)
  300MB-400MB : ████████████████ (8 images, 18%)
  400MB-500MB : ████████ (5 images, 11%)
  500MB-600MB : ████ (3 images, 7%)
  600MB-700MB : ██ (2 images, 4%)

Top 25 Images by Size
=====================
Image                           Size
----                            ----
nginx:1.21                      133.0 MB
redis:6.2                       110.0 MB
postgres:13                     232.0 MB
```

## How It Works

The plugin operates in two modes:

1. **All Images Mode** (default): When no namespace or label selector is specified, it queries all nodes in the cluster to get image sizes from `node.Status.Images`. This provides a complete view of all images across the cluster.

2. **Filtered Mode**: When a namespace or label selector is specified, it first queries pods to identify which images are in use, then cross-references with node status data to get the sizes.

Key design choices:

- Uses Kubernetes pagination for large clusters
- Read-only: only needs GET access to pods and nodes
- No registry credentials required -- all data comes from node status

## Requirements

- Kubernetes cluster v1.29+
- kubectl configured with cluster access
- RBAC: read access to pods and nodes (list, get)

## Development

### Prerequisites

- Go 1.23+
- golangci-lint

### Build and Test

```bash
# Build the plugin
make build

# Run tests
make test

# Run linter
make lint

# Run all checks (test + lint)
make check

# Build a snapshot release locally
make snapshot
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for full development guidelines.

## Releasing (for maintainers)

```bash
# 1. Ensure all tests pass
make check

# 2. Tag the release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 3. GoReleaser builds and publishes via GitHub Actions
#    See .goreleaser.yaml for build configuration

# 4. Update krew manifest sha256 hashes from release checksums.txt
#    See plugins/analyze-images.yaml

# 5. Test local krew install
kubectl krew install --manifest=plugins/analyze-images.yaml

# 6. Submit PR to krew-index repository
```

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
