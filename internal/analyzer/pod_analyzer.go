package analyzer

import (
	"context"
	"fmt"

	"github.com/rnathani/kubectl-analyze-images/internal/cluster"
	"github.com/rnathani/kubectl-analyze-images/internal/registry"
	"github.com/rnathani/kubectl-analyze-images/pkg/types"
)

// PodAnalyzer coordinates pod and image analysis
type PodAnalyzer struct {
	clusterClient  *cluster.Client
	registryClient *registry.Client
}

// NewPodAnalyzer creates a new pod analyzer
func NewPodAnalyzer() (*PodAnalyzer, error) {
	clusterClient, err := cluster.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster client: %w", err)
	}

	registryClient := registry.NewClient()

	return &PodAnalyzer{
		clusterClient:  clusterClient,
		registryClient: registryClient,
	}, nil
}

// AnalyzePods analyzes pods and their images
func (pa *PodAnalyzer) AnalyzePods(ctx context.Context, namespace string, labelSelector string) (*types.ImageAnalysis, error) {
	// Get pods from cluster
	pods, err := pa.clusterClient.ListPods(ctx, namespace, labelSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Get unique images
	uniqueImageNames := pa.clusterClient.GetUniqueImages(pods)

	// Analyze each unique image
	var images []types.Image
	totalSize := int64(0)

	for imageName := range uniqueImageNames {
		image, err := pa.registryClient.GetImageInfo(ctx, imageName)
		if err != nil {
			// Log error but continue with other images
			fmt.Printf("Warning: failed to analyze image %s: %v\n", imageName, err)
			continue
		}

		images = append(images, *image)
		totalSize += image.Size
	}

	// Calculate unique size (for now, same as total size)
	// In later phases, we'll implement layer deduplication
	uniqueSize := totalSize

	return &types.ImageAnalysis{
		Images:     images,
		TotalSize:  totalSize,
		UniqueSize: uniqueSize,
	}, nil
}
