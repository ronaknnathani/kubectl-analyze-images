# Phase 2: Kubernetes Abstraction Layer - Research

**Researched:** 2026-02-09
**Domain:** Go interface design, client-go abstractions, testable Kubernetes clients
**Confidence:** HIGH

## Summary

Phase 2 creates a Kubernetes interface abstraction layer to enable testable cluster interactions. The current implementation directly uses `*kubernetes.Clientset` throughout the codebase, making it impossible to test without a real cluster. By introducing a `kubernetes.Interface` with methods for all cluster operations (ListPods, ListNodes), we can create both a real implementation that wraps client-go and a FakeClient for testing.

This is a standard pattern in Kubernetes ecosystem projects (controller-runtime, kubectl plugins, operators). The key insight: client-go already provides `fake.NewSimpleClientset()` for creating test clients with pre-populated objects, so our FakeClient can wrap this instead of manually implementing mock behavior.

**Primary recommendation:** Use interface-driven design with client-go's built-in fake client for testing. Create `pkg/kubernetes/interface.go` defining operations, `pkg/kubernetes/client.go` wrapping real clientset, and `pkg/kubernetes/fake.go` wrapping fake clientset.

## Standard Stack

### Core Libraries (Already Present)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| k8s.io/client-go | v0.29.0 | Kubernetes API client | Official Kubernetes client library |
| k8s.io/client-go/kubernetes/fake | v0.29.0 | Fake Kubernetes client | Official testing package from client-go |
| k8s.io/api | v0.29.0 | Kubernetes API types | Required for corev1.Pod, corev1.Node types |
| k8s.io/apimachinery | v0.29.0 | API machinery | Required for metav1.ListOptions, runtime.Object |
| github.com/stretchr/testify | v1.11.1 | Testing assertions | Industry standard for Go testing (added in Phase 1) |

**No new dependencies needed.** All required packages are already in go.mod.

### Supporting Patterns
| Pattern | Purpose | When to Use |
|---------|---------|-------------|
| Interface segregation | Small focused interfaces | When component needs subset of operations |
| Constructor functions | Create interface implementations | `NewClient()`, `NewFakeClient()` |
| Table-driven tests | Parameterized test cases | Testing multiple scenarios efficiently |

## Architecture Patterns

### Recommended Project Structure
```
pkg/
├── kubernetes/
│   ├── interface.go       # Interface definition with ListPods(), ListNodes()
│   ├── client.go          # Real implementation wrapping kubernetes.Clientset
│   ├── fake.go            # Fake implementation wrapping fake.Clientset
│   └── client_test.go     # Tests using FakeClient
internal/
├── cluster/
│   ├── client.go          # Refactored to accept kubernetes.Interface
│   └── client_test.go     # NEW: Unit tests using kubernetes.FakeClient
└── analyzer/
    ├── pod_analyzer.go    # Refactored to accept kubernetes.Interface
    └── pod_analyzer_test.go # NEW: Unit tests using kubernetes.FakeClient
```

### Pattern 1: Interface Definition with Context
**What:** Define minimal interface for Kubernetes operations with context.Context support
**When to use:** When you need to abstract external dependencies for testability
**Example:**
```go
// Source: Standard Go interface pattern, client-go v0.29.0 APIs
package kubernetes

import (
    "context"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Interface defines the contract for Kubernetes cluster operations
type Interface interface {
    // ListPods lists pods in the specified namespace with optional label selector
    ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error)

    // ListNodes lists all nodes in the cluster with optional options
    ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error)
}
```

**Key design decisions:**
- Use `context.Context` as first parameter (Go standard for cancellable operations)
- Return native Kubernetes types (`*corev1.PodList`, `*corev1.NodeList`) instead of custom types
- Keep interface minimal (only methods actually needed by the application)
- Use `metav1.ListOptions` to preserve all filtering capabilities (label selectors, field selectors, etc.)

### Pattern 2: Real Client Implementation
**What:** Wrapper around kubernetes.Clientset that implements the Interface
**When to use:** For production code that needs to interact with real clusters
**Example:**
```go
// Source: Standard client-go usage pattern
package kubernetes

import (
    "context"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

// Client implements Interface using a real kubernetes.Clientset
type Client struct {
    clientset *kubernetes.Clientset
    config    *rest.Config
}

// NewClient creates a new Kubernetes client from kubeconfig
func NewClient(contextName string) (Interface, error) {
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    configOverrides := &clientcmd.ConfigOverrides{}
    if contextName != "" {
        configOverrides.CurrentContext = contextName
    }

    clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
        loadingRules, configOverrides)
    config, err := clientConfig.ClientConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create clientset: %w", err)
    }

    return &Client{
        clientset: clientset,
        config:    config,
    }, nil
}

// ListPods implements Interface
func (c *Client) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
    return c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
}

// ListNodes implements Interface
func (c *Client) ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error) {
    return c.clientset.CoreV1().Nodes().List(ctx, opts)
}
```

**Key implementation notes:**
- Constructor returns `Interface`, not `*Client` (callers depend on interface, not concrete type)
- Direct delegation to clientset methods (minimal wrapper overhead)
- Preserve all client-go error semantics (return errors as-is for proper handling)
- Keep config for potential future needs (authentication, custom transports)

### Pattern 3: Fake Client for Testing
**What:** Test implementation using client-go's fake.Clientset
**When to use:** In unit tests to simulate Kubernetes API responses without a real cluster
**Example:**
```go
// Source: client-go/kubernetes/fake package
package kubernetes

import (
    "context"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/client-go/kubernetes/fake"
)

// FakeClient implements Interface using a fake kubernetes.Clientset for testing
type FakeClient struct {
    clientset *fake.Clientset
}

// NewFakeClient creates a new fake Kubernetes client pre-populated with objects
func NewFakeClient(objects ...runtime.Object) Interface {
    return &FakeClient{
        clientset: fake.NewSimpleClientset(objects...),
    }
}

// ListPods implements Interface
func (f *FakeClient) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
    return f.clientset.CoreV1().Pods(namespace).List(ctx, opts)
}

// ListNodes implements Interface
func (f *FakeClient) ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error) {
    return f.clientset.CoreV1().Nodes().List(ctx, opts)
}
```

**Usage in tests:**
```go
// Source: Standard table-driven test pattern with testify
func TestClusterClient_ListPods(t *testing.T) {
    // Create test pods
    testPod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-pod",
            Namespace: "default",
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {Name: "nginx", Image: "nginx:1.21"},
            },
        },
    }

    // Create fake client with test data
    fakeClient := kubernetes.NewFakeClient(testPod)

    // Test ListPods
    pods, err := fakeClient.ListPods(context.Background(), "default", metav1.ListOptions{})
    require.NoError(t, err)
    assert.Len(t, pods.Items, 1)
    assert.Equal(t, "test-pod", pods.Items[0].Name)
}
```

### Pattern 4: Dependency Injection in Consumers
**What:** Accept `kubernetes.Interface` in constructors instead of creating clients internally
**When to use:** In any component that needs Kubernetes operations (analyzers, controllers, etc.)
**Example:**
```go
// Source: Dependency injection best practice
package cluster

import (
    "github.com/ronaknnathani/kubectl-analyze-images/pkg/kubernetes"
    "github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
)

// Client represents a cluster operations client
type Client struct {
    k8sClient kubernetes.Interface  // Accept interface, not concrete type
}

// NewClient creates a new cluster client with injected Kubernetes client
func NewClient(k8sClient kubernetes.Interface) *Client {
    return &Client{
        k8sClient: k8sClient,
    }
}

// ListPods lists pods and converts to internal types
func (c *Client) ListPods(ctx context.Context, namespace, labelSelector string) ([]types.Pod, error) {
    opts := metav1.ListOptions{
        ResourceVersion: "0",  // Use watch cache
    }
    if labelSelector != "" {
        opts.LabelSelector = labelSelector
    }

    podList, err := c.k8sClient.ListPods(ctx, namespace, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to list pods: %w", err)
    }

    // Convert to internal types
    pods := make([]types.Pod, len(podList.Items))
    for i, pod := range podList.Items {
        pods[i] = types.FromK8sPod(&pod)
    }

    return pods, nil
}
```

**Testing with dependency injection:**
```go
func TestClient_ListPods(t *testing.T) {
    // Arrange: Create test data
    testPod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-pod",
            Namespace: "default",
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {Name: "app", Image: "app:v1"},
            },
        },
    }

    fakeK8sClient := kubernetes.NewFakeClient(testPod)
    client := cluster.NewClient(fakeK8sClient)

    // Act: Call method under test
    pods, err := client.ListPods(context.Background(), "default", "")

    // Assert: Verify behavior
    require.NoError(t, err)
    assert.Len(t, pods, 1)
    assert.Equal(t, "test-pod", pods[0].Name)
    assert.Equal(t, []string{"app:v1"}, pods[0].Images)
}
```

### Pattern 5: Pager Integration with Interface
**What:** Continue using client-go's pager for efficient pagination through interface
**When to use:** When listing potentially large numbers of resources (pods, nodes)
**Example:**
```go
// Source: Current implementation in internal/cluster/client.go + interface pattern
package cluster

import (
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/client-go/tools/pager"
)

// GetImageSizesFromNodes uses pager with injected kubernetes.Interface
func (c *Client) GetImageSizesFromNodes(ctx context.Context) (map[string]int64, error) {
    imageSizes := make(map[string]int64)

    // Create pager that calls our interface method
    p := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
        return c.k8sClient.ListNodes(ctx, opts)
    })

    p.PageSize = 1000

    err := p.EachListItem(ctx, metav1.ListOptions{
        ResourceVersion: "0",  // Watch cache optimization
    }, func(obj runtime.Object) error {
        node := obj.(*corev1.Node)
        // Process node images
        for _, image := range node.Status.Images {
            if len(image.Names) > 0 {
                imageSizes[selectBestImageName(image.Names)] = image.SizeBytes
            }
        }
        return nil
    })

    return imageSizes, err
}
```

**Key insight:** Pager works seamlessly with our interface because it only needs a function that returns `runtime.Object`, which both real and fake clients provide.

### Anti-Patterns to Avoid

- **Exposing concrete clientset:** Never return `*kubernetes.Clientset` from interface methods. Always use Kubernetes API types (`*corev1.PodList`) or custom types.
  - **Why bad:** Breaks abstraction, prevents testing with fake client
  - **Do instead:** Return API types that both real and fake clients can produce

- **Creating clients inside business logic:** Don't call `kubernetes.NewClient()` inside analyzers or other business logic
  - **Why bad:** Makes code untestable, couples implementation to real cluster
  - **Do instead:** Accept `kubernetes.Interface` as constructor parameter (dependency injection)

- **Interface with too many methods:** Don't add every possible Kubernetes operation to the interface
  - **Why bad:** Increases maintenance burden, makes fake implementation tedious
  - **Do instead:** Start minimal (ListPods, ListNodes), add methods as needed

- **Returning custom types from interface:** Don't return `types.Pod` from interface methods
  - **Why bad:** Forces conversion logic into interface implementation, duplicates conversion across real/fake clients
  - **Do instead:** Return `*corev1.PodList` from interface, do conversion in consuming code

- **Mocking clientset directly with testify/mock:** Don't try to mock `*kubernetes.Clientset` methods
  - **Why bad:** Extremely verbose (clientset has 100+ methods), brittle, defeats purpose of fake client
  - **Do instead:** Use `kubernetes.NewFakeClient()` which wraps `fake.NewSimpleClientset()`

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Mock Kubernetes API | Custom mock structs with recorded calls | `fake.NewSimpleClientset()` from client-go | Fake client handles all CRUD operations, watches, selectors, pagination automatically |
| API pagination | Manual continue token handling | `pager.New()` from client-go | Handles chunking, errors, memory efficiently; works with real and fake clients |
| Object creation in tests | Manual struct initialization with all required fields | Helper functions + `fake.NewSimpleClientset()` | Reduces boilerplate, ensures valid objects, easier to maintain |
| Error simulation | Custom error-returning mocks | Fake client's built-in reactor mechanism | Can simulate specific API errors (not found, forbidden, etc.) declaratively |

**Key insight:** client-go's fake package is production-grade testing infrastructure used by Kubernetes itself. It handles edge cases (label selectors, field selectors, resource versions, pagination) that custom mocks would miss. Don't reinvent this wheel.

## Common Pitfalls

### Pitfall 1: Interface Method Signatures Don't Match Real Usage
**What goes wrong:** Interface methods have parameters/returns that don't align with how code actually uses the client (e.g., accepting `namespace` as string when you need structured ListOptions)
**Why it happens:** Designing interface in isolation without analyzing actual usage patterns in existing code
**How to avoid:**
1. Analyze current `internal/cluster/client.go` methods: `ListPods(ctx, namespace, labelSelector)` and `GetImageSizesFromNodes(ctx)`
2. Map to client-go APIs: `Pods(namespace).List(ctx, ListOptions)` and `Nodes().List(ctx, ListOptions)`
3. Design interface to accept `metav1.ListOptions` (which includes label selectors, field selectors, etc.)
4. Keep wrapper methods in `internal/cluster/client.go` that convert from string parameters to ListOptions
**Warning signs:** Tests are harder to write than production code; need multiple interface methods for slight variations

### Pitfall 2: Breaking Existing Spinner/Progress Logic
**What goes wrong:** Refactoring removes spinner creation/updates from cluster client, breaking user experience
**Why it happens:** Focusing on abstraction purity without considering current UX responsibilities
**How to avoid:**
1. Keep spinner logic in `internal/cluster/client.go` methods (these are the wrappers, not the interface)
2. Interface methods (`pkg/kubernetes/interface.go`) are pure API calls without UI concerns
3. Cluster client methods call interface methods but add spinner/logging around them
4. This preserves current UX while enabling testability
**Warning signs:** User sees no progress indicators; tests fail due to stderr writes

### Pitfall 3: Forgetting to Update All Call Sites
**What goes wrong:** Creating new interface but missing call sites in `pod_analyzer.go`, leaving some code coupled to old implementation
**Why it happens:** Grep for `cluster.NewClient` but miss places that use the returned client
**How to avoid:**
1. Identify all current call sites: `cmd/kubectl-analyze-images/main.go` creates client, passes to `analyzer.NewPodAnalyzerWithConfig()`
2. Change `cluster.NewClient()` to accept `kubernetes.Interface` parameter instead of creating clientset internally
3. Move clientset creation to `main.go`: call `kubernetes.NewClient()`, pass result to `cluster.NewClient(k8sInterface)`
4. Update `pod_analyzer.go` similarly: accept cluster.Client in constructor (which now wraps k8s interface)
5. Verify with `go build` that all imports resolve
**Warning signs:** Compilation errors about mismatched types; imports of old packages remain

### Pitfall 4: Over-Complicated Fake Client
**What goes wrong:** Creating elaborate fake client with custom state management, recorded calls, etc.
**Why it happens:** Not realizing client-go provides `fake.NewSimpleClientset()` that does this automatically
**How to avoid:**
1. Use `fake.NewSimpleClientset(objects...)` which accepts pre-populated runtime.Objects
2. Create test pods/nodes using native Kubernetes types (`corev1.Pod`, `corev1.Node`)
3. Pass them to `NewFakeClient(testPod, testNode)` which wraps the fake clientset
4. Fake client automatically handles List, Get, Create, Update, Delete operations
5. For error simulation, use reactor pattern (if needed, but usually unnecessary for this project)
**Warning signs:** Writing more code in fake.go than in client.go; maintaining test state manually

### Pitfall 5: Testing Implementation Instead of Behavior
**What goes wrong:** Tests verify that interface methods were called, not that correct results were produced
**Why it happens:** Over-using mocks/stubs instead of fake implementations
**How to avoid:**
1. Use fake client to create real test data: `fakeClient := kubernetes.NewFakeClient(testPod1, testPod2)`
2. Call methods under test with real inputs: `pods, err := client.ListPods(ctx, "default", "")`
3. Assert on outputs and behavior: `assert.Len(t, pods, 2)`, `assert.Equal(t, "test-pod", pods[0].Name)`
4. Don't assert on internal implementation details (number of API calls, etc.)
**Warning signs:** Tests use `mock.AssertCalled()` extensively; tests break when refactoring internals without changing behavior

### Pitfall 6: Not Handling Empty Namespace for All Namespaces
**What goes wrong:** Interface design assumes namespace is always provided, breaking "list all pods in all namespaces" use case
**Why it happens:** Not analyzing current usage where empty string means all namespaces
**How to avoid:**
1. Current code uses `c.clientset.CoreV1().Pods(namespace).List(ctx, opts)` where empty namespace means all namespaces
2. Interface method signature: `ListPods(ctx context.Context, namespace string, opts metav1.ListOptions)` handles this naturally
3. Client-go's API client already treats empty namespace as "all namespaces"
4. Test this: `fakeClient.ListPods(ctx, "", metav1.ListOptions{})` should return pods from all namespaces
**Warning signs:** Tests only use explicit namespaces; production code with empty namespace breaks after refactor

## Code Examples

Verified patterns from official sources and current codebase analysis:

### Creating Test Data
```go
// Source: client-go testing patterns
func createTestPod(name, namespace, image string) *corev1.Pod {
    return &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
            Labels: map[string]string{
                "app": "test",
            },
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name:  "container-1",
                    Image: image,
                },
            },
        },
    }
}

func createTestNode(name string, images map[string]int64) *corev1.Node {
    nodeImages := make([]corev1.ContainerImage, 0, len(images))
    for imageName, size := range images {
        nodeImages = append(nodeImages, corev1.ContainerImage{
            Names:     []string{imageName},
            SizeBytes: size,
        })
    }

    return &corev1.Node{
        ObjectMeta: metav1.ObjectMeta{
            Name: name,
        },
        Status: corev1.NodeStatus{
            Images: nodeImages,
        },
    }
}
```

### Table-Driven Test with Fake Client
```go
// Source: Standard Go testing pattern with testify
func TestClient_ListPods_WithLabelSelector(t *testing.T) {
    tests := []struct {
        name          string
        pods          []*corev1.Pod
        namespace     string
        labelSelector string
        expectedCount int
        expectedNames []string
    }{
        {
            name: "single namespace with label",
            pods: []*corev1.Pod{
                createTestPod("pod1", "default", "nginx:1.21"),
                createTestPod("pod2", "kube-system", "coredns:1.8"),
            },
            namespace:     "default",
            labelSelector: "app=test",
            expectedCount: 1,
            expectedNames: []string{"pod1"},
        },
        {
            name: "all namespaces",
            pods: []*corev1.Pod{
                createTestPod("pod1", "default", "nginx:1.21"),
                createTestPod("pod2", "kube-system", "coredns:1.8"),
            },
            namespace:     "",  // Empty means all namespaces
            labelSelector: "",
            expectedCount: 2,
            expectedNames: []string{"pod1", "pod2"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange: Create fake client with test data
            objects := make([]runtime.Object, len(tt.pods))
            for i, pod := range tt.pods {
                objects[i] = pod
            }
            fakeClient := kubernetes.NewFakeClient(objects...)

            // Act: Call method under test
            opts := metav1.ListOptions{}
            if tt.labelSelector != "" {
                opts.LabelSelector = tt.labelSelector
            }
            podList, err := fakeClient.ListPods(context.Background(), tt.namespace, opts)

            // Assert: Verify results
            require.NoError(t, err)
            assert.Len(t, podList.Items, tt.expectedCount)

            actualNames := make([]string, len(podList.Items))
            for i, pod := range podList.Items {
                actualNames[i] = pod.Name
            }
            assert.ElementsMatch(t, tt.expectedNames, actualNames)
        })
    }
}
```

### Testing GetImageSizesFromNodes with Fake Client
```go
// Source: Current implementation pattern + fake client
func TestClient_GetImageSizesFromNodes(t *testing.T) {
    // Arrange: Create test nodes with images
    node1 := createTestNode("node1", map[string]int64{
        "nginx:1.21":      104857600,  // 100 MB
        "postgres:13.4":   314572800,  // 300 MB
    })
    node2 := createTestNode("node2", map[string]int64{
        "nginx:1.21":      104857600,  // Same image on different node
        "redis:6.2":       52428800,   // 50 MB
    })

    fakeK8sClient := kubernetes.NewFakeClient(node1, node2)
    client := cluster.NewClient(fakeK8sClient)

    // Act: Get image sizes
    imageSizes, _, err := client.GetImageSizesFromNodes(context.Background())

    // Assert: Verify results
    require.NoError(t, err)
    assert.Len(t, imageSizes, 3, "Should have 3 unique images")
    assert.Equal(t, int64(104857600), imageSizes["nginx:1.21"])
    assert.Equal(t, int64(314572800), imageSizes["postgres:13.4"])
    assert.Equal(t, int64(52428800), imageSizes["redis:6.2"])
}
```

### Migration Pattern for main.go
```go
// Source: Dependency injection refactor
// BEFORE:
func runAnalyze(...) error {
    analyzer, err := analyzer.NewPodAnalyzerWithConfig(config, kubeContext)
    // ...
}

// AFTER:
func runAnalyze(...) error {
    // Create Kubernetes client (interface)
    k8sClient, err := kubernetes.NewClient(kubeContext)
    if err != nil {
        return fmt.Errorf("failed to create kubernetes client: %w", err)
    }

    // Create cluster client (wrapper with spinner logic)
    clusterClient := cluster.NewClient(k8sClient)

    // Create analyzer with cluster client
    analyzer := analyzer.NewPodAnalyzer(clusterClient, config)

    // ...
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Direct clientset usage | Interface abstraction | Kubernetes ecosystem ~2018+ | Testability, mockability, dependency injection |
| Custom mock implementations | client-go/fake package | client-go v0.11.0+ (2018) | Reduced boilerplate, better test reliability |
| Context-free APIs | Context-first signatures | Go 1.7+ (2016), Kubernetes v1.11+ (2018) | Cancellation, timeouts, request-scoped values |
| Large interfaces | Minimal focused interfaces | Interface segregation principle | Easier testing, clearer dependencies |

**Deprecated/outdated:**
- **Manual mock structs:** Before fake package existed, developers wrote custom mocks. Now client-go provides production-grade fakes.
- **Testing with real clusters:** Some projects used integration tests exclusively. Modern practice: unit tests with fake client, integration tests with real cluster (if needed).
- **Global client variables:** Old code used package-level `var Client *kubernetes.Clientset`. Modern practice: dependency injection via constructors.

## Open Questions

1. **Should interface methods accept higher-level types or raw client-go types?**
   - What we know: Current cluster.Client has methods like `ListPods(ctx, namespace, labelSelector string)` that accept strings
   - What's unclear: Should kubernetes.Interface accept strings and construct ListOptions internally, or accept ListOptions directly?
   - Recommendation: Interface accepts `metav1.ListOptions` (lower level, more flexible). Cluster client keeps string-based methods as wrappers. This preserves current API while allowing full control in tests.

2. **Should we test with fake client AND real cluster (integration tests)?**
   - What we know: Fake client covers unit testing needs, tests run fast without dependencies
   - What's unclear: Do we need integration tests that actually connect to a cluster?
   - Recommendation: Start with fake client only (covers Phase 2 goals). Real cluster testing is out of scope for Phase 2 but could be future enhancement (Phase N: Integration Testing).

3. **How much test coverage is realistic for Phase 2?**
   - What we know: Phase 1 achieved 45.8% coverage, success criteria specifies >60% for cluster operations
   - What's unclear: Should we aim for 60% overall or 60% specifically for cluster/analyzer packages?
   - Recommendation: Target 60% coverage for `internal/cluster/` and `internal/analyzer/` packages specifically (the ones being refactored). Overall project coverage will increase naturally.

## Migration Checklist

This checklist ensures all code is refactored to use the interface abstraction:

### 1. Create Interface Layer
- [ ] Create `pkg/kubernetes/interface.go` with `Interface` definition
- [ ] Create `pkg/kubernetes/client.go` with real implementation
- [ ] Create `pkg/kubernetes/fake.go` with test implementation
- [ ] Add basic tests in `pkg/kubernetes/client_test.go`

### 2. Refactor internal/cluster/client.go
- [ ] Change constructor signature: `func NewClient(k8sClient kubernetes.Interface) *Client`
- [ ] Update struct: `type Client struct { k8sClient kubernetes.Interface }`
- [ ] Refactor `ListPods()` to call `c.k8sClient.ListPods()` (keep spinner logic)
- [ ] Refactor `GetImageSizesFromNodes()` to call `c.k8sClient.ListNodes()` (keep spinner logic)
- [ ] Verify pager still works with new interface methods

### 3. Refactor internal/analyzer/pod_analyzer.go
- [ ] Update constructor to accept `*cluster.Client` instead of creating it internally
- [ ] Remove `cluster.NewClient(contextName)` call from `NewPodAnalyzerWithConfig()`
- [ ] Accept cluster client as parameter: `func NewPodAnalyzerWithConfig(clusterClient *cluster.Client, config *types.AnalysisConfig) *PodAnalyzer`

### 4. Update cmd/kubectl-analyze-images/main.go
- [ ] Create kubernetes client: `k8sClient, err := kubernetes.NewClient(kubeContext)`
- [ ] Create cluster client: `clusterClient := cluster.NewClient(k8sClient)`
- [ ] Create analyzer with cluster client: `analyzer := analyzer.NewPodAnalyzer(clusterClient, config)`
- [ ] Verify `go build` succeeds

### 5. Add Unit Tests
- [ ] Create `internal/cluster/client_test.go` with tests using fake client
- [ ] Test `ListPods()` with various namespaces and label selectors
- [ ] Test `GetImageSizesFromNodes()` with multiple nodes
- [ ] Test `GetUniqueImages()` with various pod sets
- [ ] Create `internal/analyzer/pod_analyzer_test.go` with tests
- [ ] Test `AnalyzePods()` with fake cluster client
- [ ] Verify test coverage >60% for cluster and analyzer packages

### 6. Verification
- [ ] Run `go test ./...` - all tests pass
- [ ] Run `make build` - compilation succeeds
- [ ] Run manual test: `./kubectl-analyze-images -n default` - works as before
- [ ] Check test coverage: `go test -cover ./internal/cluster ./internal/analyzer`
- [ ] Verify no regressions: spinners, output format, performance

## Sources

### Primary (HIGH confidence)
- **client-go v0.29.0 documentation** - Interface patterns, fake client usage
  - `go doc k8s.io/client-go/kubernetes/fake` - Verified fake package API
  - `go doc k8s.io/client-go/tools/pager` - Verified pager compatibility with runtime.Object
- **Current codebase analysis** - internal/cluster/client.go (268 lines), internal/analyzer/pod_analyzer.go
  - Analyzed exact method signatures, parameters, return types
  - Identified spinner/progress logic that must be preserved
  - Mapped current string parameters to metav1.ListOptions usage
- **Go standard library** - Context patterns, interface design
  - Context-first parameter pattern from Go 1.7+ standard library
  - Interface segregation from effective Go practices

### Secondary (MEDIUM confidence)
- **Phase 1 research** - Testing infrastructure, testify usage patterns
  - Established table-driven test patterns already used in pkg/util/image_test.go
  - Confirmed testify v1.11.1 is available (upgraded in Phase 1)
- **Kubernetes ecosystem patterns** - controller-runtime, kubectl plugins
  - Common pattern: define minimal interface, real + fake implementations
  - Standard practice: dependency injection in constructors

### Tertiary (LOW confidence)
- None - all research verified with local tools or official documentation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All dependencies already in go.mod, verified with `go doc`
- Architecture: HIGH - Patterns based on current code analysis + official client-go docs
- Pitfalls: HIGH - Derived from analyzing exact refactor requirements and current implementation

**Research date:** 2026-02-09
**Valid until:** 2026-03-09 (30 days - stable domain, client-go v0.29.0 is mature release)

**Key insight:** This refactor is low-risk because:
1. No new dependencies needed (client-go fake already available)
2. Interface maps 1:1 to current Client methods (minimal design decisions)
3. Fake client handles all edge cases automatically (no custom mock logic)
4. Current spinner/progress logic preserved in cluster.Client wrappers (not in interface)
5. Backward compatibility maintained (existing functionality unchanged, just testable now)
