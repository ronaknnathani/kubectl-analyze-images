# kubectl Plugin Architecture Research

**Research Date:** 2026-02-09
**Focus:** Standard project layout and design patterns for krew-distributed kubectl plugins
**Dimension:** Architecture

## Executive Summary

Well-structured kubectl plugins follow golang-standards/project-layout principles with specific kubectl plugin conventions. Key architectural patterns include:

1. **Standard Directory Layout**: `cmd/` for entrypoints, `pkg/` for public APIs, `internal/` for private implementation
2. **Interface-Driven Design**: Abstract Kubernetes clients and external dependencies for testability
3. **Thin CLI Layer**: Keep `main.go` minimal, move business logic to testable packages
4. **Clear Separation of Concerns**: Distinct packages for cluster interaction, business logic, and output formatting

The current project structure is 80% aligned with these standards but needs interface abstraction and test infrastructure.

## Research Questions & Answers

### Q1: What is the standard project layout for Go kubectl plugins?

**Answer:** The canonical layout for kubectl plugins distributed via krew follows golang-standards/project-layout with kubectl-specific conventions:

```
kubectl-plugin-name/
├── cmd/
│   └── kubectl-plugin_name/        # Main entry point (single binary)
│       └── main.go                 # Thin CLI layer, flag parsing only
├── pkg/
│   ├── plugin/                     # Public API for plugin logic
│   │   ├── plugin.go              # Main plugin interface/struct
│   │   └── options.go             # Option structs for configuration
│   ├── kubernetes/                 # Kubernetes client abstractions
│   │   ├── client.go              # Interface for K8s operations
│   │   └── fake.go                # Fake implementation for testing
│   └── output/                     # Output formatting (table, JSON, YAML)
│       ├── printer.go             # Printer interface
│       └── formatters.go          # Concrete implementations
├── internal/                       # Private implementation details
│   ├── analyzer/                  # Business logic (not exposed)
│   └── util/                      # Internal utilities
├── test/
│   ├── integration/               # Integration tests with real/fake K8s
│   └── fixtures/                  # Test data and mock responses
├── .krew.yaml                     # Krew plugin manifest
├── Makefile                       # Build targets
├── go.mod
└── README.md
```

**Key Conventions:**
- Binary name: `kubectl-plugin_name` (underscores, not hyphens in binary)
- Krew invocation: `kubectl plugin-name` (hyphens in command)
- Single entry point in `cmd/`
- Public interfaces in `pkg/` for extensibility and testing
- Private implementation in `internal/`
- Separate test directory with fixtures

**Sources:**
- Standard Go Project Layout: https://github.com/golang-standards/project-layout
- kubectl-tree (ahmetb): Uses `cmd/kubectl-tree/`, `pkg/tree/`, clean separation
- kubectl-images (chenjiandongx): Similar structure with `pkg/core/` for logic
- krew plugin guidelines: Single binary in cmd/, .krew.yaml at root

### Q2: How should interfaces be designed for testability?

**Answer:** Kubectl plugins should abstract external dependencies behind interfaces to enable unit testing without real clusters:

**Pattern 1: Kubernetes Client Interface**
```go
// pkg/kubernetes/client.go
package kubernetes

import (
    corev1 "k8s.io/api/core/v1"
    "k8s.io/client-go/kubernetes"
)

// Interface abstracts Kubernetes operations for testing
type Interface interface {
    ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error)
    ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error)
}

// Client implements Interface using real kubernetes.Clientset
type Client struct {
    clientset kubernetes.Interface  // Note: k8s.io/client-go provides Interface
}

func (c *Client) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
    return c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
}

// pkg/kubernetes/fake.go
type FakeClient struct {
    Pods  *corev1.PodList
    Nodes *corev1.NodeList
    Err   error
}

func (f *FakeClient) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
    if f.Err != nil {
        return nil, f.Err
    }
    return f.Pods, nil
}
```

**Pattern 2: Plugin Options with Validation**
```go
// pkg/plugin/options.go
package plugin

import (
    "github.com/spf13/cobra"
    "k8s.io/cli-runtime/pkg/genericclioptions"
)

type Options struct {
    configFlags *genericclioptions.ConfigFlags  // Standard kubectl config handling
    Namespace   string
    LabelSelector string
    OutputFormat string

    // Injected dependencies (interfaces)
    KubeClient kubernetes.Interface
}

func NewOptions() *Options {
    return &Options{
        configFlags: genericclioptions.NewConfigFlags(true),
    }
}

func (o *Options) Complete(cmd *cobra.Command, args []string) error {
    // Initialize KubeClient from configFlags if not injected (for testing)
    if o.KubeClient == nil {
        config, err := o.configFlags.ToRESTConfig()
        if err != nil {
            return err
        }
        o.KubeClient, err = kubernetes.NewClient(config)
        if err != nil {
            return err
        }
    }
    return nil
}

func (o *Options) Validate() error {
    // Validation logic
    if o.OutputFormat != "table" && o.OutputFormat != "json" {
        return fmt.Errorf("invalid output format: %s", o.OutputFormat)
    }
    return nil
}

func (o *Options) Run() error {
    // Business logic using o.KubeClient (interface)
    plugin := NewPlugin(o.KubeClient)
    return plugin.Analyze(o.Namespace, o.LabelSelector, o.OutputFormat)
}
```

**Pattern 3: Main.go Structure**
```go
// cmd/kubectl-plugin_name/main.go
package main

import (
    "os"
    "github.com/spf13/cobra"
    "github.com/your/plugin/pkg/plugin"
)

func main() {
    cmd := NewCmd()
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func NewCmd() *cobra.Command {
    opts := plugin.NewOptions()

    cmd := &cobra.Command{
        Use:   "kubectl-plugin_name",
        Short: "Description",
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := opts.Complete(cmd, args); err != nil {
                return err
            }
            if err := opts.Validate(); err != nil {
                return err
            }
            return opts.Run()
        },
    }

    opts.AddFlags(cmd.Flags())
    return cmd
}
```

**Testing Pattern:**
```go
// pkg/plugin/plugin_test.go
func TestPluginAnalyze(t *testing.T) {
    fakeClient := &kubernetes.FakeClient{
        Pods: &corev1.PodList{ /* test data */ },
    }

    opts := &plugin.Options{
        KubeClient: fakeClient,  // Inject fake
        Namespace: "default",
        OutputFormat: "json",
    }

    err := opts.Run()
    assert.NoError(t, err)
}
```

**Key Principles:**
- Abstract all external dependencies (K8s API, filesystem, network) behind interfaces
- Use dependency injection in Options/Plugin structs
- Provide fake implementations in `pkg/kubernetes/fake.go`
- Keep main.go untestable but minimal (just wiring)
- Put all business logic in testable packages with injected dependencies

**Sources:**
- k8s.io/cli-runtime patterns (genericclioptions.ConfigFlags)
- kubectl-tree: Uses Options pattern with Complete/Validate/Run
- kubectl code structure: Complete/Validate/Run is kubectl convention

### Q3: What is the proper separation of concerns?

**Answer:** Kubectl plugins should follow a layered architecture with clear boundaries:

**Layer 1: CLI Layer (cmd/)**
- **Responsibility:** Command definition, flag parsing, error handling
- **No business logic:** Just wiring - create Options, call Complete/Validate/Run
- **Minimal code:** main.go should be <100 lines
- **Dependencies:** cobra, plugin.Options

**Layer 2: Plugin Layer (pkg/plugin/)**
- **Responsibility:** Orchestration of business logic, workflow coordination
- **Public API:** Exposed for extension or embedding in other tools
- **No direct K8s API calls:** Use kubernetes.Interface abstraction
- **Testable:** All dependencies injected via Options
- **Dependencies:** kubernetes.Interface, output.Printer

**Layer 3: Kubernetes Layer (pkg/kubernetes/)**
- **Responsibility:** Abstract Kubernetes API operations
- **Interface-based:** Define Interface, provide real Client and FakeClient
- **No business logic:** Pure data retrieval and transformation
- **Single responsibility:** Each method does one K8s operation
- **Dependencies:** k8s.io/client-go

**Layer 4: Output Layer (pkg/output/)**
- **Responsibility:** Format and print results
- **Format-agnostic:** Support table, JSON, YAML via strategy pattern
- **No business logic:** Pure presentation
- **Testable:** Can test formatting without K8s
- **Dependencies:** tablewriter, encoding/json, gopkg.in/yaml.v3

**Layer 5: Internal Layer (internal/)**
- **Responsibility:** Implementation details not exposed publicly
- **Examples:** Complex parsing logic, caching, helper utilities
- **Not importable:** Cannot be imported by external packages
- **Refactorable:** Can change without breaking public API

**Data Flow:**
```
User Input → CLI Layer → Plugin Layer → Kubernetes Layer → K8s API
                                ↓
                         Analysis Logic (internal/)
                                ↓
                         Output Layer → User
```

**Dependency Rules:**
- CLI depends on Plugin
- Plugin depends on Kubernetes + Output + Internal
- Kubernetes depends on k8s.io/client-go only
- Output depends on nothing (just stdlib + formatting libs)
- Internal can depend on anything except public pkg/
- No circular dependencies

**Current Project Gaps:**
- Missing kubernetes.Interface abstraction (direct client-go usage)
- Missing pkg/plugin/ orchestration layer
- Analyzer in internal/ should be in pkg/ if reusable, or internal/ if not
- Reporter should be in pkg/output/ with Printer interface

### Q4: What is the build order for cleanup?

**Answer:** Based on the architecture patterns above, here's the recommended cleanup order to minimize disruption and enable incremental testing:

**Phase 1: Extract and Consolidate Utilities (Low Risk)**
1. **Create `pkg/image/` package**
   - Move `extractRegistryAndTag()` from both locations to `pkg/image/parser.go`
   - Add comprehensive tests for image parsing edge cases
   - Update imports in analyzer and types
   - **Why first:** No structural changes, just deduplication. Enables safe refactoring.

2. **Create `pkg/output/` package**
   - Move `internal/reporter/` to `pkg/output/printer.go`
   - Define `Printer` interface with `Print(analysis *types.ImageAnalysis) error` method
   - Implement `TablePrinter` and `JSONPrinter` structs
   - **Why second:** Isolates output layer, makes it independently testable.

**Phase 2: Introduce Kubernetes Interface (Medium Risk)**
3. **Create `pkg/kubernetes/` package**
   - Define `Interface` with methods: `ListPods()`, `ListNodes()`, `GetNode()`
   - Implement `Client` struct wrapping current `internal/cluster/client.go` logic
   - Create `FakeClient` for testing
   - **Why third:** Enables testing without cluster, but doesn't change behavior yet.

4. **Update `internal/cluster/client.go`**
   - Refactor to implement `kubernetes.Interface`
   - Keep pagination, spinner logic in internal/cluster
   - Make it a wrapper around `pkg/kubernetes/Client`
   - **Why fourth:** Gradual migration, old code still works during transition.

**Phase 3: Restructure Plugin Logic (Higher Risk)**
5. **Create `pkg/plugin/` package**
   - Move `internal/analyzer/pod_analyzer.go` to `pkg/plugin/analyzer.go`
   - Create `Options` struct with Complete/Validate/Run pattern
   - Inject dependencies (KubeClient interface, Printer interface)
   - **Why fifth:** This is the main refactor, but utilities are stable by now.

6. **Simplify `cmd/kubectl-analyze-images/main.go`**
   - Replace direct analyzer calls with `plugin.Options` usage
   - Remove business logic, keep only command definition
   - Use `genericclioptions.ConfigFlags` for kubeconfig handling
   - **Why sixth:** Now that pkg/plugin exists, main.go can be minimal.

**Phase 4: Clean Up Configuration and Tests (Final)**
7. **Fix `pkg/types/analysis.go`**
   - Remove unused fields from `AnalysisConfig` (MaxConcurrency, RetryCount, etc.)
   - Or implement the features they represent (decide based on roadmap)
   - **Why seventh:** Config is widely used, safer to change after refactoring.

8. **Add test infrastructure**
   - Create `test/fixtures/` with sample pod/node YAML
   - Add unit tests for pkg/plugin, pkg/kubernetes, pkg/output
   - Add integration test with FakeClient
   - **Why last:** Tests validate all previous refactoring work.

**Validation After Each Phase:**
- Run `make build` to ensure compilation
- Run `kubectl analyze-images` against test cluster to ensure behavior unchanged
- Check that all imports resolve correctly
- No functional changes until Phase 3

**Critical Path Dependencies:**
```
Phase 1 (utils) ← Phase 2 (interfaces) ← Phase 3 (plugin refactor) ← Phase 4 (cleanup)
     ↓                    ↓                        ↓                          ↓
  Test utils       Test interfaces          Test plugin                Integration tests
```

**Risk Mitigation:**
- Git branch for each phase
- Commit after each step within phase
- Keep old code commented out initially, delete only after validation
- Defer breaking changes (removing unused config) until end

## Pattern Library

### Pattern 1: Client-Go Integration

**Problem:** How to properly initialize Kubernetes client with support for kubeconfig, context switching, and in-cluster auth?

**Solution:** Use k8s.io/cli-runtime/pkg/genericclioptions for standard kubectl behavior:

```go
// pkg/plugin/options.go
import (
    "k8s.io/cli-runtime/pkg/genericclioptions"
    "k8s.io/client-go/kubernetes"
)

type Options struct {
    configFlags *genericclioptions.ConfigFlags
    // ... other fields
}

func NewOptions() *Options {
    return &Options{
        configFlags: genericclioptions.NewConfigFlags(true), // true = add --namespace flag
    }
}

func (o *Options) AddFlags(flags *pflag.FlagSet) {
    o.configFlags.AddFlags(flags)  // Adds --kubeconfig, --context, --namespace, etc.
}

func (o *Options) Complete(cmd *cobra.Command, args []string) error {
    // Get REST config from flags
    restConfig, err := o.configFlags.ToRESTConfig()
    if err != nil {
        return err
    }

    // Create clientset
    clientset, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return err
    }

    o.KubeClient = &kubernetes.Client{Clientset: clientset}
    return nil
}
```

**Benefits:**
- Automatic support for --kubeconfig, --context, --cluster, --user flags
- In-cluster authentication works automatically
- Consistent with kubectl behavior
- Users familiar with kubectl flags immediately understand plugin

**Current Project Gap:** Using manual kubeconfig loading in `internal/cluster/client.go`. Should adopt genericclioptions.

### Pattern 2: Output Formatting Strategy

**Problem:** Support multiple output formats (table, JSON, YAML) without if/else chains in business logic.

**Solution:** Strategy pattern with Printer interface:

```go
// pkg/output/printer.go
package output

type Printer interface {
    Print(data interface{}) error
}

type Format string

const (
    FormatTable Format = "table"
    FormatJSON  Format = "json"
    FormatYAML  Format = "yaml"
)

func NewPrinter(format Format, opts PrinterOptions) Printer {
    switch format {
    case FormatJSON:
        return &JSONPrinter{Writer: opts.Writer}
    case FormatYAML:
        return &YAMLPrinter{Writer: opts.Writer}
    default:
        return &TablePrinter{Writer: opts.Writer, NoColor: opts.NoColor}
    }
}

// pkg/output/table.go
type TablePrinter struct {
    Writer  io.Writer
    NoColor bool
}

func (p *TablePrinter) Print(data interface{}) error {
    analysis := data.(*types.ImageAnalysis)
    // Table rendering logic
    return nil
}

// pkg/output/json.go
type JSONPrinter struct {
    Writer io.Writer
}

func (p *JSONPrinter) Print(data interface{}) error {
    return json.NewEncoder(p.Writer).Encode(data)
}
```

**Usage in Plugin:**
```go
func (o *Options) Run() error {
    analysis, err := o.analyze()
    if err != nil {
        return err
    }

    printer := output.NewPrinter(o.OutputFormat, output.PrinterOptions{
        Writer: os.Stdout,
        NoColor: o.NoColor,
    })
    return printer.Print(analysis)
}
```

**Benefits:**
- Easy to add new formats without changing plugin logic
- Each printer is independently testable
- Can test with bytes.Buffer instead of stdout
- No if/else chains for format selection

**Current Project Gap:** Reporter uses switch statement in GenerateReport(). Should use strategy pattern with Printer interface.

### Pattern 3: Error Handling and User Feedback

**Problem:** How to provide helpful error messages and progress feedback without cluttering business logic?

**Solution:** Use k8s.io/cli-runtime/pkg/genericiooptions for consistent I/O streams:

```go
// pkg/plugin/options.go
import (
    "k8s.io/cli-runtime/pkg/genericclioptions"
)

type Options struct {
    genericclioptions.IOStreams  // Provides In, Out, ErrOut
    // ... other fields
}

func NewOptions(streams genericclioptions.IOStreams) *Options {
    return &Options{
        IOStreams: streams,
    }
}

func (o *Options) Run() error {
    fmt.Fprintf(o.Out, "Analyzing images in namespace: %s\n", o.Namespace)

    analysis, err := o.analyze()
    if err != nil {
        // Error to ErrOut
        fmt.Fprintf(o.ErrOut, "Error: failed to analyze: %v\n", err)
        return err
    }

    return o.printer.Print(analysis)
}
```

**Main.go wiring:**
```go
func NewCmd() *cobra.Command {
    streams := genericclioptions.IOStreams{
        In:     os.Stdin,
        Out:    os.Stdout,
        ErrOut: os.Stderr,
    }

    opts := plugin.NewOptions(streams)
    // ...
}
```

**Testing:**
```go
func TestRun(t *testing.T) {
    var outBuf, errBuf bytes.Buffer
    streams := genericclioptions.IOStreams{
        Out:    &outBuf,
        ErrOut: &errBuf,
    }

    opts := plugin.NewOptions(streams)
    err := opts.Run()

    assert.Contains(t, outBuf.String(), "Analyzing images")
}
```

**Benefits:**
- Testable output (no direct stdout/stderr writes)
- Consistent with kubectl I/O handling
- Separate normal output from errors
- Can capture and assert on output in tests

**Current Project Gap:** Direct fmt.Printf to stdout, os.Stderr writes. Should use IOStreams.

### Pattern 4: Progress Indicators

**Problem:** Show progress for long-running operations without blocking or cluttering code.

**Solution:** Use spinner pattern with defer for cleanup:

```go
// internal/util/spinner.go (or pkg/util if public)
package util

import (
    "io"
    "time"
    "github.com/briandowns/spinner"
)

type Spinner struct {
    s *spinner.Spinner
}

func NewSpinner(w io.Writer, message string) *Spinner {
    s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(w))
    s.Suffix = " " + message
    return &Spinner{s: s}
}

func (s *Spinner) Start() {
    s.s.Start()
}

func (s *Spinner) Stop() {
    s.s.Stop()
}

func (s *Spinner) UpdateMessage(msg string) {
    s.s.Suffix = " " + msg
}
```

**Usage:**
```go
func (o *Options) Run() error {
    spin := util.NewSpinner(o.ErrOut, "Fetching pods...")
    spin.Start()
    defer spin.Stop()  // Always stop, even on error

    pods, err := o.KubeClient.ListPods(...)
    if err != nil {
        return err  // defer stops spinner
    }

    spin.UpdateMessage(fmt.Sprintf("Analyzing %d images...", len(pods)))
    // ... more work

    return nil
}
```

**Benefits:**
- Spinner always cleaned up via defer
- No manual Stop() calls before each return
- Consistent pattern across all long operations
- Testable with NoOpSpinner in tests

**Current Project Gap:** Manual `s.Stop()` calls before errors. Should use defer pattern.

### Pattern 5: Testable Business Logic

**Problem:** How to test business logic that depends on Kubernetes API without real cluster?

**Solution:** Use table-driven tests with fake client:

```go
// pkg/kubernetes/fake.go
package kubernetes

type FakeClient struct {
    Pods      *corev1.PodList
    Nodes     *corev1.NodeList
    ListError error
}

func (f *FakeClient) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
    if f.ListError != nil {
        return nil, f.ListError
    }
    return f.Pods, nil
}

// pkg/plugin/plugin_test.go
func TestAnalyze(t *testing.T) {
    tests := []struct {
        name      string
        pods      *corev1.PodList
        nodes     *corev1.NodeList
        namespace string
        wantErr   bool
        wantImages int
    }{
        {
            name: "single pod with one image",
            pods: &corev1.PodList{
                Items: []corev1.Pod{
                    {
                        Spec: corev1.PodSpec{
                            Containers: []corev1.Container{
                                {Image: "nginx:1.21"},
                            },
                        },
                    },
                },
            },
            nodes: &corev1.NodeList{ /* ... */ },
            namespace: "default",
            wantErr: false,
            wantImages: 1,
        },
        {
            name: "api error",
            pods: nil,
            namespace: "default",
            wantErr: true,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fakeClient := &kubernetes.FakeClient{
                Pods: tt.pods,
                Nodes: tt.nodes,
            }

            plugin := NewPlugin(fakeClient)
            analysis, err := plugin.Analyze(tt.namespace, "")

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.wantImages, len(analysis.Images))
        })
    }
}
```

**Benefits:**
- Fast tests (no real K8s API calls)
- Reproducible (no flaky network issues)
- Easy to test error cases
- Can test with fixture data from real clusters

**Current Project Gap:** No tests at all. Need to add kubernetes.Interface, FakeClient, and test suite.

## Architecture Decision Records

### ADR-1: Use golang-standards/project-layout

**Context:** Need consistent directory structure for Go kubectl plugin that's familiar to contributors and follows community standards.

**Decision:** Adopt golang-standards/project-layout with kubectl-specific modifications:
- `cmd/` for main entry point
- `pkg/` for public, importable packages
- `internal/` for private implementation
- `test/` for integration tests and fixtures

**Rationale:**
- De facto standard in Go community
- Clear separation between public API and implementation
- Prevents accidental exposure of internal packages
- Kubectl community follows similar patterns (kubectl-tree, kubectl-images)

**Consequences:**
- Current `internal/` packages need review: analyzer might belong in `pkg/plugin/`
- Need to create `pkg/kubernetes/` for client abstraction
- Reporter should move to `pkg/output/`

**Status:** Accepted for implementation

### ADR-2: Abstract Kubernetes API Behind Interface

**Context:** Need testability without real Kubernetes cluster. Current code directly uses client-go types, making unit tests impossible without complex mocking.

**Decision:** Create `pkg/kubernetes/Interface` that abstracts all Kubernetes operations. Provide both real `Client` implementation and `FakeClient` for testing.

**Rationale:**
- Enables fast unit tests without K8s cluster
- Allows table-driven tests with fixture data
- Follows dependency inversion principle
- Common pattern in kubectl plugins (kubectl-tree, k8s.io/kubectl code)

**Consequences:**
- Requires refactoring `internal/cluster/client.go` to implement interface
- All K8s calls must go through interface (no direct clientset access)
- Test coverage becomes practical (can test without cluster)

**Status:** Accepted for Phase 2

### ADR-3: Use Complete/Validate/Run Pattern

**Context:** Need consistent command execution flow that's testable and follows kubectl conventions.

**Decision:** Adopt kubectl's Complete/Validate/Run pattern:
- `Complete()`: Initialize dependencies, default values
- `Validate()`: Check flag combinations, required values
- `Run()`: Execute business logic

**Rationale:**
- Standard pattern in kubectl codebase
- Separates validation from execution
- Makes testing easier (can test validation without execution)
- Familiar to kubectl contributors

**Consequences:**
- Need to create `pkg/plugin/options.go` with these methods
- Current `main.go` needs refactoring to call pattern
- All flags become fields on Options struct

**Status:** Accepted for Phase 3

### ADR-4: Use Strategy Pattern for Output Formatting

**Context:** Need to support multiple output formats (table, JSON, YAML) and make it easy to add new formats. Current switch statement in reporter is not extensible.

**Decision:** Implement strategy pattern with `Printer` interface in `pkg/output/`:
- Define `Printer` interface with `Print(data interface{}) error` method
- Implement `TablePrinter`, `JSONPrinter`, `YAMLPrinter` structs
- Factory function `NewPrinter(format Format) Printer` to create appropriate printer

**Rationale:**
- Open/closed principle: open for extension, closed for modification
- Each printer independently testable
- Easy to add new formats without touching existing code
- Can inject test Writer instead of stdout

**Consequences:**
- Reporter needs refactoring to use Printer interface
- Each format becomes separate file (table.go, json.go, yaml.go)
- Plugin code simplified (no format switching logic)

**Status:** Accepted for Phase 1

### ADR-5: Keep Image Parsing in pkg/image/

**Context:** Image name parsing logic duplicated in two places, needs consolidation. Decision needed: where to put shared parsing utilities?

**Decision:** Create `pkg/image/parser.go` with public functions for image name parsing, registry extraction, tag parsing.

**Rationale:**
- Image parsing is reusable utility, belongs in `pkg/` not `internal/`
- Other tools might want to import image parsing logic
- Makes function testable in isolation
- Single source of truth for parsing logic

**Consequences:**
- Delete duplicated code from analyzer and types
- Update imports
- Need comprehensive tests for edge cases (registries with ports, multi-part paths, digests)

**Status:** Accepted for Phase 1

## Implementation Roadmap

### Phase 1: Utilities and Output (Week 1)

**Goals:**
- Eliminate code duplication
- Make output layer independently testable
- No structural changes to main business logic

**Tasks:**
1. Create `pkg/image/parser.go`
   - Extract `parseImageReference(image string) (registry, name, tag string, err error)`
   - Add tests for edge cases (see CONCERNS.md for bugs)
   - Update analyzer and types to use shared function
   - Verify: `make build && go test ./pkg/image/...`

2. Create `pkg/output/` package
   - Define `Printer` interface
   - Move table rendering to `table.go`
   - Move JSON rendering to `json.go`
   - Add tests for each printer
   - Verify: `make build && go test ./pkg/output/...`

3. Update `internal/reporter/` to use `pkg/output/`
   - Replace switch statement with Printer usage
   - Inject printer via constructor
   - Verify: `kubectl analyze-images` still works

**Success Criteria:**
- All tests pass
- Binary builds successfully
- Manual testing shows identical output
- No duplicated code

**Risk:** Low - purely internal refactoring

### Phase 2: Kubernetes Interface (Week 2)

**Goals:**
- Enable testing without cluster
- Abstract Kubernetes API dependencies
- Maintain backward compatibility

**Tasks:**
1. Create `pkg/kubernetes/interface.go`
   - Define `Interface` with methods: `ListPods()`, `ListNodes()`, `GetNode()`
   - Document each method with parameters and return types
   - Add context.Context to all methods
   - Verify: Interface compiles

2. Create `pkg/kubernetes/client.go`
   - Implement `Client` struct wrapping k8s.io/client-go
   - Move existing cluster client logic
   - Keep pagination, timing logic
   - Verify: `make build`

3. Create `pkg/kubernetes/fake.go`
   - Implement `FakeClient` with in-memory data
   - Add helper methods to create test pods/nodes
   - Verify: Can instantiate FakeClient

4. Update `internal/cluster/client.go`
   - Refactor to use `kubernetes.Interface` internally
   - Keep spinner and user feedback logic here
   - Update imports and method calls
   - Verify: `kubectl analyze-images` still works

5. Add basic tests
   - Test real Client initialization (no K8s calls)
   - Test FakeClient returns expected data
   - Verify: `go test ./pkg/kubernetes/...`

**Success Criteria:**
- kubernetes.Interface defined and implemented
- FakeClient allows testing without cluster
- Existing functionality unchanged
- Tests pass

**Risk:** Medium - touching cluster interaction code

### Phase 3: Plugin Layer (Week 3)

**Goals:**
- Establish plugin orchestration layer
- Implement Complete/Validate/Run pattern
- Move business logic to testable packages

**Tasks:**
1. Create `pkg/plugin/options.go`
   - Define `Options` struct with all flags as fields
   - Add `configFlags *genericclioptions.ConfigFlags`
   - Implement `Complete(cmd, args)` method
   - Implement `Validate()` method
   - Verify: Compiles

2. Create `pkg/plugin/plugin.go`
   - Move analyzer logic from `internal/analyzer/pod_analyzer.go`
   - Rename to `Plugin` struct
   - Inject `kubernetes.Interface` dependency
   - Implement `Analyze()` method
   - Verify: Compiles

3. Implement `Options.Run()` method
   - Create Plugin with injected dependencies
   - Call Plugin.Analyze()
   - Call Printer.Print()
   - Verify: Compiles

4. Refactor `cmd/kubectl-analyze-images/main.go`
   - Replace direct analyzer calls with Options usage
   - Remove business logic (only command definition)
   - Use `genericclioptions.ConfigFlags`
   - Add `RunE: func() { opts.Complete(); opts.Validate(); opts.Run() }`
   - Verify: `make build && kubectl analyze-images` works

5. Add plugin tests
   - Test with FakeClient
   - Test error cases (empty namespace, no pods, etc.)
   - Test different flag combinations
   - Verify: `go test ./pkg/plugin/...`

**Success Criteria:**
- main.go is <100 lines
- Plugin logic fully testable with FakeClient
- All existing features work
- Tests cover main workflows

**Risk:** High - major refactoring, but phases 1-2 provide safety net

### Phase 4: Configuration and Testing (Week 4)

**Goals:**
- Clean up unused configuration
- Establish comprehensive test coverage
- Document architecture

**Tasks:**
1. Fix `pkg/types/analysis.go`
   - Review AnalysisConfig fields
   - Remove unused: MaxConcurrency, RetryCount, EnableCaching (or implement)
   - Keep only actively used configuration
   - Verify: No compilation errors

2. Create `test/fixtures/`
   - Add sample pod YAML from real clusters
   - Add sample node YAML with image lists
   - Create helper functions to load fixtures
   - Verify: Can load fixtures in tests

3. Add integration tests
   - Create `test/integration/analyze_test.go`
   - Test full workflow with FakeClient and fixtures
   - Test all output formats
   - Test error scenarios
   - Verify: `go test ./test/integration/...`

4. Add unit tests for remaining packages
   - Test image parsing edge cases (pkg/image)
   - Test histogram generation (pkg/types)
   - Test formatters (pkg/output)
   - Target: >80% coverage
   - Verify: `go test -cover ./...`

5. Document architecture
   - Update README.md with new structure
   - Add ARCHITECTURE.md with diagrams
   - Document testing approach
   - Add CONTRIBUTING.md
   - Verify: Documentation complete

**Success Criteria:**
- No unused configuration fields
- Test coverage >80%
- All edge cases tested
- Documentation complete

**Risk:** Low - additive work

## Reference Architectures

### Example 1: kubectl-tree by ahmetb

**Repository:** https://github.com/ahmetb/kubectl-tree

**Structure:**
```
kubectl-tree/
├── cmd/
│   └── kubectl-tree/
│       └── main.go           # 150 lines, minimal logic
├── pkg/
│   └── tree/
│       ├── tree.go           # Core tree building logic
│       └── tree_test.go      # Unit tests
├── .krew.yaml                # Krew plugin manifest
├── Makefile
└── README.md
```

**Key Patterns:**
- Very simple structure (single pkg/tree package)
- Business logic in pkg/tree, not cmd/
- Uses client-go dynamic client directly (not abstracted)
- Minimal dependencies (no unnecessary libraries)

**Applicable Lessons:**
- Keep package structure simple if plugin is small
- Put all business logic outside cmd/
- Don't over-engineer if <1000 lines of code

**Differences:**
- Our plugin needs more structure (multiple concerns: cluster, analysis, output)
- kubectl-tree is read-only display, we do analysis and aggregation

### Example 2: kubectl-images by chenjiandongx

**Repository:** https://github.com/chenjiandongx/kubectl-images

**Structure:**
```
kubectl-images/
├── cmd/
│   └── kubectl-images/
│       └── main.go           # Cobra command setup
├── pkg/
│   ├── core/
│   │   └── core.go          # Core analysis logic
│   ├── k8s/
│   │   └── client.go        # Kubernetes client wrapper
│   └── printer/
│       ├── json.go
│       └── table.go
├── test/
│   └── testdata/
├── .krew.yaml
├── Makefile
└── README.md
```

**Key Patterns:**
- Clear separation: k8s client, core logic, output printing
- Each concern in separate pkg/ subdirectory
- Testdata in test/ directory
- Uses cobra for CLI

**Applicable Lessons:**
- Separate concerns into focused packages
- Printer pattern for output formats
- Test fixtures in dedicated directory

**Differences:**
- Doesn't use kubernetes.Interface abstraction (direct client-go)
- Simpler than our needs (no histogram, fewer metrics)

### Example 3: kubectl exec-as (from k8s.io/kubectl patterns)

**Pattern Source:** Kubernetes kubectl codebase conventions

**Structure:**
```
cmd/plugin/
└── main.go                   # Only creates command

pkg/cmd/plugin/
├── plugin.go                # RunE function implementation
├── options.go               # Options struct with Complete/Validate/Run
└── plugin_test.go           # Tests with fake client
```

**Key Patterns:**
- Complete/Validate/Run separation
- Options struct holds all state and dependencies
- Use genericclioptions.ConfigFlags for kubeconfig
- Use genericclioptions.IOStreams for I/O
- Business logic injected via interfaces

**Applicable Lessons:**
- This is the "official" kubectl pattern
- Options pattern makes testing straightforward
- IOStreams enable testable output
- Complete() initializes dependencies, Validate() checks state, Run() executes

**Differences:**
- kubectl internal code has more structure than needed for plugin
- Can simplify slightly for plugin context

### Example 4: krew Plugin Guidelines

**Source:** https://krew.sigs.k8s.io/docs/developer-guide/

**Requirements:**
- Binary name: `kubectl-plugin_name` (underscores)
- Single executable (no dependencies)
- `.krew.yaml` manifest at repository root
- Cross-platform builds (linux, darwin, windows)
- Installation via `tar.gz` or `zip`

**Manifest Structure:**
```yaml
apiVersion: krew.googlecontainertools.github.com/v1alpha2
kind: Plugin
metadata:
  name: analyze-images
spec:
  version: v0.1.0
  homepage: https://github.com/ronaknnathani/kubectl-analyze-images
  shortDescription: Analyze container images in Kubernetes clusters
  description: |
    Analyzes container images from pods, extracts image sizes from node status,
    and generates reports with histograms and statistics.
  platforms:
  - selector:
      matchLabels:
        os: linux
        arch: amd64
    uri: https://github.com/ronaknnathani/kubectl-analyze-images/releases/download/v0.1.0/kubectl-analyze-images-linux-amd64.tar.gz
    sha256: "..."
    bin: kubectl-analyze-images
  # ... darwin, windows platforms
```

**Applicable Lessons:**
- Need .krew.yaml for distribution
- Must support multiple platforms
- Release process: GitHub releases with archives
- Binary must be self-contained (static linking)

## Anti-Patterns to Avoid

### Anti-Pattern 1: Business Logic in main.go

**Problem:**
```go
// cmd/kubectl-plugin/main.go - BAD
func main() {
    namespace := flag.String("namespace", "", "...")
    flag.Parse()

    // Business logic directly in main
    config, _ := rest.InClusterConfig()
    clientset, _ := kubernetes.NewForConfig(config)
    pods, _ := clientset.CoreV1().Pods(*namespace).List(...)

    for _, pod := range pods.Items {
        // Analysis logic here...
        fmt.Printf("%s: %s\n", pod.Name, pod.Status.Phase)
    }
}
```

**Why Bad:**
- Untestable (main() can't be called from tests)
- No error handling possible
- Mixes CLI concerns with business logic
- Can't reuse logic in other contexts

**Correct Approach:**
```go
// cmd/kubectl-plugin/main.go - GOOD
func main() {
    cmd := plugin.NewCommand()
    if err := cmd.Execute(); err != nil {
        os.Exit(1)
    }
}

// pkg/plugin/plugin.go
func (o *Options) Run() error {
    // Testable business logic here
    pods, err := o.KubeClient.ListPods(...)
    if err != nil {
        return fmt.Errorf("failed to list pods: %w", err)
    }
    // ... analysis
    return o.Printer.Print(analysis)
}
```

### Anti-Pattern 2: Direct client-go Usage Without Abstraction

**Problem:**
```go
// internal/analyzer/analyzer.go - BAD
type Analyzer struct {
    clientset kubernetes.Interface  // Direct client-go type
}

func (a *Analyzer) Analyze() error {
    pods, err := a.clientset.CoreV1().Pods("").List(...)
    // ...
}
```

**Why Bad:**
- Can't test without real Kubernetes cluster
- Can't use fake data
- Tightly coupled to client-go implementation
- Hard to mock for unit tests

**Correct Approach:**
```go
// pkg/kubernetes/interface.go - GOOD
type Interface interface {
    ListPods(ctx context.Context, namespace string, opts ListOptions) (*PodList, error)
}

// pkg/plugin/plugin.go
type Plugin struct {
    kubeClient kubernetes.Interface  // Our interface, not client-go
}

// pkg/kubernetes/fake.go
type FakeClient struct {
    Pods *PodList
}

func (f *FakeClient) ListPods(...) (*PodList, error) {
    return f.Pods, nil  // Return test data
}
```

### Anti-Pattern 3: God Struct with Too Many Responsibilities

**Problem:**
```go
// BAD - Single struct does everything
type Plugin struct {
    clientset kubernetes.Interface
    namespace string
    output    string
    noColor   bool
}

func (p *Plugin) Run() error {
    // Fetch from K8s
    pods, _ := p.clientset.CoreV1().Pods(p.namespace).List(...)

    // Analyze
    analysis := p.analyzePods(pods)

    // Format output
    if p.output == "json" {
        json.Marshal(analysis)
    } else {
        // Table formatting
    }

    // Print
    fmt.Println(...)

    return nil
}
```

**Why Bad:**
- Single Responsibility Principle violation
- Can't test K8s interaction separate from output
- Can't reuse analysis logic with different output
- Hard to extend (adding new output format requires changing Plugin)

**Correct Approach:**
```go
// GOOD - Separated concerns
// pkg/kubernetes/client.go
type Client struct {
    clientset kubernetes.Interface
}

func (c *Client) ListPods(...) (*PodList, error) {
    return c.clientset.CoreV1().Pods(namespace).List(...)
}

// pkg/plugin/analyzer.go
type Analyzer struct {
    kubeClient kubernetes.Interface
}

func (a *Analyzer) Analyze(namespace string) (*Analysis, error) {
    pods, err := a.kubeClient.ListPods(...)
    // Pure analysis logic
    return analysis, nil
}

// pkg/output/printer.go
type Printer interface {
    Print(data interface{}) error
}

// pkg/plugin/options.go
func (o *Options) Run() error {
    analysis, err := o.analyzer.Analyze(o.namespace)
    if err != nil {
        return err
    }
    return o.printer.Print(analysis)
}
```

### Anti-Pattern 4: Hardcoded Output to stdout/stderr

**Problem:**
```go
// BAD - Direct stdout writes
func (p *Plugin) PrintResults(analysis Analysis) {
    fmt.Println("Results:")  // Can't test, can't redirect
    for _, item := range analysis.Items {
        fmt.Printf("%s: %d\n", item.Name, item.Count)
    }
}
```

**Why Bad:**
- Can't capture output in tests
- Can't redirect to file or buffer
- Violates dependency inversion
- Hard to test output formatting

**Correct Approach:**
```go
// GOOD - Injected writer
type Printer struct {
    Out io.Writer  // Injected, not hardcoded
}

func (p *Printer) Print(analysis Analysis) error {
    fmt.Fprintf(p.Out, "Results:\n")  // Write to injected stream
    for _, item := range analysis.Items {
        fmt.Fprintf(p.Out, "%s: %d\n", item.Name, item.Count)
    }
    return nil
}

// Test
func TestPrint(t *testing.T) {
    var buf bytes.Buffer
    printer := &Printer{Out: &buf}  // Inject buffer

    printer.Print(testAnalysis)

    assert.Contains(t, buf.String(), "Results:")  // Test output
}
```

### Anti-Pattern 5: Premature Abstraction

**Problem:**
```go
// BAD - Over-engineered for simple plugin
type ImageRepository interface {
    GetImage(ref string) (Image, error)
    ListImages() ([]Image, error)
}

type ImageFactory interface {
    CreateImage(name, tag string) Image
}

type ImageValidator interface {
    Validate(img Image) error
}

type ImageTransformer interface {
    Transform(img Image) Image
}

// 10 interfaces for plugin that just lists images...
```

**Why Bad:**
- Over-engineered for simple requirements
- Adds complexity without benefit
- Hard to understand and maintain
- YAGNI violation (You Ain't Gonna Need It)

**Correct Approach:**
```go
// GOOD - Simple, sufficient abstraction
type KubeClient interface {
    ListPods(ctx context.Context, namespace string) ([]Pod, error)
    ListNodes(ctx context.Context) ([]Node, error)
}

// Only abstract what you need to test or swap
// Don't create interfaces "just in case"
```

**Rule of Thumb:** Create interface when you have:
1. Multiple implementations (real + fake)
2. Need for testing with mocks
3. Genuine need to swap implementations

Don't create interface if:
- Only one implementation exists
- No testing benefit
- No swapping needed

## Application to Current Project

### Current State Assessment

**Strengths:**
- Clear directory separation (cmd, internal, pkg)
- Uses cobra for CLI (standard)
- Separate packages for concerns (analyzer, cluster, reporter)
- Types in pkg/ (good for reuse)

**Gaps:**
- No kubernetes.Interface abstraction
- Business logic in internal/analyzer, not pkg/plugin
- Reporter doesn't use Printer interface/strategy pattern
- No Complete/Validate/Run pattern
- No test infrastructure
- main.go has business logic (not just wiring)

### Recommended Changes

**Priority 1: Enable Testing (Weeks 1-2)**
1. Create `pkg/kubernetes/Interface` and `FakeClient`
2. Refactor `internal/cluster/client.go` to use interface
3. Add basic tests for image parsing (move to pkg/image first)

**Priority 2: Restructure Plugin Logic (Week 3)**
1. Create `pkg/plugin/` with Options and Plugin structs
2. Move analyzer logic to pkg/plugin
3. Implement Complete/Validate/Run pattern
4. Simplify main.go to <100 lines

**Priority 3: Output Layer (Week 3-4)**
1. Create `pkg/output/` with Printer interface
2. Implement TablePrinter, JSONPrinter
3. Refactor reporter to use strategy pattern

**Priority 4: Testing and Documentation (Week 4)**
1. Add test fixtures (sample pods, nodes)
2. Write integration tests with FakeClient
3. Achieve >80% test coverage
4. Update documentation

### Migration Path

**Step 1: Add Tests Without Changing Structure**
- Create kubernetes.Interface
- Implement FakeClient
- Write tests for existing code using FakeClient
- This proves tests are possible, builds confidence

**Step 2: Refactor Incrementally**
- Move one package at a time (start with output)
- Keep old code working during transition
- Validate at each step with existing tests

**Step 3: Complete Restructure**
- Create pkg/plugin/ with new Options pattern
- Migrate analyzer logic
- Update main.go to use new structure
- Delete old internal/ code

**Step 4: Cleanup and Document**
- Remove unused configuration
- Add comprehensive tests
- Update documentation
- Prepare for krew distribution

### File Moves Summary

**Current Location → Target Location:**

```
internal/reporter/report.go → pkg/output/printer.go (interface)
                            → pkg/output/table.go (TablePrinter)
                            → pkg/output/json.go (JSONPrinter)

internal/analyzer/pod_analyzer.go → pkg/plugin/analyzer.go
(extractRegistryAndTag logic) → pkg/image/parser.go

internal/cluster/client.go → Keep internal (spinner, pagination)
                           → pkg/kubernetes/client.go (interface impl)
                           → pkg/kubernetes/interface.go (new)
                           → pkg/kubernetes/fake.go (new)

cmd/kubectl-analyze-images/main.go → Simplify (keep location)
                                   → pkg/plugin/options.go (new)
                                   → pkg/plugin/plugin.go (new)

pkg/types/image.go → Keep (types belong here)
pkg/types/analysis.go → Keep but clean up unused fields
pkg/types/visualization.go → Keep
pkg/types/pod.go → Keep
```

**New Files to Create:**
- pkg/kubernetes/interface.go (new abstraction)
- pkg/kubernetes/fake.go (test doubles)
- pkg/plugin/options.go (Complete/Validate/Run)
- pkg/plugin/plugin.go (orchestration)
- pkg/image/parser.go (shared utilities)
- pkg/output/printer.go (interface)
- test/fixtures/pods.yaml (test data)
- test/integration/analyze_test.go (integration tests)

**Files to Delete:**
- None initially (keep for safety)
- After validation: delete old internal/reporter/, consolidate logic

## Success Metrics

**Testability:**
- Can run full test suite in <5 seconds (no cluster)
- Test coverage >80%
- All business logic covered by unit tests
- Integration tests with FakeClient

**Maintainability:**
- main.go <100 lines
- Each package has single clear responsibility
- No code duplication
- Clear dependency graph (no cycles)

**Extensibility:**
- Adding new output format requires only new file in pkg/output/
- Adding new K8s query requires only new method on kubernetes.Interface
- Can use pkg/plugin/ as library in other tools

**Compliance:**
- Follows golang-standards/project-layout
- Follows kubectl CLI patterns (Complete/Validate/Run)
- Ready for krew distribution
- Cross-platform build support

## References and Resources

**Official Documentation:**
- Kubernetes CLI Conventions: https://kubernetes.io/docs/reference/kubectl/conventions/
- Krew Developer Guide: https://krew.sigs.k8s.io/docs/developer-guide/
- Go Project Layout: https://github.com/golang-standards/project-layout

**Example Projects:**
- kubectl-tree (ahmetb): https://github.com/ahmetb/kubectl-tree
- kubectl-images (chenjiandongx): https://github.com/chenjiandongx/kubectl-images
- kubectl-who-can (aquasecurity): https://github.com/aquasecurity/kubectl-who-can
- kubectl-slice (patrickdappollonio): https://github.com/patrickdappollonio/kubectl-slice

**Libraries:**
- k8s.io/client-go: Kubernetes client library
- k8s.io/cli-runtime: kubectl CLI utilities (genericclioptions)
- k8s.io/apimachinery: Kubernetes types and utilities
- github.com/spf13/cobra: CLI framework
- github.com/olekukonko/tablewriter: Table formatting

**Testing Resources:**
- k8s.io/client-go/kubernetes/fake: Fake Kubernetes clientset
- testing package: Go standard testing
- testify/assert: Test assertions

**Build and Distribution:**
- goreleaser: Cross-platform release automation
- GitHub Actions: CI/CD for building and testing
- krew plugin index: https://github.com/kubernetes-sigs/krew-index

---

**Next Steps:**
1. Review this architecture research with stakeholders
2. Validate approach with sample refactoring (Phase 1, Step 1)
3. Create detailed task breakdown for Phase 1
4. Begin implementation following roadmap

**Maintenance:**
- Update this document as patterns evolve
- Add ADRs for new architectural decisions
- Review after each phase completion
