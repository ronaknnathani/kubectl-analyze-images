package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUniqueImages(t *testing.T) {
	tests := []struct {
		name     string
		analysis *ImageAnalysis
		want     int
	}{
		{
			name: "multiple duplicate images",
			analysis: &ImageAnalysis{
				Images: []Image{
					{Name: "nginx:1.21", Size: 133000000},
					{Name: "nginx:1.21", Size: 133000000},
					{Name: "redis:6.2", Size: 110000000},
					{Name: "redis:6.2", Size: 110000000},
				},
			},
			want: 2,
		},
		{
			name: "all unique images",
			analysis: &ImageAnalysis{
				Images: []Image{
					{Name: "nginx:1.21", Size: 133000000},
					{Name: "redis:6.2", Size: 110000000},
					{Name: "postgres:13", Size: 314000000},
				},
			},
			want: 3,
		},
		{
			name: "empty images",
			analysis: &ImageAnalysis{
				Images: []Image{},
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unique := tt.analysis.GetUniqueImages()
			assert.Equal(t, tt.want, len(unique))
		})
	}
}

func TestGetTopImagesBySize(t *testing.T) {
	tests := []struct {
		name     string
		analysis *ImageAnalysis
		n        int
		want     []string // Expected image names in order
	}{
		{
			name: "top 2 from 5 images",
			analysis: &ImageAnalysis{
				Images: []Image{
					{Name: "small", Size: 10000000},
					{Name: "large", Size: 500000000},
					{Name: "medium", Size: 100000000},
					{Name: "largest", Size: 600000000},
					{Name: "tiny", Size: 5000000},
				},
			},
			n:    2,
			want: []string{"largest", "large"},
		},
		{
			name: "top 3 from 3 images",
			analysis: &ImageAnalysis{
				Images: []Image{
					{Name: "first", Size: 300000000},
					{Name: "second", Size: 200000000},
					{Name: "third", Size: 100000000},
				},
			},
			n:    3,
			want: []string{"first", "second", "third"},
		},
		{
			name: "request more than available",
			analysis: &ImageAnalysis{
				Images: []Image{
					{Name: "only", Size: 100000000},
				},
			},
			n:    10,
			want: []string{"only"},
		},
		{
			name: "empty images",
			analysis: &ImageAnalysis{
				Images: []Image{},
			},
			n:    5,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.analysis.GetTopImagesBySize(tt.n)
			assert.Equal(t, len(tt.want), len(result))
			for i, name := range tt.want {
				assert.Equal(t, name, result[i].Name)
			}
		})
	}
}

func TestGetTopImages_WithSortOptions(t *testing.T) {
	analysis := &ImageAnalysis{
		Images: []Image{
			{Name: "largest-unused", Size: 600, PodCount: 0, CachedOnNodes: 1},
			{Name: "most-pods", Size: 100, PodCount: 5, CachedOnNodes: 2},
			{Name: "most-cached", Size: 200, PodCount: 2, CachedOnNodes: 8},
		},
	}

	tests := []struct {
		name   string
		sortBy ImageSortBy
		want   []string
	}{
		{
			name:   "sort by size",
			sortBy: ImageSortBySize,
			want:   []string{"largest-unused", "most-cached", "most-pods"},
		},
		{
			name:   "sort by pod usage",
			sortBy: ImageSortByPods,
			want:   []string{"most-pods", "most-cached", "largest-unused"},
		},
		{
			name:   "sort by cached on nodes",
			sortBy: ImageSortByCachedOnNodes,
			want:   []string{"most-cached", "most-pods", "largest-unused"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analysis.GetTopImages(3, tt.sortBy)
			for i, name := range tt.want {
				assert.Equal(t, name, result[i].Name)
			}
		})
	}
}

func TestNewInaccessibleImage(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		wantName  string
		wantReg   string
		wantTag   string
	}{
		{
			name:      "private registry image",
			imageName: "private.registry.com/app:v1.0",
			wantName:  "private.registry.com/app:v1.0",
			wantReg:   "private.registry.com",
			wantTag:   "v1.0",
		},
		{
			name:      "docker hub image",
			imageName: "nginx:latest",
			wantName:  "nginx:latest",
			wantReg:   "docker.io",
			wantTag:   "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := NewInaccessibleImage(tt.imageName)
			assert.Equal(t, tt.wantName, img.Name)
			assert.Equal(t, tt.wantReg, img.Registry)
			assert.Equal(t, tt.wantTag, img.Tag)
			assert.True(t, img.Inaccessible)
			assert.Equal(t, int64(0), img.Size)
		})
	}
}
