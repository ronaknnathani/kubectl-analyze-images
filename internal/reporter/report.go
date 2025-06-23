package reporter

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rnathani/kubectl-analyze-images/pkg/types"
)

// Reporter handles output generation
type Reporter struct {
	outputFormat string
}

// NewReporter creates a new reporter
func NewReporter(outputFormat string) *Reporter {
	return &Reporter{
		outputFormat: outputFormat,
	}
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

// generateTableReport generates a table-formatted report
func (r *Reporter) generateTableReport(analysis *types.ImageAnalysis) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Summary
	fmt.Fprintf(w, "Image Analysis Summary\n")
	fmt.Fprintf(w, "=====================\n")
	fmt.Fprintf(w, "Total Images:\t%d\n", len(analysis.Images))
	fmt.Fprintf(w, "Total Size:\t%s\n", formatBytes(analysis.TotalSize))
	fmt.Fprintf(w, "Unique Size:\t%s\n", formatBytes(analysis.UniqueSize))
	fmt.Fprintf(w, "\n")

	// Top 25 images by size
	fmt.Fprintf(w, "Top 25 Images by Size\n")
	fmt.Fprintf(w, "=====================\n")
	fmt.Fprintf(w, "Image\tSize\tRegistry\tTag\n")
	fmt.Fprintf(w, "-----\t----\t--------\t---\n")

	topImages := analysis.GetTopImagesBySize(25)
	for _, img := range topImages {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			img.Name,
			formatBytes(img.Size),
			img.Registry,
			img.Tag)
	}

	return nil
}

// generateJSONReport generates a JSON-formatted report
func (r *Reporter) generateJSONReport(analysis *types.ImageAnalysis) error {
	// For Phase 1, we'll implement a simple JSON output
	// In later phases, we'll use proper JSON marshaling
	fmt.Printf("{\n")
	fmt.Printf("  \"summary\": {\n")
	fmt.Printf("    \"totalImages\": %d,\n", len(analysis.Images))
	fmt.Printf("    \"totalSize\": %d,\n", analysis.TotalSize)
	fmt.Printf("    \"uniqueSize\": %d\n", analysis.UniqueSize)
	fmt.Printf("  },\n")
	fmt.Printf("  \"images\": [\n")

	for i, img := range analysis.Images {
		fmt.Printf("    {\n")
		fmt.Printf("      \"name\": \"%s\",\n", img.Name)
		fmt.Printf("      \"size\": %d,\n", img.Size)
		fmt.Printf("      \"registry\": \"%s\",\n", img.Registry)
		fmt.Printf("      \"tag\": \"%s\"\n", img.Tag)
		if i < len(analysis.Images)-1 {
			fmt.Printf("    },\n")
		} else {
			fmt.Printf("    }\n")
		}
	}

	fmt.Printf("  ]\n")
	fmt.Printf("}\n")

	return nil
}

// formatBytes formats bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
