package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
)

func TestTablePrinter_Print(t *testing.T) {
	tests := []struct {
		name           string
		analysis       *types.ImageAnalysis
		showHistogram  bool
		noColor        bool
		topImages      int
		wantContains   []string
		wantNotContain []string
	}{
		{
			name: "basic output with images",
			analysis: &types.ImageAnalysis{
				Images: []types.Image{
					{Name: "nginx:1.21", Size: 133000000, Registry: "docker.io", Tag: "1.21"},
					{Name: "redis:6.2", Size: 110000000, Registry: "docker.io", Tag: "6.2"},
				},
				TotalSize:  243000000,
				UniqueSize: 243000000,
				Performance: &types.PerformanceMetrics{
					ImagesProcessed: 2,
					TotalTime:       1500 * time.Millisecond,
				},
			},
			showHistogram: false,
			noColor:       true,
			topImages:     25,
			wantContains: []string{
				"Performance Summary",
				"Image Analysis Summary",
				"nginx:1.21",
				"redis:6.2",
				"Total Images",
			},
			wantNotContain: []string{},
		},
		{
			name: "empty analysis",
			analysis: &types.ImageAnalysis{
				Images:      []types.Image{},
				TotalSize:   0,
				UniqueSize:  0,
				Performance: nil,
			},
			showHistogram: false,
			noColor:       true,
			topImages:     25,
			wantContains: []string{
				"Image Analysis Summary",
				"Total Images",
			},
			wantNotContain: []string{
				"Performance Summary",
			},
		},
		{
			name: "inaccessible images",
			analysis: &types.ImageAnalysis{
				Images: []types.Image{
					{Name: "private/image:latest", Size: 0, Registry: "private", Tag: "latest", Inaccessible: true},
				},
				TotalSize:  0,
				UniqueSize: 0,
				Performance: &types.PerformanceMetrics{
					ImagesProcessed: 1,
					TotalTime:       500 * time.Millisecond,
				},
			},
			showHistogram: false,
			noColor:       true,
			topImages:     25,
			wantContains: []string{
				"INACCESSIBLE",
				"private/image:latest",
			},
			wantNotContain: []string{},
		},
		{
			name: "histogram disabled",
			analysis: &types.ImageAnalysis{
				Images: []types.Image{
					{Name: "alpine:3.14", Size: 5000000, Registry: "docker.io", Tag: "3.14"},
					{Name: "nginx:1.21", Size: 133000000, Registry: "docker.io", Tag: "1.21"},
					{Name: "redis:6.2", Size: 110000000, Registry: "docker.io", Tag: "6.2"},
				},
				TotalSize:  248000000,
				UniqueSize: 248000000,
				Performance: &types.PerformanceMetrics{
					ImagesProcessed: 3,
					TotalTime:       2000 * time.Millisecond,
				},
			},
			showHistogram: false,
			noColor:       true,
			topImages:     25,
			wantContains: []string{
				"Image Analysis Summary",
			},
			wantNotContain: []string{
				"Image Size Distribution",
			},
		},
		{
			name: "histogram enabled with images",
			analysis: &types.ImageAnalysis{
				Images: []types.Image{
					{Name: "alpine:3.14", Size: 5000000, Registry: "docker.io", Tag: "3.14"},
					{Name: "nginx:1.21", Size: 133000000, Registry: "docker.io", Tag: "1.21"},
					{Name: "redis:6.2", Size: 110000000, Registry: "docker.io", Tag: "6.2"},
					{Name: "postgres:13", Size: 314000000, Registry: "docker.io", Tag: "13"},
				},
				TotalSize:  562000000,
				UniqueSize: 562000000,
				Performance: &types.PerformanceMetrics{
					ImagesProcessed: 4,
					TotalTime:       3000 * time.Millisecond,
				},
			},
			showHistogram: true,
			noColor:       true,
			topImages:     25,
			wantContains: []string{
				"Image Size Distribution",
			},
			wantNotContain: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			printer := NewTablePrinter(tt.showHistogram, tt.noColor, tt.topImages)

			err := printer.Print(&buf, tt.analysis)
			require.NoError(t, err, "Print should not return an error")

			output := buf.String()

			// Check for expected strings
			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "output should contain %q", want)
			}

			// Check for strings that should NOT be present
			for _, notWant := range tt.wantNotContain {
				assert.NotContains(t, output, notWant, "output should not contain %q", notWant)
			}
		})
	}
}

func TestTablePrinter_Print_PerformanceMetrics(t *testing.T) {
	analysis := &types.ImageAnalysis{
		Images: []types.Image{
			{Name: "test:latest", Size: 100000000, Registry: "docker.io", Tag: "latest"},
		},
		TotalSize:  100000000,
		UniqueSize: 100000000,
		Performance: &types.PerformanceMetrics{
			PodQueryTime:      100 * time.Millisecond,
			NodeQueryTime:     50 * time.Millisecond,
			ImageAnalysisTime: 200 * time.Millisecond,
			TotalTime:         350 * time.Millisecond,
			ImagesProcessed:   1,
		},
	}

	var buf bytes.Buffer
	printer := NewTablePrinter(false, true, 25)

	err := printer.Print(&buf, analysis)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Performance Summary")
	assert.Contains(t, output, "Pod Query Time")
	assert.Contains(t, output, "Node Query Time")
	assert.Contains(t, output, "Image Analysis Time")
	assert.Contains(t, output, "Total Time")
	assert.Contains(t, output, "Images Processed")
}

func TestTablePrinter_Print_TopImagesLimit(t *testing.T) {
	// Create 10 images
	images := make([]types.Image, 10)
	for i := 0; i < 10; i++ {
		images[i] = types.Image{
			Name:     "image" + string(rune(i+'0')) + ":latest",
			Size:     int64((10 - i) * 10000000), // Descending sizes
			Registry: "docker.io",
			Tag:      "latest",
		}
	}

	analysis := &types.ImageAnalysis{
		Images:      images,
		TotalSize:   550000000,
		UniqueSize:  550000000,
		Performance: nil,
	}

	// Test with topImages = 3
	var buf bytes.Buffer
	printer := NewTablePrinter(false, true, 3)

	err := printer.Print(&buf, analysis)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Top 3 Images by Size")

	// Count the number of table rows in the Top Images section
	// This is a simple check - the top 3 largest images should be present
	lines := strings.Split(output, "\n")
	topSectionStarted := false
	imageCount := 0
	for _, line := range lines {
		if strings.Contains(line, "Top 3 Images by Size") {
			topSectionStarted = true
			continue
		}
		if topSectionStarted && strings.Contains(line, "image") {
			imageCount++
		}
	}
	// We expect to see exactly 3 images in the top images section
	assert.Equal(t, 3, imageCount, "should display exactly 3 images")
}
