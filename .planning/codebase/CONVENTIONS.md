# Coding Conventions

**Analysis Date:** 2026-02-09

## Naming Patterns

**Files:**
- Go files use snake_case with functional names: `pod_analyzer.go`, `client.go`, `report.go`
- Main executable entry point: `main.go` in `cmd/` directory
- Test files follow Go standard: `*_test.go` (none currently exist)

**Functions:**
- Public functions use PascalCase: `NewClient()`, `ListPods()`, `GenerateReport()`, `SetNoColor()`
- Private functions use camelCase: `namespaceDisplay()`, `selectBestImageName()`, `extractRegistryAndTag()`, `isHexString()`, `containsSHA()`
- Constructor functions follow pattern: `New[Type]()` or `New[Type]WithConfig()`
- Receiver methods are concise: `(c *Client)`, `(r *Reporter)`, `(pa *PodAnalyzer)`, `(ia *ImageAnalysis)`

**Variables:**
- Short names for local variables: `ctx`, `err`, `pod`, `pods`, `img`, `images`, `s` (spinner), `config`
- Descriptive names for struct fields: `clientset`, `labelSelector`, `outputFormat`, `noColor`, `totalSize`
- Constants use PascalCase: `ResourceVersion`, `PageSize`

**Types:**
- Struct names use PascalCase: `Client`, `PodAnalyzer`, `Reporter`, `Image`, `ImageAnalysis`, `PerformanceMetrics`, `HistogramConfig`
- Interface names end with lowercase suffix pattern (not yet used in this codebase)
- Type aliases follow PascalCase

## Code Style

**Formatting:**
- Standard Go formatting (implied gofmt style)
- Two-space indentation in logs and structured output
- Line length not explicitly enforced; files range 42-281 lines

**Linting:**
- No explicit linting configuration found (no `.golangci.yml`, `.eslintrc`, etc.)
- No code formatter configuration detected
- Standard Go conventions applied implicitly

## Import Organization

**Order:**
1. Standard library imports (e.g., `context`, `fmt`, `os`, `time`, `encoding/json`)
2. Third-party imports (e.g., `github.com/spf13/cobra`, `github.com/fatih/color`, `k8s.io/*`)
3. Internal module imports (e.g., `github.com/ronaknnathani/kubectl-analyze-images/...`)

**Pattern:** Imports grouped by category with blank lines separating groups. Example from `cmd/kubectl-analyze-images/main.go`:
```go
import (
	"context"
	"fmt"
	"os"

	"github.com/ronaknnathani/kubectl-analyze-images/internal/analyzer"
	"github.com/ronaknnathani/kubectl-analyze-images/internal/reporter"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
	"github.com/spf13/cobra"
)
```

**Path Aliases:**
- No import aliases used; full paths maintained for clarity
- Kubernetes imports use abbreviated form: `corev1 "k8s.io/api/core/v1"`, `metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"`

## Error Handling

**Patterns:**
- Error wrapping with context: `fmt.Errorf("failed to create analyzer: %w", err)`
- Early returns on error: `if err != nil { return ... }`
- Propagation of wrapped errors up the call stack
- Errors prefixed with context ("failed to X:", "failed to list Y:")

**Examples:**
- `cmd/kubectl-analyze-images/main.go` line 61: `return fmt.Errorf("failed to create analyzer: %w", err)`
- `internal/cluster/client.go` line 38: `return nil, fmt.Errorf("failed to load kubeconfig: %w", err)`
- `internal/reporter/report.go` line 154: `return fmt.Errorf("failed to marshal JSON: %w", err)`

## Logging

**Framework:** Standard library `fmt` package and `os.Stderr` for status output

**Patterns:**
- User-facing messages to `os.Stderr`: `fmt.Fprintf(os.Stderr, "message\n", args)`
- Results and reports to `os.Stdout`: `fmt.Println()`, `fmt.Print()`
- Spinner library (`github.com/briandowns/spinner`) for progress indication
- Success indicator with check mark: `✓ Found %d pods...`
- Suffixes for spinners: `" Querying pods from cluster..."`

**Examples:**
- Line 69 in `cmd/kubectl-analyze-images/main.go`: `fmt.Printf("Analyzing images in namespace: %s\n", namespaceDisplay)`
- Line 112 in `internal/cluster/client.go`: `fmt.Fprintf(os.Stderr, "✓ Found %d pods across all namespaces...\n", totalPods)`

## Comments

**When to Comment:**
- Function comments precede public function declarations: `// NewClient creates a new Kubernetes client`
- Multi-line functions have inline comments for logic steps
- Helper functions documented with purpose: `// selectBestImageName selects the best canonical name...`
- Edge case handling documented: `// Handle edge cases`
- Design decisions documented: `// SHA digests are typically 64 characters...`

**Style:**
- Comments start with function name or description: `// Client represents...`, `// ListPods lists pods...`
- Sentence case with proper grammar
- Multi-line comments explain WHY, not WHAT

**Examples:**
- `internal/cluster/client.go` line 21: `// Client represents a Kubernetes cluster client`
- `internal/cluster/client.go` line 214-215: `// selectBestImageName selects the best canonical name from a list of image names\n// Prefers names without SHA digests, then falls back to the first one`

## Function Design

**Size:**
- Functions range from 3-100 lines
- Most are 20-50 lines for readability
- Largest file: `internal/cluster/client.go` (267 lines) with multiple functions

**Parameters:**
- 3-6 parameters typical for main functions
- Receiver methods pass data through struct fields
- Config objects used for optional parameters: `*types.AnalysisConfig`

**Return Values:**
- Main operations return result, metrics, and error: `(result, metrics, error)`
- Constructors return instance and error: `(*Type, error)`
- Simple getters return single value or map
- Always return errors as last return value

**Receiver Methods:**
- Pointer receivers for structs modifying state: `(r *Reporter)`, `(pa *PodAnalyzer)`
- Methods returning new data may use value receivers

## Module Design

**Exports:**
- Public types exported (PascalCase): `Client`, `Reporter`, `PodAnalyzer`, `Image`, `ImageAnalysis`
- Public methods on exported types use full documentation
- Private helper functions use camelCase: `selectBestImageName()`, `extractRegistryAndTag()`

**Package Organization:**
- `cmd/` - Entry point and CLI setup
- `internal/` - Private business logic
  - `cluster/` - Kubernetes API interactions
  - `analyzer/` - Image analysis orchestration
  - `reporter/` - Output generation
- `pkg/` - Shared types and utilities
  - `types/` - Domain models (`Image`, `Pod`, `AnalysisConfig`, etc.)

**Barrel Files:**
- No barrel exports (`__init__.go` patterns); each package is independent
- Direct type imports: `github.com/ronaknnathani/kubectl-analyze-images/pkg/types`

---

*Convention analysis: 2026-02-09*
