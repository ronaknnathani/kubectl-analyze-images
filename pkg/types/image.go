package types

import (
	"sort"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/util"
)

// ImageSortBy identifies the metric used to sort image table rows.
type ImageSortBy string

const (
	// ImageSortBySize sorts images by image size.
	ImageSortBySize ImageSortBy = "size"
	// ImageSortByPods sorts images by pod usage count.
	ImageSortByPods ImageSortBy = "pods"
	// ImageSortByCachedOnNodes sorts images by the number of nodes caching the image.
	ImageSortByCachedOnNodes ImageSortBy = "cached-on-nodes"
)

// Image represents a container image with its metadata
type Image struct {
	Name               string `json:"name"`
	Size               int64  `json:"sizeBytes"`
	Registry           string `json:"registry"`
	Tag                string `json:"tag"`
	PodCount           int    `json:"podCount"`
	ContainerCount     int    `json:"containerCount"`
	InitContainerCount int    `json:"initContainerCount"`
	NamespaceCount     int    `json:"namespaceCount"`
	CachedOnNodes      int    `json:"cachedOnNodes"`
	Inaccessible       bool   `json:"inaccessible"`
}

// ImageAnalysis represents the analysis results for images
type ImageAnalysis struct {
	Images       []Image             `json:"images"`
	TotalSize    int64               `json:"sumOfImageSizesBytes"`
	UniqueSize   int64               `json:"-"`
	PodsScanned  int                 `json:"podsScanned"`
	NodesScanned int                 `json:"nodesScanned"`
	ImagesInUse  int                 `json:"imagesInUse"`
	UnusedImages int                 `json:"unusedImages"`
	Performance  *PerformanceMetrics `json:"performance,omitempty"`
}

// GetUniqueImages returns a map of unique images by name
func (ia *ImageAnalysis) GetUniqueImages() map[string]Image {
	uniqueImages := make(map[string]Image)
	for _, img := range ia.Images {
		uniqueImages[img.Name] = img
	}
	return uniqueImages
}

// GetTopImagesBySize returns the top N images sorted by size.
func (ia *ImageAnalysis) GetTopImagesBySize(n int) []Image {
	return ia.GetTopImages(n, ImageSortBySize)
}

// GetTopImages returns the top N images sorted by the requested metric.
func (ia *ImageAnalysis) GetTopImages(n int, sortBy ImageSortBy) []Image {
	if n > len(ia.Images) {
		n = len(ia.Images)
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]Image, len(ia.Images))
	copy(sorted, ia.Images)

	sort.SliceStable(sorted, func(i, j int) bool {
		switch sortBy {
		case ImageSortByPods:
			if sorted[i].PodCount != sorted[j].PodCount {
				return sorted[i].PodCount > sorted[j].PodCount
			}
		case ImageSortByCachedOnNodes:
			if sorted[i].CachedOnNodes != sorted[j].CachedOnNodes {
				return sorted[i].CachedOnNodes > sorted[j].CachedOnNodes
			}
		default:
			if sorted[i].Size != sorted[j].Size {
				return sorted[i].Size > sorted[j].Size
			}
		}
		if sorted[i].Size != sorted[j].Size {
			return sorted[i].Size > sorted[j].Size
		}
		if sorted[i].PodCount != sorted[j].PodCount {
			return sorted[i].PodCount > sorted[j].PodCount
		}
		return sorted[i].Name < sorted[j].Name
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
