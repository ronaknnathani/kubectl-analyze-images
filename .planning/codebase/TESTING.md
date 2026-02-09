# Testing Patterns

**Analysis Date:** 2026-02-09

## Test Framework

**Runner:**
- Go's built-in `testing` package (standard library)
- No external test framework detected (no `testing.T` wrappers)
- Config: `Makefile` contains test target

**Assertion Library:**
- Not detected; tests would use native Go assertions
- No testing helper libraries imported (e.g., `testify`, `assert`, `require`)

**Run Commands:**
```bash
make test              # Run all tests (executes: go test ./...)
go test ./...          # Test all packages recursively
go test ./path         # Test specific package
go test -v             # Verbose output
go test -cover         # With coverage metrics
```

## Test File Organization

**Location:**
- Pattern: Co-located with source files (Go convention)
- Convention: `*_test.go` files in same directory as implementation
- Currently: **No test files exist in codebase** (none detected in find results)

**Naming:**
- Expected pattern: `client_test.go` for `client.go`, `pod_analyzer_test.go` for `pod_analyzer.go`
- Test functions would use pattern: `TestFunctionName(t *testing.T)`

**Structure:**
- Tests would be located in same package as code (not `_test` package suffix)
- Examples based on Go convention:
  - `internal/cluster/client_test.go` - for cluster client tests
  - `internal/analyzer/pod_analyzer_test.go` - for analyzer tests
  - `internal/reporter/report_test.go` - for reporter tests
  - `pkg/types/image_test.go` - for image type tests

## Test Structure

**Suite Organization:**
Currently no test files exist. Expected pattern based on Go conventions:
```go
package cluster

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Test setup
	// Test execution
	// Test assertions
}

func TestListPods(t *testing.T) {
	// Setup
	// Execute
	// Assert
}
```

**Patterns:**
- Setup: Initialize test data, create mock clients
- Execution: Call function under test
- Assertion: Verify results and error conditions
- Teardown: Implicit (Go testing cleanup)

**Teardown:**
- Go testing doesn't require explicit teardown for most cases
- File cleanup handled by defer statements in tests
- Context cancellation for goroutines

## Mocking

**Framework:**
- Not yet implemented; would use standard Go patterns
- Candidates: `github.com/stretchr/testify/mock`, `github.com/golang/mock`, or interface-based mocking

**Patterns (Expected):**
- Interface-based mocking by creating mock implementations
- Example structure:
```go
type MockKubernetesClient struct {
	callCount int
}

func (m *MockKubernetesClient) ListPods(ctx context.Context, namespace string) ([]Pod, error) {
	m.callCount++
	return []Pod{}, nil
}
```

**What to Mock:**
- Kubernetes API client interactions (`k8s.io/client-go`)
- Spinner display calls (optional, for clean test output)
- File system operations if added
- Network calls to registries

**What NOT to Mock:**
- Core business logic (image analysis, filtering)
- Type conversions and data transformations
- Error handling code paths
- Local data structures and slices

## Fixtures and Factories

**Test Data:**
Currently no test fixtures exist. Expected pattern for image analysis:
```go
// Factory for test pods
func createTestPod(name, namespace string, images []string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Image: images[0]},
			},
		},
	}
}

// Test data fixture
var testPods = []types.Pod{
	{
		Name:      "nginx-1",
		Namespace: "default",
		Images:    []string{"nginx:1.21", "redis:7.0"},
	},
	{
		Name:      "app-2",
		Namespace: "production",
		Images:    []string{"myapp:1.0"},
	},
}
```

**Location:**
- Helper functions: Package-level in `*_test.go` files
- Shared fixtures: Separate `testdata/` or `testfixtures/` directory
- Kubernetes mock objects: Could be in `internal/testutil/` package

## Coverage

**Requirements:**
- No coverage target enforced (no CI config detected)
- Optional best practice: Aim for 70%+ coverage for critical paths
- Critical paths to test:
  - Image analysis logic: `internal/analyzer/pod_analyzer.go`
  - Error handling in cluster client: `internal/cluster/client.go`
  - Output formatting: `internal/reporter/report.go`

**View Coverage:**
```bash
go test -cover ./...           # Summary
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out  # HTML report
```

## Test Types

**Unit Tests:**
- **Scope:** Individual functions and methods
- **Approach:** Mock dependencies, test with various inputs
- **Location:** Same package as implementation (`*_test.go`)
- **Key candidates:**
  - `extractRegistryAndTag()` - logic testing for image parsing
  - `selectBestImageName()` - selection logic with edge cases
  - `containsSHA()`, `isHexString()` - string validation
  - Type methods: `GetTopImagesBySize()`, `GenerateImageSizeHistogram()`

**Integration Tests:**
- **Scope:** Multiple components working together
- **Approach:** Test with real Kubernetes client or mock K8s API
- **Not yet implemented:** Would require `k8s.io/client-go/fake` package
- **Example:** Test full analysis flow from pod listing through report generation
- **Candidates:**
  - Full analyzer flow: list pods → get image sizes → analyze
  - Reporter formatting: JSON and table output generation

**E2E Tests:**
- **Framework:** Not currently used
- **Would require:** Live Kubernetes cluster or kind/minikube setup
- **Alternative:** Integration tests with mock API sufficient for plugin

## Common Patterns

**Async Testing:**
Go testing with contexts (expected pattern):
```go
func TestAnalyzePods(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	analyzer, _ := analyzer.NewPodAnalyzer()
	analysis, err := analyzer.AnalyzePods(ctx, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
```

**Error Testing:**
Pattern for verifying error handling:
```go
func TestNewClientWithInvalidContext(t *testing.T) {
	client, err := cluster.NewClient("nonexistent-context")
	if err == nil {
		t.Fatal("expected error for invalid context")
	}
	if client != nil {
		t.Fatal("client should be nil on error")
	}
}
```

**Table-Driven Tests:**
Recommended pattern for multiple scenarios:
```go
func TestExtractRegistryAndTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantReg  string
		wantTag  string
	}{
		{"simple", "nginx:1.21", "nginx", "1.21"},
		{"with registry", "docker.io/library/nginx:latest", "docker.io", "latest"},
		{"default tag", "nginx", "nginx", "latest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, tag := extractRegistryAndTag(tt.input)
			if reg != tt.wantReg || tag != tt.wantTag {
				t.Errorf("got %q, %q; want %q, %q", reg, tag, tt.wantReg, tt.wantTag)
			}
		})
	}
}
```

## Testing Gaps

**Currently Missing:**
- **No test files exist** - all packages need unit tests
- **Critical untested areas:**
  - `internal/cluster/client.go` - Kubernetes API interactions (267 lines)
  - `internal/analyzer/pod_analyzer.go` - Core analysis logic (168 lines)
  - `internal/reporter/report.go` - Output formatting (173 lines)
  - `pkg/types/visualization.go` - Histogram generation (281 lines)

**Recommendation:**
- Start with unit tests for utility functions: `extractRegistryAndTag()`, `selectBestImageName()`, `isHexString()`
- Add integration tests for analyzer and reporter
- Mock Kubernetes client for cluster tests

---

*Testing analysis: 2026-02-09*
