package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ronaknnathani/kubectl-analyze-images/internal/analyzer"
	"github.com/ronaknnathani/kubectl-analyze-images/internal/cluster"
	"github.com/ronaknnathani/kubectl-analyze-images/internal/reporter"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/kubernetes"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var namespace, labelSelector, outputFormat string
	var noColor bool
	var topImages int
	var kubeContext string

	rootCmd := &cobra.Command{
		Use:   "kubectl-analyze-images",
		Short: "Analyze container images from Kubernetes pods",
		Long: `A kubectl plugin to analyze container images from pods in Kubernetes clusters.
It extracts image sizes from node status and generates reports with performance metrics.`,
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyze(namespace, labelSelector, outputFormat, noColor, topImages, kubeContext)
		},
	}

	// Add flags
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Target namespace (default: all namespaces)")
	rootCmd.Flags().StringVarP(&labelSelector, "selector", "l", "", "Label selector for pods")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output (default: false)")
	rootCmd.Flags().IntVar(&topImages, "top-images", 25, "Number of top images to show in the report (default: 25)")
	rootCmd.Flags().StringVar(&kubeContext, "context", "", "Kubernetes context to use (default: current context)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runAnalyze executes the image analysis
func runAnalyze(namespace, labelSelector, outputFormat string, noColor bool, topImages int, kubeContext string) error {
	ctx := context.Background()

	// Create configuration
	config := types.DefaultAnalysisConfig()

	k8sClient, err := kubernetes.NewClient(kubeContext)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	clusterClient := cluster.NewClient(k8sClient)
	podAnalyzer := analyzer.NewPodAnalyzer(clusterClient, config)

	// Analyze pods
	namespaceDisplay := namespace
	if namespaceDisplay == "" {
		namespaceDisplay = "All"
	}
	fmt.Printf("Analyzing images in namespace: %s\n", namespaceDisplay)
	if labelSelector != "" {
		fmt.Printf("Using label selector: %s\n", labelSelector)
	}
	fmt.Println()

	analysis, err := podAnalyzer.AnalyzePods(ctx, namespace, labelSelector)
	if err != nil {
		return fmt.Errorf("failed to analyze pods: %w", err)
	}

	// Generate report
	reporter := reporter.NewReporter(outputFormat)
	reporter.SetNoColor(noColor)
	reporter.SetTopImages(topImages)
	if err := reporter.GenerateReport(analysis); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	return nil
}
