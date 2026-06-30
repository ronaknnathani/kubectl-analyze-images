package reporter

import (
	"fmt"
	"io"
	"os"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
)

// Reporter handles output generation
type Reporter struct {
	outputFormat       string
	showHistogram      bool
	noColor            bool
	topImages          int
	truncateImageNames bool
	imageNameParts     int
	sortBy             types.ImageSortBy
}

// NewReporter creates a new reporter
func NewReporter(outputFormat string) *Reporter {
	return &Reporter{
		outputFormat:   outputFormat,
		showHistogram:  true, // Make histogram default
		noColor:        false,
		topImages:      25, // Default to 25 top images
		imageNameParts: 1,
		sortBy:         types.ImageSortBySize,
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

// SetTruncateImageNames enables or disables image name truncation in table output.
func (r *Reporter) SetTruncateImageNames(truncate bool) {
	r.truncateImageNames = truncate
}

// SetImageNameParts sets how many slash-separated image path parts truncation keeps.
func (r *Reporter) SetImageNameParts(parts int) {
	r.imageNameParts = parts
}

// SetSortBy sets the metric used to sort image table rows.
func (r *Reporter) SetSortBy(sortBy types.ImageSortBy) {
	r.sortBy = sortBy
}

// GenerateReportTo generates a report to the specified writer
func (r *Reporter) GenerateReportTo(w io.Writer, analysis *types.ImageAnalysis) error {
	var printer types.Printer
	switch r.outputFormat {
	case "table":
		printer = NewTablePrinter(r.showHistogram, r.noColor, r.topImages, r.truncateImageNames, r.imageNameParts, r.sortBy)
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
