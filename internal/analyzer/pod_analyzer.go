package analyzer

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
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
	errOut        io.Writer
}

// NewPodAnalyzer creates a new pod analyzer with custom configuration
func NewPodAnalyzer(clusterClient *cluster.Client, config *types.AnalysisConfig, errOut ...io.Writer) *PodAnalyzer {
	w := io.Writer(os.Stderr)
	if len(errOut) > 0 && errOut[0] != nil {
		w = errOut[0]
	}
	return &PodAnalyzer{
		clusterClient: clusterClient,
		config:        config,
		errOut:        w,
	}
}

// AnalyzePods analyzes container images from pods
func (pa *PodAnalyzer) AnalyzePods(ctx context.Context, namespace, labelSelector string) (*types.ImageAnalysis, error) {
	overallStart := time.Now()

	var pods []types.Pod
	var podMetrics *types.PerformanceMetrics
	var inventory *cluster.ImageInventory
	var nodeMetrics *types.PerformanceMetrics
	var podErr, nodeErr error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		pods, podMetrics, podErr = pa.clusterClient.ListPods(ctx, namespace, labelSelector)
	}()
	go func() {
		defer wg.Done()
		inventory, nodeMetrics, nodeErr = pa.clusterClient.GetImageSizesFromNodes(ctx)
	}()
	wg.Wait()

	if podErr != nil {
		return nil, fmt.Errorf("failed to list pods: %w", podErr)
	}
	if nodeErr != nil {
		return nil, fmt.Errorf("failed to get image sizes from nodes: %w", nodeErr)
	}

	perfMetrics := &types.PerformanceMetrics{
		PodQueryTime:  podMetrics.PodQueryTime,
		NodeQueryTime: nodeMetrics.NodeQueryTime,
	}

	// Start timing image analysis
	imageAnalysisStart := time.Now()

	// Create spinner for image analysis
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(pa.errOut))
	if err := s.Color("cyan"); err != nil {
		return nil, fmt.Errorf("failed to color image analysis spinner: %w", err)
	}

	usage := collectImageUsage(pods)
	imagesToAnalyze := make(map[string]struct{})
	for imageName := range usage {
		imagesToAnalyze[imageName] = struct{}{}
	}
	if namespace == "" && labelSelector == "" {
		for imageName := range inventory.DisplayNames {
			imagesToAnalyze[imageName] = struct{}{}
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
		size, exists := inventory.Sizes[imageName]
		imageUsage := usage[imageName]
		registry, tag := util.ExtractRegistryAndTag(imageName)
		image := types.Image{
			Name:               imageName,
			Registry:           registry,
			Tag:                tag,
			PodCount:           imageUsage.podCount(),
			ContainerCount:     imageUsage.ContainerCount,
			InitContainerCount: imageUsage.InitContainerCount,
			NamespaceCount:     imageUsage.namespaceCount(),
			CachedOnNodes:      inventory.CachedOnNodes[imageName],
		}
		if !exists {
			image.Inaccessible = true
			perfMetrics.ImagesInaccessible++
		} else {
			image.Size = size
			totalSize += size
		}
		images = append(images, image)
		processedCount++
	}

	s.Stop()
	imageAnalysisTime := time.Since(imageAnalysisStart)

	// Update performance metrics
	perfMetrics.ImageAnalysisTime = imageAnalysisTime
	perfMetrics.TotalTime = time.Since(overallStart)
	perfMetrics.ImagesProcessed = processedCount

	// Build analysis result
	analysis := &types.ImageAnalysis{
		Images:       images,
		TotalSize:    totalSize,
		UniqueSize:   totalSize, // No deduplication in this approach
		PodsScanned:  len(pods),
		NodesScanned: inventory.NodesScanned,
		Performance:  perfMetrics,
	}
	for _, img := range images {
		if img.PodCount > 0 {
			analysis.ImagesInUse++
		} else {
			analysis.UnusedImages++
		}
	}

	return analysis, nil
}

type imageUsage struct {
	ContainerCount     int
	InitContainerCount int
	pods               map[string]struct{}
	namespaces         map[string]struct{}
}

func collectImageUsage(pods []types.Pod) map[string]imageUsage {
	usage := make(map[string]imageUsage)
	for _, pod := range pods {
		podKey := pod.Namespace + "/" + pod.Name
		for _, imageName := range pod.ContainerImages {
			stat := imageUsageForPod(usage, imageName, podKey, pod.Namespace)
			stat.ContainerCount++
			usage[imageName] = stat
		}
		for _, imageName := range pod.InitContainerImages {
			stat := imageUsageForPod(usage, imageName, podKey, pod.Namespace)
			stat.InitContainerCount++
			usage[imageName] = stat
		}
	}
	return usage
}

func imageUsageForPod(usage map[string]imageUsage, imageName, podKey, namespace string) imageUsage {
	stat := usage[imageName]
	if stat.pods == nil {
		stat.pods = make(map[string]struct{})
		stat.namespaces = make(map[string]struct{})
	}
	stat.pods[podKey] = struct{}{}
	stat.namespaces[namespace] = struct{}{}
	return stat
}

func (iu imageUsage) podCount() int {
	return len(iu.pods)
}

func (iu imageUsage) namespaceCount() int {
	return len(iu.namespaces)
}
