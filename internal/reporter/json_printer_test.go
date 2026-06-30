package reporter

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
)

type jsonReport struct {
	Performance *types.PerformanceMetrics `json:"performance,omitempty"`
	Summary     jsonSummary               `json:"summary"`
	Images      []types.Image             `json:"images"`
}

type jsonSummary struct {
	PodsScanned     int   `json:"podsScanned"`
	NodesScanned    int   `json:"nodesScanned"`
	ImagesAnalyzed  int   `json:"imagesAnalyzed"`
	ImagesInUse     int   `json:"imagesInUse"`
	UnusedImages    int   `json:"unusedImages"`
	SumOfImageSizes int64 `json:"sumOfImageSizesBytes"`
}

func TestJSONPrinter_Print(t *testing.T) {
	tests := []struct {
		name                string
		analysis            *types.ImageAnalysis
		wantImagesCount     int
		wantImagesAnalyzed  int
		wantSumOfImageSizes int64
		wantPerformance     bool
		wantImagesNotNull   bool
	}{
		{
			name: "valid JSON structure",
			analysis: &types.ImageAnalysis{
				Images: []types.Image{
					{Name: "nginx:1.21", Size: 133000000, Registry: "docker.io", Tag: "1.21"},
					{Name: "redis:6.2", Size: 110000000, Registry: "docker.io", Tag: "6.2"},
				},
				TotalSize:  243000000,
				UniqueSize: 200000000,
				Performance: &types.PerformanceMetrics{
					ImagesProcessed: 2,
					TotalTime:       1500 * time.Millisecond,
				},
			},
			wantImagesCount:     2,
			wantImagesAnalyzed:  2,
			wantSumOfImageSizes: 243000000,
			wantPerformance:     true,
			wantImagesNotNull:   true,
		},
		{
			name: "empty images produces valid JSON",
			analysis: &types.ImageAnalysis{
				Images:      []types.Image{},
				TotalSize:   0,
				UniqueSize:  0,
				Performance: nil,
			},
			wantImagesCount:     0,
			wantImagesAnalyzed:  0,
			wantSumOfImageSizes: 0,
			wantPerformance:     false,
			wantImagesNotNull:   true,
		},
		{
			name: "performance metrics included",
			analysis: &types.ImageAnalysis{
				Images: []types.Image{
					{Name: "image1:latest", Size: 50000000, Registry: "docker.io", Tag: "latest"},
					{Name: "image2:latest", Size: 60000000, Registry: "docker.io", Tag: "latest"},
					{Name: "image3:latest", Size: 70000000, Registry: "docker.io", Tag: "latest"},
					{Name: "image4:latest", Size: 80000000, Registry: "docker.io", Tag: "latest"},
					{Name: "image5:latest", Size: 90000000, Registry: "docker.io", Tag: "latest"},
				},
				TotalSize:  350000000,
				UniqueSize: 350000000,
				Performance: &types.PerformanceMetrics{
					ImagesProcessed: 5,
					TotalTime:       2000 * time.Millisecond,
				},
			},
			wantImagesCount:     5,
			wantImagesAnalyzed:  5,
			wantSumOfImageSizes: 350000000,
			wantPerformance:     true,
			wantImagesNotNull:   true,
		},
		{
			name: "no performance when nil",
			analysis: &types.ImageAnalysis{
				Images: []types.Image{
					{Name: "alpine:3.14", Size: 5000000, Registry: "docker.io", Tag: "3.14"},
				},
				TotalSize:   5000000,
				UniqueSize:  5000000,
				Performance: nil,
			},
			wantImagesCount:     1,
			wantImagesAnalyzed:  1,
			wantSumOfImageSizes: 5000000,
			wantPerformance:     false,
			wantImagesNotNull:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printer := NewJSONPrinter()

			err := printer.Print(&buf, tt.analysis)
			require.NoError(t, err, "Print should not return an error")

			var result jsonReport
			err = json.Unmarshal(buf.Bytes(), &result)
			require.NoError(t, err, "output should be valid JSON")

			if tt.wantImagesNotNull {
				require.NotNil(t, result.Images, "images should not be nil")
				assert.Equal(t, tt.wantImagesCount, len(result.Images), "images array should have correct length")
			}

			assert.Equal(t, tt.wantImagesAnalyzed, result.Summary.ImagesAnalyzed, "imagesAnalyzed should match")
			assert.Equal(t, tt.wantSumOfImageSizes, result.Summary.SumOfImageSizes, "sumOfImageSizesBytes should match")

			if tt.wantPerformance {
				require.NotNil(t, result.Performance, "performance should not be nil")
				assert.Equal(t, tt.analysis.Performance.ImagesProcessed, result.Performance.ImagesProcessed)
			} else {
				assert.Nil(t, result.Performance, "performance should be nil when not provided")
			}
		})
	}
}

func TestJSONPrinter_Print_InaccessibleImage(t *testing.T) {
	analysis := &types.ImageAnalysis{
		Images: []types.Image{
			{Name: "private/image:latest", Size: 0, Registry: "private", Tag: "latest", Inaccessible: true},
		},
		TotalSize:   0,
		UniqueSize:  0,
		Performance: nil,
	}

	var buf bytes.Buffer
	printer := NewJSONPrinter()

	err := printer.Print(&buf, analysis)
	require.NoError(t, err)

	var result jsonReport
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	require.Len(t, result.Images, 1)
	assert.Equal(t, "private/image:latest", result.Images[0].Name)
	assert.True(t, result.Images[0].Inaccessible)
	assert.Equal(t, int64(0), result.Images[0].Size)
}

func TestJSONPrinter_Print_CompletePerformanceMetrics(t *testing.T) {
	analysis := &types.ImageAnalysis{
		Images: []types.Image{
			{Name: "test:latest", Size: 100000000, Registry: "docker.io", Tag: "latest"},
		},
		TotalSize:  100000000,
		UniqueSize: 100000000,
		Performance: &types.PerformanceMetrics{
			PodQueryTime:       100 * time.Millisecond,
			NodeQueryTime:      50 * time.Millisecond,
			ImageAnalysisTime:  200 * time.Millisecond,
			TotalTime:          350 * time.Millisecond,
			ImagesProcessed:    1,
			ImagesFailed:       0,
			ImagesInaccessible: 0,
		},
	}

	var buf bytes.Buffer
	printer := NewJSONPrinter()

	err := printer.Print(&buf, analysis)
	require.NoError(t, err)

	var result jsonReport
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	require.NotNil(t, result.Performance, "performance should be present")
	assert.Equal(t, 1, result.Performance.ImagesProcessed)
	assert.Equal(t, 0, result.Performance.ImagesFailed)
	assert.Equal(t, 0, result.Performance.ImagesInaccessible)
}
