package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/util"
)

// Reporter handles output generation
type Reporter struct {
	outputFormat  string
	showHistogram bool
	noColor       bool
	topImages     int
}

// NewReporter creates a new reporter
func NewReporter(outputFormat string) *Reporter {
	return &Reporter{
		outputFormat:  outputFormat,
		showHistogram: true, // Make histogram default
		noColor:       false,
		topImages:     25, // Default to 25 top images
	}
}

// SetShowHistogram enables or disables histogram display
func (r *Reporter) SetShowHistogram(show bool) {
	r.showHistogram = show
}

// SetNoColor enables or disables colored output
func (r *Reporter) SetNoColor(noColor bool) {
	r.noColor = noColor
}

// SetTopImages sets the number of top images to display
func (r *Reporter) SetTopImages(count int) {
	r.topImages = count
}

// GenerateReport generates a report from image analysis
func (r *Reporter) GenerateReport(analysis *types.ImageAnalysis) error {
	switch r.outputFormat {
	case "table":
		return r.generateTableReport(analysis)
	case "json":
		return r.generateJSONReport(analysis)
	default:
		return fmt.Errorf("unsupported output format: %s", r.outputFormat)
	}
}

// generateTableReport generates a table-formatted report using tablewriter
func (r *Reporter) generateTableReport(analysis *types.ImageAnalysis) error {
	// Performance Summary
	if analysis.Performance != nil {
		fmt.Println("Performance Summary")
		fmt.Println("==================")

		performanceTable := tablewriter.NewWriter(os.Stdout)
		performanceTable.Header("Metric", "Value")
		if analysis.Performance.PodQueryTime > 0 {
			performanceTable.Append("Pod Query Time", analysis.Performance.PodQueryTime.String())
		}
		if analysis.Performance.NodeQueryTime > 0 {
			performanceTable.Append("Node Query Time", analysis.Performance.NodeQueryTime.String())
		}
		performanceTable.Append("Image Analysis Time", analysis.Performance.ImageAnalysisTime.String())
		performanceTable.Append("Total Time", analysis.Performance.TotalTime.String())
		performanceTable.Append("Images Processed", strconv.Itoa(analysis.Performance.ImagesProcessed))
		performanceTable.Render()
		fmt.Println()
	}

	// Image Analysis Summary
	fmt.Println("Image Analysis Summary")
	fmt.Println("=====================")

	summaryTable := tablewriter.NewWriter(os.Stdout)
	summaryTable.Header("Metric", "Value")
	summaryTable.Append("Total Images", strconv.Itoa(len(analysis.Images)))
	summaryTable.Append("Unique Images", strconv.Itoa(len(analysis.GetUniqueImages())))
	summaryTable.Append("Total Size", util.FormatBytes(analysis.TotalSize))
	summaryTable.Render()
	fmt.Println()

	// Image Size Distribution Histogram (if requested and we have images)
	if r.showHistogram && len(analysis.Images) > 0 {
		fmt.Println("Image Size Distribution")
		fmt.Println("=======================")

		config := types.DefaultHistogramConfig()
		config.Title = "Image Size Distribution"
		config.Height = 15
		config.Width = 60
		config.ShowColors = !r.noColor // Disable colors if noColor flag is set

		histogramData := analysis.GenerateImageSizeHistogram(config)
		fmt.Print(histogramData.RenderASCII(config, analysis))
	}

	// Top images by size
	if len(analysis.Images) > 0 {
		fmt.Println()
		fmt.Printf("Top %d Images by Size\n", r.topImages)
		fmt.Println("=====================")

		imageTable := tablewriter.NewWriter(os.Stdout)
		imageTable.Header("Image", "Size")

		topImages := analysis.GetTopImagesBySize(r.topImages)
		for _, img := range topImages {
			size := util.FormatBytes(img.Size)
			if img.Inaccessible {
				size = "INACCESSIBLE"
			}
			imageTable.Append(img.Name, size)
		}
		imageTable.Render()
		fmt.Println()
	}

	return nil
}

// generateJSONReport generates a JSON-formatted report using proper JSON marshaling
func (r *Reporter) generateJSONReport(analysis *types.ImageAnalysis) error {
	// Create a structured report for JSON marshaling
	report := struct {
		Performance *types.PerformanceMetrics `json:"performance,omitempty"`
		Summary     struct {
			TotalImages int   `json:"totalImages"`
			TotalSize   int64 `json:"totalSize"`
			UniqueSize  int64 `json:"uniqueSize"`
		} `json:"summary"`
		Images []types.Image `json:"images"`
	}{
		Performance: analysis.Performance,
		Images:      analysis.Images,
	}

	report.Summary.TotalImages = len(analysis.Images)
	report.Summary.TotalSize = analysis.TotalSize
	report.Summary.UniqueSize = analysis.UniqueSize

	// Marshal to JSON with proper formatting
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}
