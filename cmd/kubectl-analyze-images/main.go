package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ronaknnathani/kubectl-analyze-images/pkg/plugin"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	o := &plugin.AnalyzeOptions{}

	rootCmd := &cobra.Command{
		Use:   "kubectl-analyze-images",
		Short: "Analyze container images from Kubernetes pods",
		Long: `A kubectl plugin to analyze container images from pods in Kubernetes clusters.
It extracts image sizes from node status and generates reports with performance metrics.`,
		Version:       fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			return o.Run(context.Background())
		},
	}

	// Bind flags directly to AnalyzeOptions fields
	rootCmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "Target namespace (default: all namespaces)")
	rootCmd.Flags().StringVarP(&o.LabelSelector, "selector", "l", "", "Label selector for pods")
	rootCmd.Flags().StringVarP(&o.OutputFormat, "output", "o", "table", "Output format: table, json")
	rootCmd.Flags().BoolVar(&o.NoColor, "no-color", false, "Disable colored output (default: false)")
	rootCmd.Flags().IntVar(&o.TopImages, "top-images", 25, "Number of top images to show in the report (default: 25)")
	rootCmd.Flags().StringVar(&o.KubeContext, "context", "", "Kubernetes context to use (default: current context)")
	rootCmd.Flags().BoolVar(&o.TruncateImageNames, "truncate-image-names", false, "Truncate image names in table output (default: false)")
	rootCmd.Flags().IntVar(&o.ImageNameParts, "image-name-parts", 1, "Number of trailing slash-separated image name parts to keep when truncating")
	rootCmd.Flags().StringVar(&o.SortBy, "sort-by", string(plugin.DefaultSortBy), "Sort image table by: size, pods, cached-on-nodes")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
