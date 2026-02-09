package types

import (
	"sort"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/util"
)

// Image represents a container image with its metadata
type Image struct {
	Name         string
	Size         int64
	Registry     string
	Tag          string
	Inaccessible bool // True if the image cannot be accessed
}

// ImageAnalysis represents the analysis results for images
type ImageAnalysis struct {
	Images      []Image
	TotalSize   int64
	UniqueSize  int64 // Size after deduplication
	Performance *PerformanceMetrics
}

// GetUniqueImages returns a map of unique images by name
func (ia *ImageAnalysis) GetUniqueImages() map[string]Image {
	uniqueImages := make(map[string]Image)
	for _, img := range ia.Images {
		uniqueImages[img.Name] = img
	}
	return uniqueImages
}

// GetTopImagesBySize returns the top N images sorted by size
func (ia *ImageAnalysis) GetTopImagesBySize(n int) []Image {
	if n > len(ia.Images) {
		n = len(ia.Images)
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]Image, len(ia.Images))
	copy(sorted, ia.Images)

	// Sort by size (descending) using Go's sort package
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Size > sorted[j].Size
	})

	return sorted[:n]
}

// NewInaccessibleImage creates an image entry for an inaccessible image
func NewInaccessibleImage(imageName string) *Image {
	registry, tag := util.ExtractRegistryAndTag(imageName)
	return &Image{
		Name:         imageName,
		Size:         0,
		Registry:     registry,
		Tag:          tag,
		Inaccessible: true,
	}
}
