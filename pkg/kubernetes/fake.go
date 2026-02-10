package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

// FakeClient implements Interface using a fake Kubernetes clientset for testing.
type FakeClient struct {
	clientset *fake.Clientset
}

// Compile-time assertion that FakeClient implements Interface
var _ Interface = (*FakeClient)(nil)

// NewFakeClient creates a new fake Kubernetes client for testing.
// Accepts variadic runtime.Object arguments to pre-populate the fake clientset.
// Returns Interface, not *FakeClient, to match the production constructor signature.
func NewFakeClient(objects ...runtime.Object) Interface {
	return &FakeClient{
		clientset: fake.NewSimpleClientset(objects...),
	}
}

// ListPods lists pods in the given namespace with the given options.
func (f *FakeClient) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
	return f.clientset.CoreV1().Pods(namespace).List(ctx, opts)
}

// ListNodes lists nodes in the cluster with the given options.
func (f *FakeClient) ListNodes(ctx context.Context, opts metav1.ListOptions) (*corev1.NodeList, error) {
	return f.clientset.CoreV1().Nodes().List(ctx, opts)
}
