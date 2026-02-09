package types

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/util"
)

// HistogramConfig holds configuration for histogram generation
type HistogramConfig struct {
	Bins       int    // Number of bins in the histogram
	Width      int    // Width of the histogram in characters
	Height     int    // Maximum height of bars in characters
	Title      string // Title for the histogram
	ShowStats  bool   // Whether to show statistics
	ShowColors bool   // Whether to show colors in output
}

// DefaultHistogramConfig returns default histogram configuration
func DefaultHistogramConfig() *HistogramConfig {
	return &HistogramConfig{
		Bins:       10,
		Width:      60,
		Height:     20,
		Title:      "Image Size Distribution",
		ShowStats:  true,
		ShowColors: true,
	}
}

// HistogramBin represents a single bin in the histogram
type HistogramBin struct {
	Min   float64  // Minimum value for this bin
	Max   float64  // Maximum value for this bin
	Count int      // Number of items in this bin
	Items []string // Names of items in this bin (for reference)
}

// HistogramData contains the histogram data and statistics
type HistogramData struct {
	Bins     []HistogramBin
	MinValue float64
	MaxValue float64
	Mean     float64
	StdDev   float64
	Total    int
}

// GenerateImageSizeHistogram creates a histogram from image analysis
func (ia *ImageAnalysis) GenerateImageSizeHistogram(config *HistogramConfig) *HistogramData {
	if len(ia.Images) == 0 {
		return &HistogramData{}
	}

	// Extract sizes and calculate statistics (single pass)
	sizes := make([]float64, len(ia.Images))
	var sum float64
	minVal := math.Inf(1)
	maxVal := math.Inf(-1)

	for i, img := range ia.Images {
		size := float64(img.Size)
		sizes[i] = size
		sum += size
		if size < minVal {
			minVal = size
		}
		if size > maxVal {
			maxVal = size
		}
	}

	mean := sum / float64(len(sizes))

	// Calculate standard deviation
	var variance float64
	for _, size := range sizes {
		variance += math.Pow(size-mean, 2)
	}
	stdDev := math.Sqrt(variance / float64(len(sizes)))

	// Create bins
	bins := make([]HistogramBin, config.Bins)
	binWidth := (maxVal - minVal) / float64(config.Bins)

	// Initialize bins
	for i := 0; i < config.Bins; i++ {
		bins[i] = HistogramBin{
			Min:   minVal + float64(i)*binWidth,
			Max:   minVal + float64(i+1)*binWidth,
			Count: 0,
			Items: make([]string, 0),
		}
	}

	// Assign images to bins
	for i, img := range ia.Images {
		size := sizes[i]
		binIndex := int((size - minVal) / binWidth)

		// Handle edge cases
		if binIndex < 0 {
			binIndex = 0
		}
		if binIndex >= config.Bins {
			binIndex = config.Bins - 1
		}

		bins[binIndex].Count++
		bins[binIndex].Items = append(bins[binIndex].Items, img.Name)
	}

	return &HistogramData{
		Bins:     bins,
		MinValue: minVal,
		MaxValue: maxVal,
		Mean:     mean,
		StdDev:   stdDev,
		Total:    len(ia.Images),
	}
}

// RenderASCII renders the histogram as ASCII art
func (hd *HistogramData) RenderASCII(config *HistogramConfig, analysis *ImageAnalysis) string {
	if len(hd.Bins) == 0 {
		return "No data to display\n"
	}

	var result strings.Builder

	// Find maximum count for scaling
	maxCount := 0
	for _, bin := range hd.Bins {
		if bin.Count > maxCount {
			maxCount = bin.Count
		}
	}

	if maxCount == 0 {
		return "No data in histogram\n"
	}

	// Render horizontal histogram bars (kubectl-node_resource style)
	// Note: Title is handled by the reporter, so we don't print it here

	// Define bar width for horizontal bars
	barMaxWidth := 40 // Maximum width for bars

	// Define color functions based on percentage
	greenBar := color.New(color.FgGreen).SprintFunc()
	yellowBar := color.New(color.FgYellow).SprintFunc()
	redBar := color.New(color.FgRed).SprintFunc()
	cyanLabel := color.New(color.FgCyan).SprintFunc()

	for i, bin := range hd.Bins {
		// Skip empty bins
		if bin.Count == 0 {
			continue
		}

		// Calculate bar width
		barWidth := int(float64(bin.Count) * float64(barMaxWidth) / float64(maxCount))
		if barWidth == 0 && bin.Count > 0 {
			barWidth = 1 // Ensure at least one character for non-zero counts
		}

		// Format the bin range label with color
		rangeLabel := fmt.Sprintf("%6s-%6s", util.FormatBytesShort(int64(bin.Min)), util.FormatBytesShort(int64(bin.Max)))
		if config.ShowColors {
			rangeLabel = cyanLabel(rangeLabel)
		}

		// Create the bar with color based on bin position (size range)
		percentage := float64(bin.Count) / float64(hd.Total) * 100
		var bar string
		barChars := strings.Repeat("█", barWidth)

		// Color coding based on size ranges (bin position)
		// Lower bins (smaller sizes) = green, middle = yellow, higher bins (larger sizes) = red
		if config.ShowColors {
			binPosition := float64(i) / float64(len(hd.Bins)-1) * 100 // 0-100% based on bin position
			if binPosition < 33 {
				bar = greenBar(barChars)
			} else if binPosition < 67 {
				bar = yellowBar(barChars)
			} else {
				bar = redBar(barChars)
			}
		} else {
			bar = barChars
		}

		// Format the count and percentage
		countLabel := fmt.Sprintf("(%d images, %.0f%%)", bin.Count, percentage)

		// Print the row
		result.WriteString(fmt.Sprintf("  %s : %-*s %s\n", rangeLabel, barMaxWidth, bar, countLabel))
	}

	// Statistics in a compact format
	if config.ShowStats {
		result.WriteString("\n")
		result.WriteString("Image Size Summary\n")
		result.WriteString("==================\n")

		result.WriteString(fmt.Sprintf("Total Images: %d\n", hd.Total))
		result.WriteString(fmt.Sprintf("Size Range: %s - %s\n", util.FormatBytes(int64(hd.MinValue)), util.FormatBytes(int64(hd.MaxValue))))
		result.WriteString(fmt.Sprintf("Mean Size: %s\n", util.FormatBytes(int64(hd.Mean))))

		// Add percentile information similar to kubectl-node_resource
		result.WriteString("\nSize Percentiles\n")
		result.WriteString("================\n")

		// Use actual image sizes for percentile calculation (more accurate)
		actualSizes := make([]int64, 0, hd.Total)
		for _, bin := range hd.Bins {
			for _, itemName := range bin.Items {
				// Find the actual image size for this item
				for _, img := range analysis.Images {
					if img.Name == itemName {
						actualSizes = append(actualSizes, img.Size)
						break
					}
				}
			}
		}

		if len(actualSizes) > 0 {
			// Sort sizes for percentile calculation
			sort.Slice(actualSizes, func(i, j int) bool {
				return actualSizes[i] < actualSizes[j]
			})

			// Calculate percentiles
			p100 := actualSizes[len(actualSizes)-1]
			p90 := actualSizes[int(float64(len(actualSizes))*0.9)]
			p50 := actualSizes[len(actualSizes)/2]
			p10 := actualSizes[int(float64(len(actualSizes))*0.1)]
			p0 := actualSizes[0]

			result.WriteString(fmt.Sprintf("  - P0 (Min)      : %s\n", util.FormatBytes(p0)))
			result.WriteString(fmt.Sprintf("  - P10           : %s\n", util.FormatBytes(p10)))
			result.WriteString(fmt.Sprintf("  - P50 (Median)  : %s\n", util.FormatBytes(p50)))
			result.WriteString(fmt.Sprintf("  - P90           : %s\n", util.FormatBytes(p90)))
			result.WriteString(fmt.Sprintf("  - P100 (Max)    : %s\n", util.FormatBytes(p100)))
		}
	}

	return result.String()
}
