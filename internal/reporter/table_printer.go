package reporter

import (
	"fmt"
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/util"
)

// TablePrinter formats output as ASCII tables
type TablePrinter struct {
	showHistogram bool
	noColor       bool
	topImages     int
}

// NewTablePrinter creates a new table printer
func NewTablePrinter(showHistogram, noColor bool, topImages int) *TablePrinter {
	return &TablePrinter{
		showHistogram: showHistogram,
		noColor:       noColor,
		topImages:     topImages,
	}
}

// Print writes the analysis as formatted tables to the provided writer
func (tp *TablePrinter) Print(w io.Writer, analysis *types.ImageAnalysis) error {
	// Performance Summary
	if analysis.Performance != nil {
		fmt.Fprintln(w, "Performance Summary")
		fmt.Fprintln(w, "==================")

		performanceTable := tablewriter.NewWriter(w)
		performanceTable.Header("Metric", "Value")
		if analysis.Performance.PodQueryTime > 0 {
			_ = performanceTable.Append("Pod Query Time", analysis.Performance.PodQueryTime.String())
		}
		if analysis.Performance.NodeQueryTime > 0 {
			_ = performanceTable.Append("Node Query Time", analysis.Performance.NodeQueryTime.String())
		}
		_ = performanceTable.Append("Image Analysis Time", analysis.Performance.ImageAnalysisTime.String())
		_ = performanceTable.Append("Total Time", analysis.Performance.TotalTime.String())
		_ = performanceTable.Append("Images Processed", strconv.Itoa(analysis.Performance.ImagesProcessed))
		_ = performanceTable.Render()
		fmt.Fprintln(w)
	}

	// Image Analysis Summary
	fmt.Fprintln(w, "Image Analysis Summary")
	fmt.Fprintln(w, "=====================")

	summaryTable := tablewriter.NewWriter(w)
	summaryTable.Header("Metric", "Value")
	_ = summaryTable.Append("Total Images", strconv.Itoa(len(analysis.Images)))
	_ = summaryTable.Append("Unique Images", strconv.Itoa(len(analysis.GetUniqueImages())))
	_ = summaryTable.Append("Total Size", util.FormatBytes(analysis.TotalSize))
	_ = summaryTable.Render()
	fmt.Fprintln(w)

	// Image Size Distribution Histogram (if requested and we have images)
	if tp.showHistogram && len(analysis.Images) > 0 {
		fmt.Fprintln(w, "Image Size Distribution")
		fmt.Fprintln(w, "=======================")

		config := types.DefaultHistogramConfig()
		config.Title = "Image Size Distribution"
		config.Height = 15
		config.Width = 60
		config.ShowColors = !tp.noColor // Disable colors if noColor flag is set

		histogramData := analysis.GenerateImageSizeHistogram(config)
		fmt.Fprint(w, histogramData.RenderASCII(config, analysis))
	}

	// Top images by size
	if len(analysis.Images) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Top %d Images by Size\n", tp.topImages)
		fmt.Fprintln(w, "=====================")

		imageTable := tablewriter.NewWriter(w)
		imageTable.Header("Image", "Size")

		topImages := analysis.GetTopImagesBySize(tp.topImages)
		for _, img := range topImages {
			size := util.FormatBytes(img.Size)
			if img.Inaccessible {
				size = "INACCESSIBLE"
			}
			_ = imageTable.Append(img.Name, size)
		}
		_ = imageTable.Render()
		fmt.Fprintln(w)
	}

	return nil
}
