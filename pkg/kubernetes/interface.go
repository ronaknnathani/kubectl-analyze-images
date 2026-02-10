package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Interface defines the contract for Kubernetes cluster operations.
// Both real and fake implementations satisfy this interface.
type Interface interface {
	ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error)
	ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error)
}
