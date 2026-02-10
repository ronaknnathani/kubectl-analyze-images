package analyzer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ronaknnathani/kubectl-analyze-images/internal/cluster"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/util"
)

// PodAnalyzer coordinates pod and image analysis
type PodAnalyzer struct {
	clusterClient *cluster.Client
	config        *types.AnalysisConfig
}

// NewPodAnalyzer creates a new pod analyzer with custom configuration
func NewPodAnalyzer(clusterClient *cluster.Client, config *types.AnalysisConfig) *PodAnalyzer {
	return &PodAnalyzer{
		clusterClient: clusterClient,
		config:        config,
	}
}

// AnalyzePods analyzes container images from pods
func (pa *PodAnalyzer) AnalyzePods(ctx context.Context, namespace, labelSelector string) (*types.ImageAnalysis, error) {
	overallStart := time.Now()

	var pods []types.Pod
	var perfMetrics *types.PerformanceMetrics
	var err error

	// Only query pods if namespace or label selector is specified
	if namespace != "" || labelSelector != "" {
		pods, perfMetrics, err = pa.clusterClient.ListPods(ctx, namespace, labelSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to list pods: %w", err)
		}
	} else {
		// No filters specified, use all images from nodes
		perfMetrics = &types.PerformanceMetrics{}
	}

	// Get image sizes from node status
	imageSizes, nodeMetrics, err := pa.clusterClient.GetImageSizesFromNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get image sizes from nodes: %w", err)
	}

	// Merge performance metrics
	if perfMetrics == nil {
		perfMetrics = nodeMetrics
	} else {
		perfMetrics.NodeQueryTime = nodeMetrics.NodeQueryTime
	}

	// Start timing image analysis
	imageAnalysisStart := time.Now()

	// Create spinner for image analysis
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	_ = s.Color("cyan")

	// Determine which images to analyze
	var imagesToAnalyze map[string]bool
	if len(pods) > 0 {
		// Use images from pods if we queried pods
		imagesToAnalyze = pa.clusterClient.GetUniqueImages(pods)
	} else {
		// Use all images from nodes if no pod filters
		imagesToAnalyze = make(map[string]bool)
		for imageName := range imageSizes {
			imagesToAnalyze[imageName] = true
		}
	}

	// Start spinner for analysis
	s.Suffix = fmt.Sprintf(" Analyzing %d images...", len(imagesToAnalyze))
	s.Start()

	// Create images from node data
	images := make([]types.Image, 0, len(imagesToAnalyze))
	var totalSize int64
	var processedCount int

	for imageName := range imagesToAnalyze {
		size, exists := imageSizes[imageName]
		if !exists {
			// Image not found in node status, mark as inaccessible
			registry, tag := util.ExtractRegistryAndTag(imageName)
			images = append(images, types.Image{
				Name:         imageName,
				Size:         0,
				Registry:     registry,
				Tag:          tag,
				Inaccessible: true,
			})
		} else {
			// Image found, create entry with size
			registry, tag := util.ExtractRegistryAndTag(imageName)
			images = append(images, types.Image{
				Name:         imageName,
				Size:         size,
				Registry:     registry,
				Tag:          tag,
				Inaccessible: false,
			})
			totalSize += size
		}
		processedCount++
	}

	s.Stop()
	imageAnalysisTime := time.Since(imageAnalysisStart)

	// Show completion message
	fmt.Fprintf(os.Stderr, "✓ Completed analyzing %d images (time: %v)\n", processedCount, imageAnalysisTime)

	// Update performance metrics
	perfMetrics.ImageAnalysisTime = imageAnalysisTime
	perfMetrics.TotalTime = time.Since(overallStart)
	perfMetrics.ImagesProcessed = processedCount

	// Build analysis result
	analysis := &types.ImageAnalysis{
		Images:      images,
		TotalSize:   totalSize,
		UniqueSize:  totalSize, // No deduplication in this approach
		Performance: perfMetrics,
	}

	return analysis, nil
}
