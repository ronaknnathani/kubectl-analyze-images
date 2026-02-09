# Codebase Structure

**Analysis Date:** 2026-02-09

## Directory Layout

```
kubectl-analyze-images/
├── cmd/
│   └── kubectl-analyze-images/
│       └── main.go                    # CLI entry point with Cobra command definition
├── internal/
│   ├── analyzer/
│   │   └── pod_analyzer.go           # Orchestrates analysis workflow
│   ├── cluster/
│   │   └── client.go                 # Kubernetes API client and queries
│   └── reporter/
│       └── report.go                 # Report generation (table/JSON)
├── pkg/
│   └── types/
│       ├── image.go                  # Image and ImageAnalysis types
│       ├── pod.go                    # Pod type and conversion
│       ├── analysis.go               # AnalysisConfig and PerformanceMetrics
│       └── visualization.go          # Histogram types and rendering
├── go.mod                            # Go module definition
├── go.sum                            # Dependency checksums
├── Makefile                          # Build and install targets
└── README.md                         # Project documentation
```

## Directory Purposes

**cmd/kubectl-analyze-images/:**
- Purpose: Command-line interface and entry point
- Contains: Single `main.go` file with Cobra command definition, flag parsing, and CLI workflow
- Key files: `cmd/kubectl-analyze-images/main.go`

**internal/analyzer/:**
- Purpose: Business logic layer for image analysis coordination
- Contains: Pod analyzer orchestration, image aggregation logic, registry/tag extraction
- Key files: `internal/analyzer/pod_analyzer.go` (PodAnalyzer struct and AnalyzePods method)

**internal/cluster/:**
- Purpose: Kubernetes cluster interaction and data retrieval
- Contains: Client initialization, pod listing with pagination, node image extraction, image name canonicalization
- Key files: `internal/cluster/client.go` (Client struct, ListPods, GetImageSizesFromNodes, utility functions)

**internal/reporter/:**
- Purpose: Output generation and formatting
- Contains: Report generation logic, table formatting via tablewriter, JSON marshaling, byte formatting utilities
- Key files: `internal/reporter/report.go` (Reporter struct, generateTableReport, generateJSONReport)

**pkg/types/:**
- Purpose: Domain types and data structures used throughout application
- Contains: Image, Pod, ImageAnalysis, AnalysisConfig, PerformanceMetrics, HistogramConfig, histogram generation and rendering
- Key files:
  - `pkg/types/image.go` - Image struct, ImageAnalysis, sorting by size
  - `pkg/types/pod.go` - Pod struct, Kubernetes conversion
  - `pkg/types/analysis.go` - Configuration and performance metrics
  - `pkg/types/visualization.go` - Histogram generation and ASCII rendering

## Key File Locations

**Entry Points:**
- `cmd/kubectl-analyze-images/main.go`: Cobra command setup, flag definitions, calls runAnalyze()

**Configuration:**
- `pkg/types/analysis.go` (AnalysisConfig struct with defaults)
- `pkg/types/visualization.go` (HistogramConfig struct with defaults)
- No external config files - all configuration via CLI flags

**Core Logic:**
- `internal/analyzer/pod_analyzer.go` (AnalyzePods orchestration)
- `internal/cluster/client.go` (cluster data retrieval and processing)
- `pkg/types/image.go` (ImageAnalysis data aggregation methods)

**Testing:**
- Not detected in current structure (no test files present)

## Naming Conventions

**Files:**
- Lowercase with underscores for multi-word names: `pod_analyzer.go`, `client.go`
- Single file per package in most cases (internal packages are small)

**Directories:**
- Lowercase: `cmd`, `internal`, `pkg`, `analyzer`, `cluster`, `reporter`, `types`
- Descriptive names matching responsibility: `analyzer` orchestrates analysis, `cluster` handles K8s API, `reporter` generates output

**Functions:**
- Exported: PascalCase (`NewClient`, `ListPods`, `AnalyzePods`, `GenerateReport`)
- Unexported: camelCase (`runAnalyze`, `generateTableReport`, `extractRegistryAndTag`, `namespaceDisplay`)

**Structs/Types:**
- PascalCase: `Client`, `PodAnalyzer`, `Reporter`, `Image`, `Pod`, `ImageAnalysis`, `PerformanceMetrics`, `HistogramConfig`

**Variables:**
- Short, descriptive camelCase: `analyzer`, `config`, `analysis`, `reporter`, `pods`, `imageSizes`, `perfMetrics`

## Where to Add New Code

**New Feature (e.g., CSV export):**
- Primary code: Add new export method to `internal/reporter/report.go` (e.g., `generateCSVReport()`)
- Add case to `GenerateReport()` switch statement
- Add flag in `cmd/kubectl-analyze-images/main.go`
- Tests: Add corresponding test file (if testing is added to project)

**New Cluster Query (e.g., PVC size analysis):**
- Implementation: Add new method to `Client` struct in `internal/cluster/client.go` (e.g., `GetPVCSizes()`)
- Type definition: Add PVC-related types in `pkg/types/` (new file or existing)
- Integration: Call from `PodAnalyzer` in `internal/analyzer/pod_analyzer.go`

**New Analysis Feature (e.g., image layer analysis):**
- Implementation: Add method to `ImageAnalysis` in `pkg/types/image.go` (e.g., `AnalyzeLayers()`)
- Orchestration: Call from `PodAnalyzer.AnalyzePods()` method
- Reporting: Update reporter to display new analysis results

**Utilities/Helpers:**
- Shared byte formatting: `pkg/types/visualization.go` (formatBytes, formatBytesShort)
- Image parsing: Add to `pkg/types/image.go` (extractRegistryAndTag)
- String utilities: Add to appropriate type file or new `pkg/util/` package

**New Package:**
- Small focused packages under `internal/` (if internal only) or `pkg/` (if exported)
- Example: `internal/cache/` for caching implementation, `pkg/config/` for configuration management

## Special Directories

**internal/**
- Purpose: Private packages not exported outside this module
- Generated: No
- Committed: Yes - contains core business logic

**pkg/**
- Purpose: Public packages that could be imported by other modules
- Generated: No
- Committed: Yes - contains domain types

**cmd/**
- Purpose: Entry point for executable binaries
- Generated: No
- Committed: Yes - contains main function

**.git/**
- Purpose: Git repository metadata
- Generated: Yes (by git)
- Committed: No

## Type Relationships

**Data Flow Hierarchy:**

```
AnalysisConfig ─┐
                ├─→ PodAnalyzer ──→ ImageAnalysis ──→ Reporter ──→ Output
Kubernetes CLI ┘                        │
                                        └─→ PerformanceMetrics
                                        └─→ []Image (with Pod conversion)
```

**Type Dependencies:**

- `PodAnalyzer` depends on `Client` and `AnalysisConfig`
- `Client` returns `[]Pod` and `map[string]int64` (image sizes)
- `PodAnalyzer` converts pods and sizes into `[]Image` within `ImageAnalysis`
- `Reporter` takes `ImageAnalysis` and generates output
- `HistogramData` generated from `ImageAnalysis.GenerateImageSizeHistogram()`

## Package Import Organization

**Observed Pattern (no linting config found):**

1. Standard library imports (fmt, context, time, os, etc.)
2. Kubernetes imports (k8s.io/api, k8s.io/client-go, etc.)
3. Third-party imports (github.com external packages)
4. Internal imports (github.com/ronaknnathani/kubectl-analyze-images/...)

Example from `cmd/kubectl-analyze-images/main.go`:
```go
import (
    "context"    // stdlib
    "fmt"        // stdlib
    "os"         // stdlib

    "github.com/ronaknnathani/kubectl-analyze-images/internal/analyzer"    // internal
    "github.com/ronaknnathani/kubectl-analyze-images/internal/reporter"    // internal
    "github.com/ronaknnathani/kubectl-analyze-images/pkg/types"            // internal
    "github.com/spf13/cobra"                                              // third-party
)
```

## Module Path

**Module:** `github.com/ronaknnathani/kubectl-analyze-images`

**Import Paths:**
- Cluster client: `github.com/ronaknnathani/kubectl-analyze-images/internal/cluster`
- Pod analyzer: `github.com/ronaknnathani/kubectl-analyze-images/internal/analyzer`
- Reporter: `github.com/ronaknnathani/kubectl-analyze-images/internal/reporter`
- Types: `github.com/ronaknnathani/kubectl-analyze-images/pkg/types`

---

*Structure analysis: 2026-02-09*
