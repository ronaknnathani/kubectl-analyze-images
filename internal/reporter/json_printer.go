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
			PodsScanned     int   `json:"podsScanned"`
			NodesScanned    int   `json:"nodesScanned"`
			ImagesAnalyzed  int   `json:"imagesAnalyzed"`
			ImagesInUse     int   `json:"imagesInUse"`
			UnusedImages    int   `json:"unusedImages"`
			SumOfImageSizes int64 `json:"sumOfImageSizesBytes"`
		} `json:"summary"`
		Images []types.Image `json:"images"`
	}{
		Performance: analysis.Performance,
		Images:      analysis.Images,
	}

	report.Summary.PodsScanned = analysis.PodsScanned
	report.Summary.NodesScanned = analysis.NodesScanned
	report.Summary.ImagesAnalyzed = len(analysis.Images)
	report.Summary.ImagesInUse = analysis.ImagesInUse
	report.Summary.UnusedImages = analysis.UnusedImages
	report.Summary.SumOfImageSizes = analysis.TotalSize

	// Use json.NewEncoder to write directly to the writer
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
