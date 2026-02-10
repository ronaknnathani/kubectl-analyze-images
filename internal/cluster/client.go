package cluster

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/pager"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/kubernetes"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
)

// Client represents a Kubernetes cluster client
type Client struct {
	k8sClient kubernetes.Interface
}

// NewClient creates a new Kubernetes client
func NewClient(k8sClient kubernetes.Interface) *Client {
	return &Client{
		k8sClient: k8sClient,
	}
}

// ListPods lists pods with optional filters and performance metrics using pager
func (c *Client) ListPods(ctx context.Context, namespace, labelSelector string) ([]types.Pod, *types.PerformanceMetrics, error) {
	// Create and start spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	if namespace == "" {
		s.Suffix = " Querying pods from cluster (namespace: All)..."
	} else {
		s.Suffix = fmt.Sprintf(" Querying pods from cluster (namespace: %s)...", namespace)
	}
	_ = s.Color("cyan")
	s.Start()
	defer s.Stop()

	startTime := time.Now()

	var allPods []types.Pod
	var totalPods int

	// List options with ResourceVersion=0 for watch cache optimization
	listOptions := metav1.ListOptions{
		ResourceVersion: "0", // Use watch cache for better performance
	}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	// Use pager to efficiently list all pods
	pager := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
		return c.k8sClient.ListPods(ctx, namespace, opts)
	})

	// Set page size for efficient pagination
	pager.PageSize = 1000

	// List all pods using pager
	err := pager.EachListItem(ctx, listOptions, func(obj runtime.Object) error {
		pod := obj.(*corev1.Pod)
		allPods = append(allPods, types.FromK8sPod(pod))
		totalPods = len(allPods)

		// Update spinner with progress every 100 pods
		if totalPods%100 == 0 {
			s.Suffix = fmt.Sprintf(" Querying pods from cluster (namespace: %s)... %d pods found",
				namespaceDisplay(namespace), totalPods)
		}

		return nil
	})

	if err != nil {
		s.Stop() // Stop spinner before returning error
		return nil, nil, fmt.Errorf("failed to list pods: %w", err)
	}

	podQueryTime := time.Since(startTime)
	s.Stop() // Stop spinner before success message

	// Show success message with pod count
	if namespace == "" {
		fmt.Fprintf(os.Stderr, "✓ Found %d pods across all namespaces (query time: %v)\n", totalPods, podQueryTime)
	} else {
		fmt.Fprintf(os.Stderr, "✓ Found %d pods in namespace %s (query time: %v)\n", totalPods, namespace, podQueryTime)
	}

	metrics := &types.PerformanceMetrics{
		PodQueryTime: podQueryTime,
	}

	return allPods, metrics, nil
}

// GetImageSizesFromNodes gets image sizes from node status
func (c *Client) GetImageSizesFromNodes(ctx context.Context) (map[string]int64, *types.PerformanceMetrics, error) {
	// Create and start spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = " Querying image sizes from nodes..."
	_ = s.Color("cyan")
	s.Start()
	defer s.Stop()

	startTime := time.Now()

	// List options with ResourceVersion=0 for watch cache optimization
	listOptions := metav1.ListOptions{
		ResourceVersion: "0", // Use watch cache for better performance
	}

	// Use pager to efficiently list all nodes
	pager := pager.New(func(ctx context.Context, opts metav1.ListOptions) (runtime.Object, error) {
		return c.k8sClient.ListNodes(ctx, opts)
	})

	// Set page size for efficient pagination
	pager.PageSize = 1000

	// Extract image sizes from node status
	imageSizes := make(map[string]int64)
	var totalImages int
	var totalNodes int

	// List all nodes using pager
	err := pager.EachListItem(ctx, listOptions, func(obj runtime.Object) error {
		node := obj.(*corev1.Node)
		totalNodes++

		for _, image := range node.Status.Images {
			if len(image.Names) > 0 {
				// Select the best canonical name
				imageName := selectBestImageName(image.Names)
				imageSizes[imageName] = image.SizeBytes
				totalImages++
			}
		}

		// Update spinner with progress every 10 nodes
		if totalNodes%10 == 0 {
			s.Suffix = fmt.Sprintf(" Querying image sizes from nodes... %d nodes processed", totalNodes)
		}

		return nil
	})

	if err != nil {
		s.Stop()
		return nil, nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodeQueryTime := time.Since(startTime)
	s.Stop()

	fmt.Fprintf(os.Stderr, "✓ Found %d unique images from %d nodes (query time: %v)\n",
		len(imageSizes), totalNodes, nodeQueryTime)

	metrics := &types.PerformanceMetrics{
		NodeQueryTime: nodeQueryTime,
	}

	return imageSizes, metrics, nil
}

// namespaceDisplay returns a display name for the namespace
func namespaceDisplay(namespace string) string {
	if namespace == "" {
		return "All"
	}
	return namespace
}

// GetUniqueImages extracts unique images from pods
func (c *Client) GetUniqueImages(pods []types.Pod) map[string]bool {
	uniqueImages := make(map[string]bool)

	for _, pod := range pods {
		for _, image := range pod.Images {
			uniqueImages[image] = true
		}
	}

	return uniqueImages
}

// selectBestImageName selects the best canonical name from a list of image names
// Prefers names without SHA digests, then falls back to the first name
func selectBestImageName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return names[0]
	}

	// Look for names without SHA digests first
	for _, name := range names {
		if !containsSHA(name) {
			return name
		}
	}

	// If all names contain SHA, return the first one
	return names[0]
}

// containsSHA checks if an image name contains a SHA digest
func containsSHA(name string) bool {
	// SHA digests are typically 64 characters long and contain only hex characters
	// They often appear after @ or : in image names
	parts := strings.Split(name, "@")
	if len(parts) > 1 {
		digest := parts[1]
		if len(digest) == 64 && isHexString(digest) {
			return true
		}
	}

	// Also check for SHA in tag format (less common but possible)
	parts = strings.Split(name, ":")
	if len(parts) > 1 {
		tag := parts[len(parts)-1]
		if len(tag) == 64 && isHexString(tag) {
			return true
		}
	}

	return false
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
