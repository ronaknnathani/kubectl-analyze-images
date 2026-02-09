package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
)

// JSONPrinter formats output as JSON
type JSONPrinter struct{}

// NewJSONPrinter creates a new JSON printer
func NewJSONPrinter() *JSONPrinter {
	return &JSONPrinter{}
}

// Print writes the analysis as JSON to the provided writer
func (jp *JSONPrinter) Print(w io.Writer, analysis *types.ImageAnalysis) error {
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

	// Use json.NewEncoder to write directly to the writer
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
