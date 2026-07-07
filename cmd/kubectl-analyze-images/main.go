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
		Use:   "kubectl analyze-images",
		Short: "Visualize Kubernetes image usage and sizes",
		Long: `Visualize the largest and most-used container images in a Kubernetes cluster.
The report combines image sizes, pod usage, and cached-on-node counts in one easy-to-scan view.`,
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
	rootCmd.SetVersionTemplate("kubectl analyze-images version {{.Version}}\n")

	// Bind flags directly to AnalyzeOptions fields
	rootCmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "Target namespace (default: all namespaces)")
	rootCmd.Flags().StringVarP(&o.LabelSelector, "selector", "l", "", "Label selector for pods")
	rootCmd.Flags().StringVarP(&o.OutputFormat, "output", "o", "table", "Output format: table, json")
	rootCmd.Flags().BoolVar(&o.NoColor, "no-color", false, "Disable colored output (default: false)")
	rootCmd.Flags().IntVar(&o.TopImages, "top-images", 25, "Number of top images to show in the report (default: 25)")
	rootCmd.Flags().StringVar(&o.KubeContext, "context", "", "Kubernetes context to use (default: current context)")
	rootCmd.Flags().StringVar(&o.Kubeconfig, "kubeconfig", "", "Path to the kubeconfig file (default: standard kubectl loading rules)")
	rootCmd.Flags().BoolVar(&o.TruncateImageNames, "truncate-image-names", false, "Truncate image names in table output (default: false)")
	rootCmd.Flags().IntVar(&o.ImageNameParts, "image-name-parts", 1, "Number of trailing slash-separated image name parts to keep when truncating")
	rootCmd.Flags().StringVar(&o.SortBy, "sort-by", string(plugin.DefaultSortBy), "Sort image table by: size, pods, cached-on-nodes")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
