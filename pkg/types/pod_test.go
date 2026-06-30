package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFromK8sPod_SegregatesContainerImages(t *testing.T) {
	pod := FromK8sPod(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "app", Image: "app:v1"},
			},
			InitContainers: []corev1.Container{
				{Name: "setup", Image: "setup:v1"},
			},
		},
	})

	assert.Equal(t, []string{"app:v1"}, pod.ContainerImages)
	assert.Equal(t, []string{"setup:v1"}, pod.InitContainerImages)
	assert.ElementsMatch(t, []string{"app:v1", "setup:v1"}, pod.Images)
}
