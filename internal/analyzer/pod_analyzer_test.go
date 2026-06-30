package analyzer

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ronaknnathani/kubectl-analyze-images/internal/cluster"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/kubernetes"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
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

func createTestPodWithInitContainers(name, namespace string, images, initImages []string) *corev1.Pod {
	containers := make([]corev1.Container, len(images))
	for i, img := range images {
		containers[i] = corev1.Container{Name: fmt.Sprintf("container-%d", i), Image: img}
	}
	initContainers := make([]corev1.Container, len(initImages))
	for i, img := range initImages {
		initContainers[i] = corev1.Container{Name: fmt.Sprintf("init-container-%d", i), Image: img}
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: corev1.PodSpec{
			Containers:     containers,
			InitContainers: initContainers,
		},
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

func TestPodAnalyzer_AnalyzePods_WithPods(t *testing.T) {
	ctx := context.Background()

	// Create test pods
	pod1 := createTestPod("pod1", "default", "nginx:1.21")
	pod2 := createTestPod("pod2", "default", "redis:6.2")

	// Create test node with image sizes
	node1 := createTestNode("node1", map[string]int64{
		"nginx:1.21": 100000000, // 100MB
		"redis:6.2":  50000000,  // 50MB
	})

	// Create fake client with pods and node
	objects := []runtime.Object{pod1, pod2, node1}
	fakeK8s := kubernetes.NewFakeClient(objects...)

	// Create cluster client and analyzer
	clusterClient := cluster.NewClient(fakeK8s)
	config := types.DefaultAnalysisConfig()
	podAnalyzer := NewPodAnalyzer(clusterClient, config)

	// Analyze pods
	result, err := podAnalyzer.AnalyzePods(ctx, "default", "")

	// Assert no error
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Assert result has 2 images
	assert.Len(t, result.Images, 2)

	// Assert total size is 150MB
	assert.Equal(t, int64(150000000), result.TotalSize)

	// Verify both images have correct properties
	imageMap := make(map[string]types.Image)
	for _, img := range result.Images {
		imageMap[img.Name] = img
	}

	nginxImg, exists := imageMap["nginx:1.21"]
	assert.True(t, exists, "nginx:1.21 should exist")
	assert.Equal(t, int64(100000000), nginxImg.Size)
	assert.Equal(t, "1.21", nginxImg.Tag)
	assert.False(t, nginxImg.Inaccessible)

	redisImg, exists := imageMap["redis:6.2"]
	assert.True(t, exists, "redis:6.2 should exist")
	assert.Equal(t, int64(50000000), redisImg.Size)
	assert.Equal(t, "6.2", redisImg.Tag)
	assert.False(t, redisImg.Inaccessible)
}

func TestPodAnalyzer_AnalyzePods_SegregatesInitContainers(t *testing.T) {
	ctx := context.Background()
	pod1 := createTestPodWithInitContainers("pod1", "default", []string{"app:v1"}, []string{"setup:v1"})
	pod2 := createTestPodWithInitContainers("pod2", "default", []string{"app:v1"}, []string{"app:v1", "setup:v1"})
	node := createTestNode("node1", map[string]int64{
		"app:v1":   100000000,
		"setup:v1": 50000000,
	})

	fakeK8s := kubernetes.NewFakeClient(pod1, pod2, node)
	clusterClient := cluster.NewClient(fakeK8s)
	podAnalyzer := NewPodAnalyzer(clusterClient, types.DefaultAnalysisConfig())

	result, err := podAnalyzer.AnalyzePods(ctx, "default", "")
	require.NoError(t, err)

	imageMap := make(map[string]types.Image)
	for _, img := range result.Images {
		imageMap[img.Name] = img
	}

	appImg := imageMap["app:v1"]
	assert.Equal(t, 2, appImg.PodCount)
	assert.Equal(t, 2, appImg.ContainerCount)
	assert.Equal(t, 1, appImg.InitContainerCount)

	setupImg := imageMap["setup:v1"]
	assert.Equal(t, 2, setupImg.PodCount)
	assert.Equal(t, 0, setupImg.ContainerCount)
	assert.Equal(t, 2, setupImg.InitContainerCount)
}

func TestPodAnalyzer_AnalyzePods_NoNamespace(t *testing.T) {
	ctx := context.Background()

	// Create test node with 3 images
	node1 := createTestNode("node1", map[string]int64{
		"nginx:1.21":  100000000,
		"redis:6.2":   50000000,
		"postgres:13": 200000000,
	})

	// Create fake client with only node
	objects := []runtime.Object{node1}
	fakeK8s := kubernetes.NewFakeClient(objects...)

	// Create cluster client and analyzer
	clusterClient := cluster.NewClient(fakeK8s)
	config := types.DefaultAnalysisConfig()
	podAnalyzer := NewPodAnalyzer(clusterClient, config)

	// Analyze with no namespace (should use all node images)
	result, err := podAnalyzer.AnalyzePods(ctx, "", "")

	// Assert no error
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Assert result has 3 images from node
	assert.Len(t, result.Images, 3)

	// Assert total size
	assert.Equal(t, int64(350000000), result.TotalSize)

	// Verify all images are accessible
	for _, img := range result.Images {
		assert.False(t, img.Inaccessible, "Image %s should be accessible", img.Name)
		assert.Greater(t, img.Size, int64(0), "Image %s should have non-zero size", img.Name)
	}
}

func TestPodAnalyzer_AnalyzePods_MissingImageSize(t *testing.T) {
	ctx := context.Background()

	// Create test pod with custom image
	pod1 := createTestPod("pod1", "default", "custom:latest")

	// Create test node that does NOT have custom:latest
	node1 := createTestNode("node1", map[string]int64{
		"nginx:1.21": 100000000,
	})

	// Create fake client with pod and node
	objects := []runtime.Object{pod1, node1}
	fakeK8s := kubernetes.NewFakeClient(objects...)

	// Create cluster client and analyzer
	clusterClient := cluster.NewClient(fakeK8s)
	config := types.DefaultAnalysisConfig()
	podAnalyzer := NewPodAnalyzer(clusterClient, config)

	// Analyze pods
	result, err := podAnalyzer.AnalyzePods(ctx, "default", "")

	// Assert no error
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Assert result has 1 image
	assert.Len(t, result.Images, 1)

	// Assert the image is marked as inaccessible
	customImg := result.Images[0]
	assert.Equal(t, "custom:latest", customImg.Name)
	assert.True(t, customImg.Inaccessible, "Image should be marked as inaccessible")
	assert.Equal(t, int64(0), customImg.Size, "Inaccessible image should have size 0")
	assert.Equal(t, "latest", customImg.Tag)
}

func TestPodAnalyzer_AnalyzePods_MultipleNamespaces(t *testing.T) {
	ctx := context.Background()

	// Create test pods in different namespaces
	pod1 := createTestPod("pod1", "default", "nginx:1.21")
	pod2 := createTestPod("pod2", "kube-system", "coredns:1.9")

	// Create test node with image sizes
	node1 := createTestNode("node1", map[string]int64{
		"nginx:1.21":  100000000,
		"coredns:1.9": 40000000,
	})

	// Create fake client with pods and node
	objects := []runtime.Object{pod1, pod2, node1}
	fakeK8s := kubernetes.NewFakeClient(objects...)

	// Create cluster client and analyzer
	clusterClient := cluster.NewClient(fakeK8s)
	config := types.DefaultAnalysisConfig()
	podAnalyzer := NewPodAnalyzer(clusterClient, config)

	// Analyze only default namespace
	result, err := podAnalyzer.AnalyzePods(ctx, "default", "")

	// Assert no error
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Assert result has 1 image (only from default namespace)
	assert.Len(t, result.Images, 1)
	assert.Equal(t, "nginx:1.21", result.Images[0].Name)

	// Analyze all namespaces (no namespace filter)
	result, err = podAnalyzer.AnalyzePods(ctx, "", "")

	// Assert no error
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Assert result has 2 images (all from node)
	assert.Len(t, result.Images, 2)
}
