package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// testPod creates a test pod with the given name, namespace, and container images.
func testPod(name, namespace string, images ...string) *corev1.Pod {
	containers := make([]corev1.Container, len(images))
	for i, img := range images {
		containers[i] = corev1.Container{Name: fmt.Sprintf("c%d", i), Image: img}
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       corev1.PodSpec{Containers: containers},
	}
}

// testNode creates a test node with the given name and image sizes.
func testNode(name string, images map[string]int64) *corev1.Node {
	nodeImages := make([]corev1.ContainerImage, 0, len(images))
	for img, size := range images {
		nodeImages = append(nodeImages, corev1.ContainerImage{
			Names:     []string{img},
			SizeBytes: size,
		})
	}
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     corev1.NodeStatus{Images: nodeImages},
	}
}

// --- Complete tests ---

func TestAnalyzeOptions_Complete(t *testing.T) {
	t.Run("defaults populated", func(t *testing.T) {
		o := &AnalyzeOptions{
			KubernetesClient: kubernetes.NewFakeClient(), // pre-inject to avoid kubeconfig requirement
		}
		err := o.Complete()
		require.NoError(t, err)
		assert.Equal(t, "table", o.OutputFormat)
		assert.Equal(t, 25, o.TopImages)
		assert.NotNil(t, o.Out)
		assert.NotNil(t, o.ErrOut)
	})

	t.Run("preserves explicit values", func(t *testing.T) {
		buf := &bytes.Buffer{}
		o := &AnalyzeOptions{
			OutputFormat:     "json",
			TopImages:        10,
			Out:              buf,
			KubernetesClient: kubernetes.NewFakeClient(),
		}
		err := o.Complete()
		require.NoError(t, err)
		assert.Equal(t, "json", o.OutputFormat)
		assert.Equal(t, 10, o.TopImages)
		assert.Equal(t, buf, o.Out) // same buffer instance
	})

	t.Run("skips kubernetes client when pre-injected", func(t *testing.T) {
		fakeClient := kubernetes.NewFakeClient()
		o := &AnalyzeOptions{
			KubernetesClient: fakeClient,
		}
		err := o.Complete()
		require.NoError(t, err)
		assert.Equal(t, fakeClient, o.KubernetesClient) // same instance
	})
}

// --- Validate tests ---

func TestAnalyzeOptions_Validate(t *testing.T) {
	tests := []struct {
		name        string
		opts        AnalyzeOptions
		expectError string
	}{
		{name: "valid table format", opts: AnalyzeOptions{OutputFormat: "table", TopImages: 25}},
		{name: "valid json format", opts: AnalyzeOptions{OutputFormat: "json", TopImages: 10}},
		{name: "invalid output format", opts: AnalyzeOptions{OutputFormat: "yaml", TopImages: 25}, expectError: "invalid output format"},
		{name: "topImages zero", opts: AnalyzeOptions{OutputFormat: "table", TopImages: 0}, expectError: "must be at least 1"},
		{name: "topImages negative", opts: AnalyzeOptions{OutputFormat: "table", TopImages: -5}, expectError: "must be at least 1"},
		{name: "topImages one is valid", opts: AnalyzeOptions{OutputFormat: "table", TopImages: 1}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate()
			if tc.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// --- Run tests ---

func TestAnalyzeOptions_Run_TableOutput(t *testing.T) {
	pod1 := testPod("pod1", "default", "nginx:1.21")
	pod2 := testPod("pod2", "default", "redis:6.2")
	node := testNode("node1", map[string]int64{
		"nginx:1.21": 100000000, // 100MB
		"redis:6.2":  50000000,  // 50MB
	})

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	o := &AnalyzeOptions{
		Namespace:        "default",
		OutputFormat:     "table",
		TopImages:        25,
		ShowHistogram:    true,
		NoColor:          true,
		KubernetesClient: kubernetes.NewFakeClient(pod1, pod2, node),
		Out:              out,
		ErrOut:           errOut,
	}

	err := o.Run(context.Background())
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Analyzing images in namespace: default")
	assert.Contains(t, output, "Image Analysis Summary")
	assert.Contains(t, output, "nginx:1.21")
	assert.Contains(t, output, "redis:6.2")
}

func TestAnalyzeOptions_Run_JSONOutput(t *testing.T) {
	pod1 := testPod("pod1", "default", "nginx:1.21")
	node := testNode("node1", map[string]int64{
		"nginx:1.21": 100000000,
	})

	out := &bytes.Buffer{}

	o := &AnalyzeOptions{
		Namespace:        "default",
		OutputFormat:     "json",
		TopImages:        25,
		KubernetesClient: kubernetes.NewFakeClient(pod1, node),
		Out:              out,
		ErrOut:           &bytes.Buffer{},
	}

	err := o.Run(context.Background())
	require.NoError(t, err)

	// The output contains "Analyzing images..." header lines followed by JSON
	// Find the JSON portion (starts with '{')
	output := out.String()
	jsonStart := strings.Index(output, "{")
	require.True(t, jsonStart >= 0, "expected JSON output, got: %s", output)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(output[jsonStart:]), &result)
	require.NoError(t, err, "failed to parse JSON output")

	// Verify structure
	summary, ok := result["summary"].(map[string]interface{})
	require.True(t, ok, "expected summary object in JSON")
	assert.Equal(t, float64(1), summary["totalImages"])
}

func TestAnalyzeOptions_Run_AllNamespaces(t *testing.T) {
	// No pods needed for all-namespaces mode -- analyzer uses node images directly
	node := testNode("node1", map[string]int64{
		"nginx:1.21":  100000000,
		"redis:6.2":   50000000,
		"postgres:13": 200000000,
	})

	out := &bytes.Buffer{}

	o := &AnalyzeOptions{
		Namespace:        "", // empty = all namespaces
		OutputFormat:     "table",
		TopImages:        25,
		NoColor:          true,
		KubernetesClient: kubernetes.NewFakeClient(node),
		Out:              out,
		ErrOut:           &bytes.Buffer{},
	}

	err := o.Run(context.Background())
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Analyzing images in namespace: All")
	assert.Contains(t, output, "nginx:1.21")
	assert.Contains(t, output, "redis:6.2")
	assert.Contains(t, output, "postgres:13")
}

func TestAnalyzeOptions_Run_WithLabelSelector(t *testing.T) {
	// Create pods with labels
	pod1 := testPod("pod1", "default", "nginx:1.21")
	pod1.Labels = map[string]string{"app": "web"}

	pod2 := testPod("pod2", "default", "redis:6.2")
	pod2.Labels = map[string]string{"app": "cache"}

	node := testNode("node1", map[string]int64{
		"nginx:1.21": 100000000,
		"redis:6.2":  50000000,
	})

	out := &bytes.Buffer{}

	objects := []runtime.Object{pod1, pod2, node}

	o := &AnalyzeOptions{
		Namespace:        "default",
		LabelSelector:    "app=web",
		OutputFormat:     "table",
		TopImages:        25,
		NoColor:          true,
		KubernetesClient: kubernetes.NewFakeClient(objects...),
		Out:              out,
		ErrOut:           &bytes.Buffer{},
	}

	err := o.Run(context.Background())
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Using label selector: app=web")
	// With label selector app=web, only nginx:1.21 should be analyzed
	assert.Contains(t, output, "nginx:1.21")
}
