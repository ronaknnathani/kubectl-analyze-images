package reporter

import (
	"fmt"
	"io"
	"os"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
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

// GenerateReportTo generates a report to the specified writer
func (r *Reporter) GenerateReportTo(w io.Writer, analysis *types.ImageAnalysis) error {
	var printer types.Printer
	switch r.outputFormat {
	case "table":
		printer = NewTablePrinter(r.showHistogram, r.noColor, r.topImages)
	case "json":
		printer = NewJSONPrinter()
	default:
		return fmt.Errorf("unsupported output format: %s", r.outputFormat)
	}
	return printer.Print(w, analysis)
}

// GenerateReport generates a report to os.Stdout
func (r *Reporter) GenerateReport(analysis *types.ImageAnalysis) error {
	return r.GenerateReportTo(os.Stdout, analysis)
}
