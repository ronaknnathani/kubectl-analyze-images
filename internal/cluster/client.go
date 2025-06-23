package cluster

import (
	"context"
	"fmt"

	"github.com/rnathani/kubectl-analyze-images/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client represents a Kubernetes cluster client
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewClient creates a new Kubernetes client
func NewClient() (*Client, error) {
	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

// ListPods lists pods with optional filters
func (c *Client) ListPods(ctx context.Context, namespace string, labelSelector string) ([]types.Pod, error) {
	var pods []types.Pod

	// List options
	listOptions := metav1.ListOptions{}
	if labelSelector != "" {
		listOptions.LabelSelector = labelSelector
	}

	// Get pods
	k8sPods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Convert to our types
	for _, k8sPod := range k8sPods.Items {
		pod := types.FromK8sPod(&k8sPod)
		pods = append(pods, pod)
	}

	return pods, nil
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
