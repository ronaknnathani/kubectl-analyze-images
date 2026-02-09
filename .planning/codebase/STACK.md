# Technology Stack

**Analysis Date:** 2026-02-09

## Languages

**Primary:**
- Go 1.23.0 - Main implementation language for kubectl plugin
- Go 1.23.10 (toolchain) - Build toolchain version

**Secondary:**
- None

## Runtime

**Environment:**
- Go 1.23.0+ required for development and running the plugin
- Kubernetes cluster with kubectl configured for runtime

**Package Manager:**
- Go modules (go mod)
- Lockfile: `go.sum` present

## Frameworks

**Core:**
- Cobra v1.8.0 - CLI command framework and command-line parsing (`cmd/kubectl-analyze-images/main.go`)
- k8s.io/client-go v0.29.0 - Kubernetes client library for API interaction (`internal/cluster/client.go`)

**CLI/Output:**
- github.com/spf13/cobra v1.8.0 - Command structure and flag parsing
- github.com/olekukonko/tablewriter v1.0.7 - Table formatting for output (`internal/reporter/report.go`)
- github.com/briandowns/spinner v1.23.2 - Loading spinner UI feedback (`internal/cluster/client.go`, `internal/analyzer/pod_analyzer.go`)
- github.com/fatih/color v1.15.0 - Colored terminal output (`pkg/types/visualization.go`)

**Kubernetes APIs:**
- k8s.io/api v0.29.0 - Kubernetes API definitions (Pod, Node core types)
- k8s.io/apimachinery v0.29.0 - Kubernetes API machinery and utilities

## Key Dependencies

**Critical:**
- k8s.io/client-go v0.29.0 - Enables cluster communication and pod/node querying via `kubernetes.Clientset`
- github.com/spf13/cobra v1.8.0 - Powers CLI interface and flag handling

**Infrastructure:**
- golang.org/x/oauth2 v0.14.0 - Authentication support for Kubernetes API access
- google.golang.org/protobuf v1.31.0 - Kubernetes API message serialization
- gopkg.in/yaml.v3 v3.0.1 - YAML parsing for kubeconfig files

**Utilities:**
- github.com/google/uuid v1.3.1 - UUID generation (indirect, via k8s.io dependencies)
- golang.org/x/net v0.18.0 - Network utilities for HTTP communication
- golang.org/x/sys v0.14.0 - System-level operations

## Configuration

**Environment:**
- Kubeconfig file loading via `k8s.io/client-go/tools/clientcmd` - Standard Kubernetes configuration
- Context selection supported via `--context` flag (default: current context)
- No explicit environment variables required for basic operation

**Build:**
- `Makefile` orchestrates builds
- Go build target: `cmd/kubectl-analyze-images/main.go`
- Output binary: `kubectl-analyze-images`

**Analysis Configuration:**
- Defined in `pkg/types/analysis.go` - `AnalysisConfig` struct
- Default concurrency: 10 concurrent operations
- Default timeout: 30 seconds per operation
- Cache TTL: 24 hours (configurable)
- Pod page size: 500 pods per page

## Platform Requirements

**Development:**
- Go 1.23.0 or later
- `kubectl` installed and configured with cluster access
- Unix-like environment (Linux/macOS/WSL for Windows)
- Standard development tools (make, git)

**Production/Runtime:**
- Kubernetes cluster v1.29.0 compatible (matching client-go v0.29.0)
- `kubectl` binary installed in PATH
- Valid kubeconfig file (~/.kube/config) with cluster credentials
- Kubernetes API read permissions for:
  - Pods across specified namespaces
  - Nodes and node status
  - Label selectors for pod filtering

**Plugin Deployment:**
- Installs to `~/.kube/plugins/analyze-images/` for kubectl plugin discovery
- Or to `~/.local/bin/` as standalone executable
- Executable permissions required (`chmod +x`)

---

*Stack analysis: 2026-02-09*
