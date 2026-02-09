# Architecture

**Analysis Date:** 2026-02-09

## Pattern Overview

**Overall:** Layered CLI application with clear separation between command handling, cluster interaction, analysis, and reporting.

**Key Characteristics:**
- Command-line interface using Cobra framework
- Three-tier architecture: CLI layer → Business logic layer → External integration layer
- Unidirectional data flow from cluster queries through analysis to reporting
- Dependency injection pattern for cluster client initialization
- Configuration-driven behavior (filters, output formats, display options)

## Layers

**Presentation Layer (CLI & Reporting):**
- Purpose: Handle user interaction and output generation
- Location: `cmd/kubectl-analyze-images/main.go`, `internal/reporter/report.go`
- Contains: Cobra command definitions, flag parsing, report formatting (table/JSON), visualization
- Depends on: Analysis results (`types.ImageAnalysis`), configuration (`types.AnalysisConfig`)
- Used by: End users via kubectl plugin

**Business Logic Layer (Analysis):**
- Purpose: Coordinate pod and image analysis, aggregate data from cluster
- Location: `internal/analyzer/pod_analyzer.go`
- Contains: `PodAnalyzer` struct, orchestration logic, image collection and aggregation
- Depends on: Cluster client, type definitions, performance metrics tracking
- Used by: CLI layer for executing analysis

**Cluster Integration Layer:**
- Purpose: Abstract Kubernetes API interactions, query pods and node image data
- Location: `internal/cluster/client.go`
- Contains: `Client` struct wrapping kubernetes clientset, pod listing, node image queries
- Depends on: Kubernetes client-go library, kubeconfig loading
- Used by: Pod analyzer for cluster data retrieval

**Type/Domain Layer:**
- Purpose: Define core data structures and domain logic
- Location: `pkg/types/`
- Contains: Image, Pod, ImageAnalysis, PerformanceMetrics, HistogramConfig, AnalysisConfig
- Depends on: Kubernetes core types for conversion
- Used by: All other layers

## Data Flow

**Analysis Flow (Primary):**

1. **User Input** → CLI layer receives flags (namespace, selector, format, context)
2. **Cluster Client Init** → Creates authenticated Kubernetes client with specified context
3. **Pod Query** (conditional) → If namespace or selector specified, queries pods to extract image names
4. **Node Query** → Queries all nodes to extract image sizes from `node.Status.Images`
5. **Image Aggregation** → Maps pod images to sizes, marks inaccessible images
6. **Report Generation** → Formats aggregated data (table or JSON with histograms)
7. **Output** → Writes to stdout with progress spinners on stderr

**Key Data Structures Through Flow:**

```
User Flags (namespace, selector, etc.)
    ↓
AnalysisConfig (concurrency, timeout, caching config)
    ↓
Kubernetes Client (authenticated connection)
    ↓
[]Pod (extracted from cluster) + map[string]int64 (image sizes from nodes)
    ↓
[]Image (unified with metadata: name, size, registry, tag, accessibility)
    ↓
ImageAnalysis (images + total size + performance metrics)
    ↓
Report (table/JSON formatted output with histograms)
```

**State Management:**
- Stateless: No in-process state persistence between commands
- Cluster state queried fresh on each run via Kubernetes API
- Performance metrics collected during analysis execution
- All data flows in one direction: cluster → analysis → report

## Key Abstractions

**PodAnalyzer:**
- Purpose: Coordinates analysis workflow, bridges cluster interaction and reporting
- Examples: `internal/analyzer/pod_analyzer.go` (primary), `NewPodAnalyzer()`, `AnalyzePods()`
- Pattern: Facade pattern providing high-level analysis interface while delegating to cluster client

**Cluster Client:**
- Purpose: Encapsulates Kubernetes API interaction, provides pod and node image queries
- Examples: `internal/cluster/client.go`, `NewClient()`, `ListPods()`, `GetImageSizesFromNodes()`
- Pattern: Adapter pattern between Kubernetes client-go and application domain types

**Reporter:**
- Purpose: Encapsulates output generation logic, supports multiple formats
- Examples: `internal/reporter/report.go`, `NewReporter()`, `GenerateReport()`, `generateTableReport()`, `generateJSONReport()`
- Pattern: Strategy pattern with format selection (table vs JSON)

**Image/Pod Domain Types:**
- Purpose: Represent cluster resources in application domain
- Examples: `types.Pod` (name, namespace, images), `types.Image` (name, size, registry, tag, accessibility)
- Pattern: Domain model with minimal logic, used throughout application

## Entry Points

**CLI Entry Point:**
- Location: `cmd/kubectl-analyze-images/main.go` (main function)
- Triggers: `kubectl analyze-images` command with flags
- Responsibilities: Parse flags, instantiate analyzer, invoke analysis, handle errors, exit codes

**Analysis Entry Point:**
- Location: `internal/analyzer/pod_analyzer.go` (AnalyzePods method)
- Triggers: Called from CLI after analyzer instantiation
- Responsibilities: Orchestrate cluster queries, aggregate data, create ImageAnalysis result

**Cluster Query Entry Points:**
- `internal/cluster/client.go` (ListPods, GetImageSizesFromNodes)
- Triggers: Called from PodAnalyzer based on filtering requirements
- Responsibilities: Execute Kubernetes API calls, paginate results, handle errors, collect performance metrics

## Error Handling

**Strategy:** Error propagation with context-wrapped messages

**Patterns:**
- `fmt.Errorf()` with `%w` verb for error wrapping preserves stack context
- Early return on error at each layer
- CLI layer converts top-level errors to stderr output and exit code 1
- Spinner cleanup (defer s.Stop()) ensures proper cleanup even on error

**Error Locations:**
- Kubeconfig loading failure → Client creation fails
- Pod query timeout → Returned to CLI with wrapped error
- Node query failure → No image sizes available, images marked inaccessible
- Invalid output format → Reporter returns error

## Cross-Cutting Concerns

**Logging:**
- Standard `fmt.Printf/Fprintf` for informational messages
- Spinners (progress indicators) written to stderr to avoid mixing with output
- No structured logging framework; simple human-readable progress messages

**Validation:**
- Flags validated by Cobra (type validation, constraints)
- Image names parsed for registry/tag extraction with safe defaults ("unknown" registry, "latest" tag)
- Output format validated in reporter with fallback to error

**Authentication:**
- Handled entirely by kubernetes/client-go kubeconfig loading
- Supports context selection for multi-cluster scenarios
- No explicit auth logic in application code

**Performance:**
- Uses Kubernetes watch cache (`ResourceVersion="0"`) for optimized node/pod queries
- Paginated queries via Kubernetes pager (page size 1000)
- Progress spinners updated every 100 pods / 10 nodes to minimize overhead
- Histogram generation uses single-pass statistics collection

---

*Architecture analysis: 2026-02-09*
