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

### From GitHub releases

Download the latest release for your platform from the
[releases page](https://github.com/ronaknnathani/kubectl-analyze-images/releases).

**macOS (Apple Silicon):**

```bash
curl -LO https://github.com/ronaknnathani/kubectl-analyze-images/releases/latest/download/kubectl-analyze-images_1.0.0_darwin_arm64.tar.gz
tar xzf kubectl-analyze-images_1.0.0_darwin_arm64.tar.gz
sudo mv kubectl-analyze-images /usr/local/bin/
```

**macOS (Intel):**

```bash
curl -LO https://github.com/ronaknnathani/kubectl-analyze-images/releases/latest/download/kubectl-analyze-images_1.0.0_darwin_amd64.tar.gz
tar xzf kubectl-analyze-images_1.0.0_darwin_amd64.tar.gz
sudo mv kubectl-analyze-images /usr/local/bin/
```

**Linux (amd64):**

```bash
curl -LO https://github.com/ronaknnathani/kubectl-analyze-images/releases/latest/download/kubectl-analyze-images_1.0.0_linux_amd64.tar.gz
tar xzf kubectl-analyze-images_1.0.0_linux_amd64.tar.gz
sudo mv kubectl-analyze-images /usr/local/bin/
```

**Linux (arm64):**

```bash
curl -LO https://github.com/ronaknnathani/kubectl-analyze-images/releases/latest/download/kubectl-analyze-images_1.0.0_linux_arm64.tar.gz
tar xzf kubectl-analyze-images_1.0.0_linux_arm64.tar.gz
sudo mv kubectl-analyze-images /usr/local/bin/
```

**Windows:** Download the `.zip` for your architecture from the [releases page](https://github.com/ronaknnathani/kubectl-analyze-images/releases) and add the binary to your PATH.

### Via krew

A krew plugin manifest is included at `plugins/analyze-images.yaml`. To install locally:

```bash
kubectl krew install --manifest=plugins/analyze-images.yaml
```

### From source

```bash
git clone https://github.com/ronaknnathani/kubectl-analyze-images.git
cd kubectl-analyze-images
make build
make install
```

Requires Go 1.23+ and golangci-lint.

## Usage

Once installed in your PATH, kubectl automatically discovers the plugin. You can invoke it as either `kubectl analyze-images` or directly as `kubectl-analyze-images`.

```bash
# Analyze all images in the cluster
kubectl analyze-images

# Analyze images in a specific namespace
kubectl analyze-images -n production

# Filter by label selector
kubectl analyze-images -n production -l app=web

# JSON output for scripting
kubectl analyze-images -o json

# Use a specific kubectl context
kubectl analyze-images --context=prod-cluster

# Show top 50 images (default is 25)
kubectl analyze-images --top-images=50

# Disable colored output (useful for piping)
kubectl analyze-images --no-color
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--namespace` | `-n` | (all namespaces) | Target namespace |
| `--selector` | `-l` | | Label selector for pods |
| `--output` | `-o` | `table` | Output format: `table` or `json` |
| `--context` | | (current context) | Kubernetes context to use |
| `--no-color` | | `false` | Disable colored output |
| `--top-images` | | `25` | Number of top images to show |
| `--version` | | | Show version information |

### Example output

```
Analyzing images in namespace: All

✓ Found 312 unique images from 12 nodes (query time: 1.2s)
✓ Completed analyzing 312 images (time: 150ms)

Performance Summary
==================
+---------------------+-------+
| Metric              | Value |
+---------------------+-------+
| Node Query Time     | 1.2s  |
| Image Analysis Time | 150ms |
| Total Time          | 1.4s  |
| Images Processed    | 312   |
+---------------------+-------+

Image Analysis Summary
=====================
+---------------+--------+
| Metric        | Value  |
+---------------+--------+
| Total Images  | 312    |
| Unique Images | 289    |
| Total Size    | 45 GB  |
+---------------+--------+

Image Size Distribution
=======================
   0B-100MB : ████████████████████████████████████████ (95 images, 30%)
 100MB-200MB : ████████████████████████████ (82 images, 26%)
 200MB-300MB : ████████████████ (52 images, 17%)
 300MB-500MB : ████████ (38 images, 12%)
 500MB-  1GB : ████ (28 images, 9%)
   1GB-  2GB : ██ (17 images, 5%)

Top 25 Images by Size
=====================
+------------------------------------------+---------+
| Image                                    | Size    |
+------------------------------------------+---------+
| gcr.io/ml-platform/training-gpu:v2.1     | 1.8 GB  |
| docker.io/nvidia/cuda:12.0-devel         | 1.5 GB  |
| quay.io/prometheus/prometheus:v2.47       | 232 MB  |
| docker.io/library/postgres:15            | 210 MB  |
| docker.io/library/nginx:1.25             | 133 MB  |
| ...                                      | ...     |
+------------------------------------------+---------+
```

### JSON output

```bash
kubectl analyze-images -o json | jq '.summary'
```

```json
{
  "totalImages": 312,
  "totalSize": 48318382080,
  "uniqueSize": 44891258880
}
```

## How it works

The plugin operates in two modes:

1. **All Images Mode** (default): When no namespace or label selector is specified, it queries all nodes in the cluster to get image sizes from `node.status.images`. This provides a complete view of all images cached across the cluster.

2. **Filtered Mode**: When a namespace or label selector is specified, it first queries pods to identify which images are in use, then cross-references with node status data to get the sizes.

Key design choices:

- Uses Kubernetes API pagination for large clusters (1000 items per page)
- Read-only: only needs GET/LIST access to pods and nodes
- No registry credentials required -- all data comes from node status
- Progress spinners on stderr keep stdout clean for piping

## Requirements

- Kubernetes cluster with kubectl access configured
- RBAC: read access to pods and nodes (list, get)

## Development

```bash
make build          # Build (runs lint + test first)
make test           # Run tests
make lint           # Run golangci-lint
make check          # Run tests and linter
make test-coverage  # Run tests with coverage report
make snapshot       # Build snapshot release locally (goreleaser)
```

## Releasing

```bash
# 1. Ensure all checks pass
make check

# 2. Tag the release
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0

# 3. GoReleaser builds and publishes via GitHub Actions
#    See .goreleaser.yaml for build configuration

# 4. Update krew manifest sha256 hashes from release checksums.txt
#    See plugins/analyze-images.yaml
```

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
