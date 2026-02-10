---
phase: 02-kubernetes-abstraction-layer
plan: 02
subsystem: dependency-injection
tags: [refactor, testing, dependency-injection, fake-client, unit-tests]
dependency_graph:
  requires:
    - pkg/kubernetes/Interface (from 02-01)
    - pkg/kubernetes/NewClient (from 02-01)
    - pkg/kubernetes/NewFakeClient (from 02-01)
  provides:
    - internal/cluster/NewClient accepting kubernetes.Interface
    - internal/analyzer/NewPodAnalyzer accepting *cluster.Client
    - Complete unit test suite for cluster and analyzer packages
  affects:
    - cmd/kubectl-analyze-images/main.go (dependency injection wiring)
tech_stack:
  added:
    - Table-driven unit tests using testify
  patterns:
    - Dependency injection constructor pattern
    - Test helper functions for creating test fixtures
    - FakeClient-based unit testing without real cluster
key_files:
  created:
    - internal/cluster/client_test.go
    - internal/analyzer/pod_analyzer_test.go
    - cmd/kubectl-analyze-images/main.go
  modified:
    - internal/cluster/client.go
    - internal/analyzer/pod_analyzer.go
    - pkg/util/image.go (bug fix)
    - pkg/util/image_test.go (test expectations updated)
    - pkg/types/image_test.go (test expectations updated)
decisions:
  - Removed all kubeconfig loading logic from cluster.Client (now in pkg/kubernetes)
  - cluster.Client constructor no longer returns error (no I/O operations)
  - analyzer.PodAnalyzer constructor simplified to accept dependencies
  - Main.go creates explicit dependency chain: kubernetes.NewClient -> cluster.NewClient -> analyzer.NewPodAnalyzer
  - Fixed default registry from "unknown" to "docker.io" for single-component images
  - Spinner/UX code preserved in cluster client (tested incidentally, not targeted for coverage)
metrics:
  duration: 336s
  completed: 2026-02-10T03:27:29Z
---

# Phase 02 Plan 02: Dependency Injection Refactor and Unit Tests Summary

**One-liner:** Refactored cluster and analyzer packages for dependency injection with kubernetes.Interface, added comprehensive unit tests achieving 86.6% and 93.6% coverage

## What Was Built

Completed the interface-driven refactor by updating cluster and analyzer packages to use the kubernetes abstraction layer created in plan 02-01. Added comprehensive unit tests using FakeClient for full testability without requiring a real Kubernetes cluster.

### Task 1: Dependency Injection Refactor

**internal/cluster/client.go:**
- Changed `Client` struct from holding `*kubernetes.Clientset` to `kubernetes.Interface`
- Simplified constructor: `NewClient(k8sClient kubernetes.Interface) *Client` (no error return)
- Removed all kubeconfig loading logic (moved to pkg/kubernetes in plan 02-01)
- Updated `ListPods` and `GetImageSizesFromNodes` to call interface methods
- Preserved all spinner/progress UX logic unchanged
- Used pager pattern with interface: `pager.New(func(ctx, opts) { return c.k8sClient.ListPods(ctx, namespace, opts) })`

**internal/analyzer/pod_analyzer.go:**
- Removed `NewPodAnalyzer()` and `NewPodAnalyzerWithConfig(config, contextName)` constructors
- Added new constructor: `NewPodAnalyzer(clusterClient *cluster.Client, config *types.AnalysisConfig) *PodAnalyzer`
- Constructor no longer creates cluster client internally - accepts it as dependency
- No changes to `AnalyzePods` method (already used injected cluster client)

**cmd/kubectl-analyze-images/main.go:**
- Added imports for `pkg/kubernetes` and `internal/cluster`
- Implemented explicit dependency injection chain in `runAnalyze`:
  1. `k8sClient, err := kubernetes.NewClient(kubeContext)`
  2. `clusterClient := cluster.NewClient(k8sClient)`
  3. `podAnalyzer := analyzer.NewPodAnalyzer(clusterClient, config)`
- Renamed variable from `analyzer` to `podAnalyzer` to avoid shadowing package import

**Commits:**
- db79a51: Refactored cluster and analyzer for dependency injection
- 72fc4e8: Added main.go with dependency injection chain

### Task 2: Unit Tests and Bug Fix

**internal/cluster/client_test.go:**
- Test helper: `createTestPod(name, namespace, images...)` for pod fixtures
- Test helper: `createTestNode(name, imageSizes)` for node fixtures
- `TestClient_ListPods`: 4 test cases covering single pod, multiple namespaces, no pods, multiple containers
- `TestClient_GetImageSizesFromNodes`: 3 test cases covering single node, overlapping images, no images
- `TestClient_GetUniqueImages`: 2 test cases covering deduplication and empty list
- `TestSelectBestImageName`: 4 test cases for SHA vs non-SHA name selection
- All tests use `kubernetes.NewFakeClient(objects...)` for test doubles
- Coverage: 86.6% (spinner code covered incidentally during test execution)

**internal/analyzer/pod_analyzer_test.go:**
- `TestPodAnalyzer_AnalyzePods_WithPods`: Tests end-to-end analysis with pods and node images
- `TestPodAnalyzer_AnalyzePods_NoNamespace`: Tests analysis using all node images (no pod filter)
- `TestPodAnalyzer_AnalyzePods_MissingImageSize`: Tests inaccessible image handling (pod image not on nodes)
- `TestPodAnalyzer_AnalyzePods_MultipleNamespaces`: Tests namespace filtering behavior
- Coverage: 93.6%

**Bug Fix (Rule 1 - Auto-fix bugs):**
- **Issue:** `pkg/util/image.go` `ExtractRegistryAndTag` failed to extract tags from single-component images like "nginx:1.21"
- **Root cause:** Function only checked for tag when `len(parts) >= 2` (after splitting on `/`), so single-component images always got default "latest" tag
- **Fix:** Moved tag extraction logic outside the registry check, so tags are extracted from the last part regardless of slash count
- **Additional fix:** Changed default registry from "unknown" to "docker.io" (matches Kubernetes behavior for unqualified images)
- **Test updates:** Updated expectations in `pkg/util/image_test.go` and `pkg/types/image_test.go` to match correct behavior
- **Files modified:** pkg/util/image.go, pkg/util/image_test.go, pkg/types/image_test.go

**Commit:**
- 65f32f3: Added unit tests and fixed image tag extraction bug

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed incorrect tag extraction for single-component images**
- **Found during:** Task 2 (test execution revealed tag extraction failure)
- **Issue:** `ExtractRegistryAndTag("nginx:1.21")` returned registry="unknown", tag="latest" instead of tag="1.21"
- **Root cause:** Tag extraction logic was inside `if len(parts) >= 2` block, so single-component images (no slash) never had tags extracted
- **Fix:** Moved tag extraction outside registry check, extract from last path component regardless of slash count
- **Impact:** Correct tag extraction for all image formats (single-component, multi-component, with/without registry)
- **Files modified:** pkg/util/image.go, pkg/util/image_test.go, pkg/types/image_test.go
- **Commit:** 65f32f3 (included with Task 2)

**2. [Rule 1 - Bug Fix] Changed default registry from "unknown" to "docker.io"**
- **Found during:** Task 2 (fixing tag extraction bug)
- **Issue:** Single-component images like "nginx" were assigned registry="unknown", which is semantically incorrect
- **Reasoning:** Kubernetes implicitly treats unqualified images as docker.io/library/[image] - using "docker.io" as default better reflects actual behavior
- **Impact:** More accurate registry reporting for single-component images
- **Files modified:** pkg/util/image.go, test expectations in pkg/util/image_test.go and pkg/types/image_test.go
- **Commit:** 65f32f3 (included with Task 2)

## Verification Results

All verification checks passed:
- ✅ `go build ./...` - compiles entire project without errors
- ✅ `go vet ./...` - passes
- ✅ `go test ./...` - all tests pass (Phase 1 + Phase 2)
- ✅ `go test -cover ./internal/cluster/` - 86.6% coverage (exceeds 60% target)
- ✅ `go test -cover ./internal/analyzer/` - 93.6% coverage (exceeds 60% target)
- ✅ `make build` - succeeds
- ✅ No direct `*kubernetes.Clientset` usage in internal/cluster/client.go
- ✅ No client creation in internal/analyzer/pod_analyzer.go
- ✅ main.go has 3-step dependency injection chain

### Coverage Details

**internal/cluster package: 86.6%**
- ListPods: fully covered (including spinner output to stderr)
- GetImageSizesFromNodes: fully covered (including spinner output to stderr)
- GetUniqueImages: fully covered
- selectBestImageName: fully covered
- Helper functions (namespaceDisplay, containsSHA, isHexString): fully covered
- Note: Spinner lines show as "covered" because tests execute through them, but coverage target focused on API calls and data transformation logic

**internal/analyzer package: 93.6%**
- AnalyzePods: fully covered across multiple scenarios
- Constructor: fully covered
- Edge cases: missing image sizes, namespace filtering, all-namespaces mode

## Task Breakdown

| Task | Name | Commits | Files | Status |
|------|------|---------|-------|--------|
| 1 | Refactor cluster client, analyzer, and main.go | db79a51, 72fc4e8 | internal/cluster/client.go, internal/analyzer/pod_analyzer.go, cmd/kubectl-analyze-images/main.go | ✅ Complete |
| 2 | Add unit tests for cluster and analyzer | 65f32f3 | internal/cluster/client_test.go, internal/analyzer/pod_analyzer_test.go, pkg/util/image.go, pkg/util/image_test.go, pkg/types/image_test.go | ✅ Complete |

## Technical Details

### Dependency Injection Pattern

Before (plan 02-01):
```go
// Cluster client creates its own Kubernetes clientset
clusterClient, err := cluster.NewClient(contextName)

// Analyzer creates its own cluster client
analyzer, err := analyzer.NewPodAnalyzerWithConfig(config, contextName)
```

After (plan 02-02):
```go
// Main creates Kubernetes client
k8sClient, err := kubernetes.NewClient(kubeContext)

// Injects into cluster client
clusterClient := cluster.NewClient(k8sClient)

// Injects into analyzer
podAnalyzer := analyzer.NewPodAnalyzer(clusterClient, config)
```

### Pager Integration with Interface

The pager pattern works seamlessly with the interface because `*corev1.PodList` and `*corev1.NodeList` both implement `runtime.Object`:

```go
pager := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
    return c.k8sClient.ListPods(ctx, namespace, opts)
})
```

The interface method returns typed lists, pager expects `runtime.Object`, and Go's type system handles the implicit conversion.

### Test Architecture

Tests use `kubernetes.NewFakeClient(objects...)` to inject pre-populated fake clientsets:

```go
// Create test fixtures
pod1 := createTestPod("pod1", "default", "nginx:1.21")
node1 := createTestNode("node1", map[string]int64{"nginx:1.21": 100000000})

// Inject into fake client
fakeK8s := kubernetes.NewFakeClient(pod1, node1)

// Create cluster client with fake
clusterClient := cluster.NewClient(fakeK8s)

// Test cluster operations without real cluster
pods, metrics, err := clusterClient.ListPods(ctx, "default", "")
```

This pattern enables:
- Zero external dependencies for tests
- Fast execution (milliseconds vs seconds)
- Deterministic test data
- Full control over edge cases (missing images, empty namespaces, etc.)

## Impact

### Immediate
- All Kubernetes API interactions now go through interface abstraction
- Cluster and analyzer packages fully testable without real cluster
- Main.go explicitly shows dependency flow (easier to understand and modify)
- Bug fix improves tag extraction accuracy for all image formats

### Architecture
- Completes Phase 2 dependency injection refactor
- Establishes testing pattern for future packages
- Enables future work: registry client, policy engine can follow same pattern
- Clean separation: kubernetes client creation (main) vs. kubernetes operations (packages)

### Testing
- High coverage in critical packages validates core functionality
- Spinner/UX code tested incidentally (doesn't need explicit coverage)
- Tests serve as usage examples for cluster and analyzer packages
- Foundation for integration tests in future phases

## Next Steps (Phase 3 and beyond)

Phase 2 is now complete. The kubernetes abstraction layer is fully integrated, tested, and ready for use. Future phases can:
- Use `pkg/kubernetes.Interface` for any new Kubernetes operations
- Follow the same dependency injection pattern established here
- Write unit tests using `kubernetes.NewFakeClient` for test doubles
- Build on top of cluster and analyzer packages without modifying them

## Self-Check: PASSED

Verified all claimed artifacts exist:

**Created files:**
- ✅ internal/cluster/client_test.go
- ✅ internal/analyzer/pod_analyzer_test.go
- ✅ cmd/kubectl-analyze-images/main.go

**Modified files:**
- ✅ internal/cluster/client.go (refactored to use interface)
- ✅ internal/analyzer/pod_analyzer.go (refactored for dependency injection)
- ✅ pkg/util/image.go (bug fix for tag extraction)
- ✅ pkg/util/image_test.go (updated expectations)
- ✅ pkg/types/image_test.go (updated expectations)

**Commits:**
- ✅ db79a51 (Task 1 - refactor cluster and analyzer)
- ✅ 72fc4e8 (Task 1 - add main.go)
- ✅ 65f32f3 (Task 2 - tests and bug fix)

**Verification:**
- ✅ All tests pass: `go test ./...`
- ✅ Coverage exceeds 60%: cluster 86.6%, analyzer 93.6%
- ✅ Build succeeds: `make build`
- ✅ No regressions in Phase 1 tests
