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

func TestJSONPrinter_Print(t *testing.T) {
	tests := []struct {
		name              string
		analysis          *types.ImageAnalysis
		wantImagesCount   int
		wantTotalImages   int
		wantTotalSize     int64
		wantUniqueSize    int64
		wantPerformance   bool
		wantImagesNotNull bool
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
			wantImagesCount:   2,
			wantTotalImages:   2,
			wantTotalSize:     243000000,
			wantUniqueSize:    200000000,
			wantPerformance:   true,
			wantImagesNotNull: true,
		},
		{
			name: "empty images produces valid JSON",
			analysis: &types.ImageAnalysis{
				Images:      []types.Image{},
				TotalSize:   0,
				UniqueSize:  0,
				Performance: nil,
			},
			wantImagesCount:   0,
			wantTotalImages:   0,
			wantTotalSize:     0,
			wantUniqueSize:    0,
			wantPerformance:   false,
			wantImagesNotNull: true,
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
			wantImagesCount:   5,
			wantTotalImages:   5,
			wantTotalSize:     350000000,
			wantUniqueSize:    350000000,
			wantPerformance:   true,
			wantImagesNotNull: true,
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
			wantImagesCount:   1,
			wantTotalImages:   1,
			wantTotalSize:     5000000,
			wantUniqueSize:    5000000,
			wantPerformance:   false,
			wantImagesNotNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printer := NewJSONPrinter()

			err := printer.Print(&buf, tt.analysis)
			require.NoError(t, err, "Print should not return an error")

			// Unmarshal the JSON to verify structure
			var result map[string]interface{}
			err = json.Unmarshal(buf.Bytes(), &result)
			require.NoError(t, err, "output should be valid JSON")

			// Check images array
			images, ok := result["images"]
			require.True(t, ok, "result should have 'images' key")
			if tt.wantImagesNotNull {
				require.NotNil(t, images, "images should not be nil")
				imagesArray, isArray := images.([]interface{})
				require.True(t, isArray, "images should be an array")
				assert.Equal(t, tt.wantImagesCount, len(imagesArray), "images array should have correct length")
			}

			// Check summary object
			summary, ok := result["summary"]
			require.True(t, ok, "result should have 'summary' key")
			require.NotNil(t, summary, "summary should not be nil")

			summaryMap, ok := summary.(map[string]interface{})
			require.True(t, ok, "summary should be an object")

			totalImages, ok := summaryMap["totalImages"]
			require.True(t, ok, "summary should have 'totalImages'")
			assert.Equal(t, float64(tt.wantTotalImages), totalImages, "totalImages should match")

			totalSize, ok := summaryMap["totalSize"]
			require.True(t, ok, "summary should have 'totalSize'")
			assert.Equal(t, float64(tt.wantTotalSize), totalSize, "totalSize should match")

			uniqueSize, ok := summaryMap["uniqueSize"]
			require.True(t, ok, "summary should have 'uniqueSize'")
			assert.Equal(t, float64(tt.wantUniqueSize), uniqueSize, "uniqueSize should match")

			// Check performance
			performance, hasPerformance := result["performance"]
			if tt.wantPerformance {
				assert.True(t, hasPerformance, "result should have 'performance' key")
				assert.NotNil(t, performance, "performance should not be nil")

				perfMap, ok := performance.(map[string]interface{})
				require.True(t, ok, "performance should be an object")

				_, hasImagesProcessed := perfMap["ImagesProcessed"]
				assert.True(t, hasImagesProcessed, "performance should have 'ImagesProcessed' field")
			} else if hasPerformance {
				// When performance is nil, it should be omitted or null
				assert.Nil(t, performance, "performance should be nil when not provided")
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

	// Verify the JSON is valid
	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	// Check that the inaccessible image is included
	images, ok := result["images"].([]interface{})
	require.True(t, ok)
	require.Len(t, images, 1)

	image, ok := images[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "private/image:latest", image["Name"])
	assert.Equal(t, true, image["Inaccessible"])
	assert.Equal(t, float64(0), image["Size"])
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
			CacheHits:          10,
			CacheMisses:        5,
		},
	}

	var buf bytes.Buffer
	printer := NewJSONPrinter()

	err := printer.Print(&buf, analysis)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	performance, ok := result["performance"].(map[string]interface{})
	require.True(t, ok, "performance should be present")

	// Verify all performance fields are present
	assert.Equal(t, float64(1), performance["ImagesProcessed"])
	assert.Equal(t, float64(0), performance["ImagesFailed"])
	assert.Equal(t, float64(0), performance["ImagesInaccessible"])
	assert.Equal(t, float64(10), performance["CacheHits"])
	assert.Equal(t, float64(5), performance["CacheMisses"])
}
