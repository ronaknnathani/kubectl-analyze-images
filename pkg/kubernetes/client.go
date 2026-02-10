package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client implements Interface using a real Kubernetes clientset.
type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// Compile-time assertion that Client implements Interface
var _ Interface = (*Client)(nil)

// NewClient creates a new Kubernetes client that loads kubeconfig.
// If contextName is empty, uses the current context from kubeconfig.
// Returns Interface, not *Client, to enable dependency injection.
func NewClient(contextName string) (Interface, error) {
	// Load kubeconfig with context
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if contextName != "" {
		configOverrides.CurrentContext = contextName
	}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := clientConfig.ClientConfig()
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

// ListPods lists pods in the given namespace with the given options.
func (c *Client) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
	return c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
}

// ListNodes lists nodes in the cluster with the given options.
func (c *Client) ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error) {
	return c.clientset.CoreV1().Nodes().List(ctx, opts)
}
