package reporter

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/util"
)

// TablePrinter formats output as ASCII tables
type TablePrinter struct {
	showHistogram      bool
	noColor            bool
	topImages          int
	truncateImageNames bool
	imageNameParts     int
	sortBy             types.ImageSortBy
}

// NewTablePrinter creates a new table printer
func NewTablePrinter(showHistogram, noColor bool, topImages int, truncateImageNames bool, imageNameParts int, sortBy types.ImageSortBy) *TablePrinter {
	return &TablePrinter{
		showHistogram:      showHistogram,
		noColor:            noColor,
		topImages:          topImages,
		truncateImageNames: truncateImageNames,
		imageNameParts:     imageNameParts,
		sortBy:             sortBy,
	}
}

// Print writes the analysis as formatted tables to the provided writer
func (tp *TablePrinter) Print(w io.Writer, analysis *types.ImageAnalysis) error {
	// Performance Summary
	if analysis.Performance != nil {
		if err := writeLines(w, "Performance Summary", "=================="); err != nil {
			return err
		}

		performanceTable := tablewriter.NewWriter(w)
		performanceTable.Header("Metric", "Value")
		if analysis.Performance.PodQueryTime > 0 {
			if err := performanceTable.Append("Pod Query Time", analysis.Performance.PodQueryTime.String()); err != nil {
				return fmt.Errorf("failed to append pod query time: %w", err)
			}
		}
		if analysis.Performance.NodeQueryTime > 0 {
			if err := performanceTable.Append("Node Query Time", analysis.Performance.NodeQueryTime.String()); err != nil {
				return fmt.Errorf("failed to append node query time: %w", err)
			}
		}
		if err := appendRows(performanceTable,
			[]string{"Image Analysis Time", analysis.Performance.ImageAnalysisTime.String()},
			[]string{"Total Time", analysis.Performance.TotalTime.String()},
			[]string{"Images Processed", strconv.Itoa(analysis.Performance.ImagesProcessed)},
		); err != nil {
			return err
		}
		if err := performanceTable.Render(); err != nil {
			return fmt.Errorf("failed to render performance table: %w", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("failed to write performance table spacer: %w", err)
		}
	}

	// Image Analysis Summary
	if err := writeLines(w, "Image Analysis Summary", "====================="); err != nil {
		return err
	}

	summaryTable := tablewriter.NewWriter(w)
	summaryTable.Header("Metric", "Value")
	if err := appendRows(summaryTable,
		[]string{"Pods Scanned", strconv.Itoa(analysis.PodsScanned)},
		[]string{"Nodes Scanned", strconv.Itoa(analysis.NodesScanned)},
		[]string{"Images Analyzed", strconv.Itoa(len(analysis.Images))},
		[]string{"Images In Use", strconv.Itoa(analysis.ImagesInUse)},
		[]string{"Images Not Used By Pods", strconv.Itoa(analysis.UnusedImages)},
		[]string{"Sum Of Image Sizes", util.FormatBytes(analysis.TotalSize)},
	); err != nil {
		return err
	}
	if err := summaryTable.Render(); err != nil {
		return fmt.Errorf("failed to render summary table: %w", err)
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("failed to write summary table spacer: %w", err)
	}

	// Image Size Distribution Histogram (if requested and we have images)
	if tp.showHistogram && len(analysis.Images) > 0 {
		if err := writeLines(w, "Image Size Distribution", "======================="); err != nil {
			return err
		}

		config := types.DefaultHistogramConfig()
		config.Title = "Image Size Distribution"
		config.Height = 15
		config.Width = 60
		config.ShowColors = !tp.noColor // Disable colors if noColor flag is set

		histogramData := analysis.GenerateImageSizeHistogram(config)
		if _, err := fmt.Fprint(w, histogramData.RenderASCII(config, analysis)); err != nil {
			return fmt.Errorf("failed to write histogram: %w", err)
		}
	}

	// Top images by size
	if len(analysis.Images) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("failed to write image table spacer: %w", err)
		}
		if _, err := fmt.Fprintf(w, "Top %d Images by %s\n", tp.topImages, sortDisplayName(tp.sortBy)); err != nil {
			return fmt.Errorf("failed to write image table title: %w", err)
		}
		if _, err := fmt.Fprintln(w, "=============================="); err != nil {
			return fmt.Errorf("failed to write image table underline: %w", err)
		}

		imageTable := tablewriter.NewWriter(w)
		imageTable.Header("Image", "Pods", "Containers", "Init Containers", "Namespaces", "Size", "Cached On Nodes")

		topImages := analysis.GetTopImages(tp.topImages, tp.sortBy)
		for _, img := range topImages {
			size := util.FormatBytes(img.Size)
			if img.Inaccessible {
				size = "INACCESSIBLE"
			}
			if err := imageTable.Append(
				tp.displayImageName(img.Name),
				strconv.Itoa(img.PodCount),
				strconv.Itoa(img.ContainerCount),
				strconv.Itoa(img.InitContainerCount),
				strconv.Itoa(img.NamespaceCount),
				size,
				strconv.Itoa(img.CachedOnNodes),
			); err != nil {
				return fmt.Errorf("failed to append image row for %q: %w", img.Name, err)
			}
		}
		if err := imageTable.Render(); err != nil {
			return fmt.Errorf("failed to render image table: %w", err)
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("failed to write image table ending newline: %w", err)
		}
	}

	return nil
}

func (tp *TablePrinter) displayImageName(name string) string {
	if !tp.truncateImageNames {
		return name
	}
	parts := strings.Split(name, "/")
	if tp.imageNameParts >= len(parts) {
		return name
	}
	return strings.Join(parts[len(parts)-tp.imageNameParts:], "/")
}

func sortDisplayName(sortBy types.ImageSortBy) string {
	switch sortBy {
	case types.ImageSortByPods:
		return "Pod Usage"
	case types.ImageSortByCachedOnNodes:
		return "Cached On Nodes"
	default:
		return "Size"
	}
}

func writeLines(w io.Writer, lines ...string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return fmt.Errorf("failed to write line %q: %w", line, err)
		}
	}
	return nil
}

func appendRows(table *tablewriter.Table, rows ...[]string) error {
	for _, row := range rows {
		if err := table.Append(row[0], row[1]); err != nil {
			return fmt.Errorf("failed to append table row %q: %w", strings.Join(row, "="), err)
		}
	}
	return nil
}
