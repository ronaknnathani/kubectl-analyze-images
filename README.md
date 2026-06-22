# kubectl-analyze-images

A kubectl plugin that analyzes container image sizes across Kubernetes clusters using node status data. No registry credentials required.

## Features

- Analyze image sizes from node status (no external registry queries needed)
- Show how many pods and containers use each image
- Show how many nodes have each image cached locally
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

# Show full image names without truncation
kubectl analyze-images --wide

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
| `--wide` | | `false` | Show full image names without truncation |
| `--version` | | | Show version information |

### Example output

```
Analyzing images in namespace: All

✓ Found 560 pods across all namespaces (query time: 1.1s)
✓ Found 312 unique images from 12 nodes (query time: 1.2s)
✓ Completed analyzing 312 images (time: 150ms)

Performance Summary
==================
+---------------------+-------+
| Metric              | Value |
+---------------------+-------+
| Node Query Time     | 1.2s  |
| Pod Query Time      | 1.1s  |
| Image Analysis Time | 150ms |
| Total Time          | 1.3s  |
| Images Processed    | 312   |
+---------------------+-------+

Image Analysis Summary
=====================
+---------------+--------+
| Metric                  | Value  |
+-------------------------+--------+
| Pods Scanned            | 560    |
| Nodes Scanned           | 12     |
| Total Images            | 312    |
| Images In Use           | 289    |
| Images Not Used By Pods | 23     |
| Unique Images           | 312    |
| Total Unique Size       | 45 GB  |
+-------------------------+--------+

Image Size Distribution
=======================
   0B-100MB : ████████████████████████████████████████ (95 images, 30%)
 100MB-200MB : ████████████████████████████ (82 images, 26%)
 200MB-300MB : ████████████████ (52 images, 17%)
 300MB-500MB : ████████ (38 images, 12%)
 500MB-  1GB : ████ (28 images, 9%)
   1GB-  2GB : ██ (17 images, 5%)

Top 25 Images by Size and Usage
==============================
+--------------------------------------+-----+------------+------------+--------+-----------------+
| Image                                | Pods| Containers | Namespaces | Size   | Cached On Nodes |
+--------------------------------------+-----+------------+------------+--------+-----------------+
| gcr.io/ml-platform/training-gpu:v2.1 | 4   | 4          | 1          | 1.8 GB | 3               |
| docker.io/nvidia/cuda:12.0-devel     | 0   | 0          | 0          | 1.5 GB | 2               |
| quay.io/prometheus/prometheus:v2.47  | 3   | 3          | 1          | 232 MB | 12              |
+--------------------------------------+-----+------------+------------+--------+-----------------+
```

### JSON output

```bash
kubectl analyze-images -o json | jq '.summary'
```

```json
{
  "podsScanned": 560,
  "nodesScanned": 12,
  "totalImages": 312,
  "imagesInUse": 289,
  "unusedImages": 23,
  "totalSizeBytes": 48318382080,
  "uniqueSizeBytes": 44891258880
}
```

## How it works

The plugin operates in two modes:

1. **All Images Mode** (default): When no namespace or label selector is specified, it queries pods and nodes. Pod data provides usage counts, and node status provides image sizes plus how many nodes have each image cached locally.

2. **Filtered Mode**: When a namespace or label selector is specified, it reports only images used by matching pods, then cross-references node status data to get sizes and cached-on-nodes counts.

Key design choices:

- Uses Kubernetes API pagination for large clusters (1000 items per page)
- Read-only: only needs GET/LIST access to pods and nodes
- No registry credentials required -- all data comes from node status
- Progress spinners on stderr keep stdout clean for piping
- JSON output writes only JSON to stdout; progress stays on stderr

## Requirements

- Kubernetes cluster with kubectl access configured
- RBAC: read access to pods and nodes (list, get)

## Development

```bash
make build          # Build the plugin
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
