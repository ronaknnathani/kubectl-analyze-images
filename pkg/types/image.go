package types

// Image represents a container image with its metadata
type Image struct {
	Name     string
	Size     int64
	Layers   []Layer
	Registry string
	Tag      string
}

// Layer represents a container image layer
type Layer struct {
	Digest string
	Size   int64
}

// ImageAnalysis represents the analysis results for images
type ImageAnalysis struct {
	Images     []Image
	TotalSize  int64
	UniqueSize int64 // Size after deduplication
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

	// Sort by size (descending)
	sorted := make([]Image, len(ia.Images))
	copy(sorted, ia.Images)

	// Simple bubble sort for now (will be optimized later)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Size < sorted[j+1].Size {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted[:n]
}
