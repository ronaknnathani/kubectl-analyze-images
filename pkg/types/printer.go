package types

import "io"

// Printer defines the interface for output formatters.
// Implementations write analysis results to the provided writer.
type Printer interface {
	Print(w io.Writer, analysis *ImageAnalysis) error
}
