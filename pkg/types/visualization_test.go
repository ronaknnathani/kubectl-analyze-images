package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultHistogramConfig(t *testing.T) {
	config := DefaultHistogramConfig()
	assert.Equal(t, 10, config.Bins)
	assert.Equal(t, 60, config.Width)
	assert.Equal(t, 20, config.Height)
	assert.Equal(t, "Image Size Distribution", config.Title)
	assert.True(t, config.ShowStats)
	assert.True(t, config.ShowColors)
}

func TestGenerateImageSizeHistogram(t *testing.T) {
	tests := []struct {
		name     string
		analysis *ImageAnalysis
		config   *HistogramConfig
		wantBins int
	}{
		{
			name: "basic histogram with 4 images",
			analysis: &ImageAnalysis{
				Images: []Image{
					{Name: "small", Size: 10000000},
					{Name: "medium", Size: 100000000},
					{Name: "large", Size: 200000000},
					{Name: "huge", Size: 500000000},
				},
			},
			config: &HistogramConfig{
				Bins:       5,
				Width:      60,
				Height:     20,
				Title:      "Test",
				ShowStats:  true,
				ShowColors: false,
			},
			wantBins: 5,
		},
		{
			name: "empty images",
			analysis: &ImageAnalysis{
				Images: []Image{},
			},
			config: &HistogramConfig{
				Bins:       10,
				Width:      60,
				Height:     20,
				Title:      "Test",
				ShowStats:  true,
				ShowColors: false,
			},
			wantBins: 0,
		},
		{
			name: "single image",
			analysis: &ImageAnalysis{
				Images: []Image{
					{Name: "only", Size: 100000000},
				},
			},
			config: &HistogramConfig{
				Bins:       10,
				Width:      60,
				Height:     20,
				Title:      "Test",
				ShowStats:  true,
				ShowColors: false,
			},
			wantBins: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.analysis.GenerateImageSizeHistogram(tt.config)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantBins, len(result.Bins))

			if len(tt.analysis.Images) > 0 {
				assert.Equal(t, len(tt.analysis.Images), result.Total)
				assert.Greater(t, result.MaxValue, 0.0)
			}
		})
	}
}

func TestGenerateImageSizeHistogram_Statistics(t *testing.T) {
	analysis := &ImageAnalysis{
		Images: []Image{
			{Name: "img1", Size: 100000000},
			{Name: "img2", Size: 200000000},
			{Name: "img3", Size: 300000000},
		},
	}

	config := &HistogramConfig{
		Bins:       10,
		Width:      60,
		Height:     20,
		Title:      "Test",
		ShowStats:  true,
		ShowColors: false,
	}

	result := analysis.GenerateImageSizeHistogram(config)

	assert.Equal(t, 3, result.Total)
	assert.Equal(t, 100000000.0, result.MinValue)
	assert.Equal(t, 300000000.0, result.MaxValue)
	assert.Equal(t, 200000000.0, result.Mean)
	assert.Greater(t, result.StdDev, 0.0)
}

func TestRenderASCII(t *testing.T) {
	tests := []struct {
		name         string
		data         *HistogramData
		config       *HistogramConfig
		analysis     *ImageAnalysis
		wantContains []string
	}{
		{
			name: "empty histogram",
			data: &HistogramData{
				Bins: []HistogramBin{},
			},
			config: &HistogramConfig{
				ShowStats:  false,
				ShowColors: false,
			},
			analysis:     &ImageAnalysis{},
			wantContains: []string{"No data to display"},
		},
		{
			name: "histogram with data",
			data: &HistogramData{
				Bins: []HistogramBin{
					{Min: 0, Max: 100000000, Count: 2, Items: []string{"img1", "img2"}},
					{Min: 100000000, Max: 200000000, Count: 1, Items: []string{"img3"}},
				},
				MinValue: 0,
				MaxValue: 200000000,
				Mean:     100000000,
				StdDev:   50000000,
				Total:    3,
			},
			config: &HistogramConfig{
				ShowStats:  true,
				ShowColors: false,
			},
			analysis: &ImageAnalysis{
				Images: []Image{
					{Name: "img1", Size: 50000000},
					{Name: "img2", Size: 75000000},
					{Name: "img3", Size: 150000000},
				},
			},
			wantContains: []string{"Image Size Summary", "Total Images", "Size Range"},
		},
		{
			name: "histogram without stats",
			data: &HistogramData{
				Bins: []HistogramBin{
					{Min: 0, Max: 100000000, Count: 1, Items: []string{"img1"}},
				},
				MinValue: 0,
				MaxValue: 100000000,
				Mean:     50000000,
				StdDev:   0,
				Total:    1,
			},
			config: &HistogramConfig{
				ShowStats:  false,
				ShowColors: false,
			},
			analysis: &ImageAnalysis{
				Images: []Image{
					{Name: "img1", Size: 50000000},
				},
			},
			wantContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.RenderASCII(tt.config, tt.analysis)
			assert.NotEmpty(t, result)

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}
		})
	}
}

func TestRenderASCII_NoDataInBins(t *testing.T) {
	data := &HistogramData{
		Bins: []HistogramBin{
			{Min: 0, Max: 100000000, Count: 0, Items: []string{}},
			{Min: 100000000, Max: 200000000, Count: 0, Items: []string{}},
		},
		MinValue: 0,
		MaxValue: 200000000,
		Mean:     100000000,
		StdDev:   0,
		Total:    0,
	}

	config := &HistogramConfig{
		ShowStats:  false,
		ShowColors: false,
	}

	result := data.RenderASCII(config, &ImageAnalysis{})

	// When all bins have count 0, it should show "No data in histogram"
	assert.Contains(t, result, "No data in histogram")
}

func TestRenderASCII_SkipsEmptyBins(t *testing.T) {
	data := &HistogramData{
		Bins: []HistogramBin{
			{Min: 0, Max: 100000000, Count: 2, Items: []string{"img1", "img2"}},
			{Min: 100000000, Max: 200000000, Count: 0, Items: []string{}}, // Empty bin
			{Min: 200000000, Max: 300000000, Count: 1, Items: []string{"img3"}},
		},
		MinValue: 0,
		MaxValue: 300000000,
		Mean:     100000000,
		StdDev:   50000000,
		Total:    3,
	}

	config := &HistogramConfig{
		ShowStats:  false,
		ShowColors: false,
	}

	analysis := &ImageAnalysis{
		Images: []Image{
			{Name: "img1", Size: 50000000},
			{Name: "img2", Size: 75000000},
			{Name: "img3", Size: 250000000},
		},
	}

	result := data.RenderASCII(config, analysis)

	// Count the number of histogram bars (lines with █)
	lines := strings.Split(result, "\n")
	barCount := 0
	for _, line := range lines {
		if strings.Contains(line, "█") {
			barCount++
		}
	}

	// Should only have 2 bars (skipping the empty bin)
	assert.Equal(t, 2, barCount)
}
