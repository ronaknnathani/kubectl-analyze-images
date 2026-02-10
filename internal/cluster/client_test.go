package cluster

import (
	"context"
	"fmt"
	"testing"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/kubernetes"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// createTestPod creates a test pod with the given name, namespace, and images
func createTestPod(name, namespace string, images ...string) *corev1.Pod {
	containers := make([]corev1.Container, len(images))
	for i, img := range images {
		containers[i] = corev1.Container{Name: fmt.Sprintf("container-%d", i), Image: img}
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec:       corev1.PodSpec{Containers: containers},
	}
}

// createTestNode creates a test node with the given name and images
func createTestNode(name string, images map[string]int64) *corev1.Node {
	nodeImages := make([]corev1.ContainerImage, 0, len(images))
	for imgName, size := range images {
		nodeImages = append(nodeImages, corev1.ContainerImage{
			Names:     []string{imgName},
			SizeBytes: size,
		})
	}
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     corev1.NodeStatus{Images: nodeImages},
	}
}

func TestClient_ListPods(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		pods          []*corev1.Pod
		namespace     string
		labelSelector string
		expectedCount int
		checkPod      func(t *testing.T, pods []types.Pod)
	}{
		{
			name: "single pod in default namespace",
			pods: []*corev1.Pod{
				createTestPod("pod1", "default", "nginx:1.21"),
			},
			namespace:     "default",
			expectedCount: 1,
			checkPod: func(t *testing.T, pods []types.Pod) {
				assert.Equal(t, "pod1", pods[0].Name)
				assert.Equal(t, "default", pods[0].Namespace)
				assert.Len(t, pods[0].Images, 1)
				assert.Equal(t, "nginx:1.21", pods[0].Images[0])
			},
		},
		{
			name: "pods across namespaces with empty namespace (all)",
			pods: []*corev1.Pod{
				createTestPod("pod1", "default", "nginx:1.21"),
				createTestPod("pod2", "kube-system", "coredns:1.9"),
			},
			namespace:     "",
			expectedCount: 2,
			checkPod: func(t *testing.T, pods []types.Pod) {
				assert.Len(t, pods, 2)
			},
		},
		{
			name: "no pods in namespace",
			pods: []*corev1.Pod{
				createTestPod("pod1", "default", "nginx:1.21"),
			},
			namespace:     "other-ns",
			expectedCount: 0,
		},
		{
			name: "pod with multiple containers",
			pods: []*corev1.Pod{
				createTestPod("pod1", "default", "nginx:1.21", "redis:6.2"),
			},
			namespace:     "default",
			expectedCount: 1,
			checkPod: func(t *testing.T, pods []types.Pod) {
				assert.Len(t, pods[0].Images, 2)
				assert.Contains(t, pods[0].Images, "nginx:1.21")
				assert.Contains(t, pods[0].Images, "redis:6.2")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert pods to runtime.Object slice
			objects := make([]runtime.Object, len(tt.pods))
			for i, pod := range tt.pods {
				objects[i] = pod
			}

			// Create fake client and cluster client
			fakeK8s := kubernetes.NewFakeClient(objects...)
			clusterClient := NewClient(fakeK8s)

			// List pods
			pods, metrics, err := clusterClient.ListPods(ctx, tt.namespace, tt.labelSelector)

			// Assert no error
			require.NoError(t, err)
			assert.NotNil(t, metrics)

			// Assert pod count
			assert.Len(t, pods, tt.expectedCount)

			// Run custom checks if provided
			if tt.checkPod != nil && len(pods) > 0 {
				tt.checkPod(t, pods)
			}
		})
	}
}

func TestClient_GetImageSizesFromNodes(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		nodes         []*corev1.Node
		expectedSizes map[string]int64
		expectedCount int
	}{
		{
			name: "single node with images",
			nodes: []*corev1.Node{
				createTestNode("node1", map[string]int64{
					"nginx:1.21": 100000000,
					"redis:6.2":  50000000,
				}),
			},
			expectedSizes: map[string]int64{
				"nginx:1.21": 100000000,
				"redis:6.2":  50000000,
			},
			expectedCount: 2,
		},
		{
			name: "multiple nodes with overlapping images",
			nodes: []*corev1.Node{
				createTestNode("node1", map[string]int64{
					"nginx:1.21": 100000000,
					"redis:6.2":  50000000,
				}),
				createTestNode("node2", map[string]int64{
					"nginx:1.21": 110000000, // Different size (last write wins)
					"postgres:13": 200000000,
				}),
			},
			expectedSizes: map[string]int64{
				"nginx:1.21":  110000000, // Last write wins
				"redis:6.2":   50000000,
				"postgres:13": 200000000,
			},
			expectedCount: 3,
		},
		{
			name: "node with no images",
			nodes: []*corev1.Node{
				createTestNode("node1", map[string]int64{}),
			},
			expectedSizes: map[string]int64{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert nodes to runtime.Object slice
			objects := make([]runtime.Object, len(tt.nodes))
			for i, node := range tt.nodes {
				objects[i] = node
			}

			// Create fake client and cluster client
			fakeK8s := kubernetes.NewFakeClient(objects...)
			clusterClient := NewClient(fakeK8s)

			// Get image sizes
			imageSizes, metrics, err := clusterClient.GetImageSizesFromNodes(ctx)

			// Assert no error
			require.NoError(t, err)
			assert.NotNil(t, metrics)

			// Assert image count
			assert.Len(t, imageSizes, tt.expectedCount)

			// Assert image sizes match expected
			for imgName, expectedSize := range tt.expectedSizes {
				actualSize, exists := imageSizes[imgName]
				assert.True(t, exists, "Image %s should exist", imgName)
				assert.Equal(t, expectedSize, actualSize, "Size mismatch for image %s", imgName)
			}
		})
	}
}

func TestClient_GetUniqueImages(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		pods          []*corev1.Pod
		namespace     string
		expectedCount int
		expectedImages []string
	}{
		{
			name: "unique images from multiple pods",
			pods: []*corev1.Pod{
				createTestPod("pod1", "default", "nginx:1.21", "redis:6.2"),
				createTestPod("pod2", "default", "nginx:1.21", "postgres:13"),
			},
			namespace:     "default",
			expectedCount: 3,
			expectedImages: []string{"nginx:1.21", "redis:6.2", "postgres:13"},
		},
		{
			name:          "empty pod list",
			pods:          []*corev1.Pod{},
			namespace:     "default",
			expectedCount: 0,
			expectedImages: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert pods to runtime.Object slice
			objects := make([]runtime.Object, len(tt.pods))
			for i, pod := range tt.pods {
				objects[i] = pod
			}

			// Create fake client and cluster client
			fakeK8s := kubernetes.NewFakeClient(objects...)
			clusterClient := NewClient(fakeK8s)

			// List pods first
			pods, _, err := clusterClient.ListPods(ctx, tt.namespace, "")
			require.NoError(t, err)

			// Get unique images
			uniqueImages := clusterClient.GetUniqueImages(pods)

			// Assert count
			assert.Len(t, uniqueImages, tt.expectedCount)

			// Assert expected images present
			for _, img := range tt.expectedImages {
				assert.True(t, uniqueImages[img], "Image %s should be present", img)
			}
		})
	}
}

func TestSelectBestImageName(t *testing.T) {
	tests := []struct {
		name     string
		names    []string
		expected string
	}{
		{
			name:     "prefers name without SHA",
			names:    []string{"nginx:1.21", "sha256:" + string(make([]byte, 64))},
			expected: "nginx:1.21",
		},
		{
			name:     "single name returned",
			names:    []string{"nginx:1.21"},
			expected: "nginx:1.21",
		},
		{
			name:     "empty list returns empty",
			names:    []string{},
			expected: "",
		},
		{
			name:     "prefers first non-SHA name",
			names:    []string{"docker.io/nginx@sha256:abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234", "nginx:1.21"},
			expected: "nginx:1.21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := selectBestImageName(tt.names)
			assert.Equal(t, tt.expected, result)
		})
	}
}
