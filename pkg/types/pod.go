package types

import (
	corev1 "k8s.io/api/core/v1"
)

// Pod represents a simplified pod structure for analysis
type Pod struct {
	Name      string
	Namespace string
	Images    []string
}

// PodList represents a collection of pods
type PodList struct {
	Pods []Pod
}

// FromK8sPod converts a Kubernetes pod to our internal Pod type
func FromK8sPod(k8sPod *corev1.Pod) Pod {
	pod := Pod{
		Name:      k8sPod.Name,
		Namespace: k8sPod.Namespace,
		Images:    make([]string, 0),
	}

	// Extract container images
	for _, container := range k8sPod.Spec.Containers {
		if container.Image != "" {
			pod.Images = append(pod.Images, container.Image)
		}
	}

	// Extract init container images
	for _, container := range k8sPod.Spec.InitContainers {
		if container.Image != "" {
			pod.Images = append(pod.Images, container.Image)
		}
	}

	return pod
}
