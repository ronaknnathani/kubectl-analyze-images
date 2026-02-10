package plugin

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ronaknnathani/kubectl-analyze-images/internal/analyzer"
	"github.com/ronaknnathani/kubectl-analyze-images/internal/cluster"
	"github.com/ronaknnathani/kubectl-analyze-images/internal/reporter"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/kubernetes"
	"github.com/ronaknnathani/kubectl-analyze-images/pkg/types"
)

// AnalyzeOptions holds all the configuration and dependencies for running image analysis.
// It follows the kubectl plugin Complete/Validate/Run pattern.
type AnalyzeOptions struct {
	// CLI flags
	Namespace     string
	LabelSelector string
	OutputFormat  string
	NoColor       bool
	TopImages     int
	KubeContext   string
	ShowHistogram bool

	// Injected dependencies
	KubernetesClient kubernetes.Interface
	Out              io.Writer
	ErrOut           io.Writer
}

// Complete populates defaults for unset fields and creates the kubernetes client
// if one has not been injected. Tests can pre-inject a FakeClient to skip creation.
func (o *AnalyzeOptions) Complete() error {
	// Set defaults for unset fields
	if o.OutputFormat == "" {
		o.OutputFormat = "table"
	}
	if o.TopImages == 0 {
		o.TopImages = 25
	}
	if o.Out == nil {
		o.Out = os.Stdout
	}
	if o.ErrOut == nil {
		o.ErrOut = os.Stderr
	}

	// Create kubernetes client if not injected (production path)
	if o.KubernetesClient == nil {
		k8sClient, err := kubernetes.NewClient(o.KubeContext)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes client: %w", err)
		}
		o.KubernetesClient = k8sClient
	}

	return nil
}

// Validate checks that all options have valid values.
func (o *AnalyzeOptions) Validate() error {
	// Validate output format
	switch o.OutputFormat {
	case "table", "json":
		// valid
	default:
		return fmt.Errorf("invalid output format %q: must be \"table\" or \"json\"", o.OutputFormat)
	}

	// Validate top images count
	if o.TopImages < 1 {
		return fmt.Errorf("--top-images must be at least 1, got %d", o.TopImages)
	}

	return nil
}

// Run orchestrates the full analysis pipeline: create cluster client, create
// analyzer, run analysis, and generate report.
func (o *AnalyzeOptions) Run(ctx context.Context) error {
	// Create analysis configuration
	config := types.DefaultAnalysisConfig()

	// Create cluster client with injected kubernetes interface
	clusterClient := cluster.NewClient(o.KubernetesClient)

	// Create analyzer with injected cluster client
	podAnalyzer := analyzer.NewPodAnalyzer(clusterClient, config)

	// Display analysis parameters
	namespaceDisplay := o.Namespace
	if namespaceDisplay == "" {
		namespaceDisplay = "All"
	}
	fmt.Fprintf(o.Out, "Analyzing images in namespace: %s\n", namespaceDisplay)
	if o.LabelSelector != "" {
		fmt.Fprintf(o.Out, "Using label selector: %s\n", o.LabelSelector)
	}
	fmt.Fprintln(o.Out)

	// Run analysis
	analysis, err := podAnalyzer.AnalyzePods(ctx, o.Namespace, o.LabelSelector)
	if err != nil {
		return fmt.Errorf("failed to analyze pods: %w", err)
	}

	// Generate report
	rep := reporter.NewReporter(o.OutputFormat)
	rep.SetNoColor(o.NoColor)
	rep.SetTopImages(o.TopImages)
	if err := rep.GenerateReport(analysis); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	return nil
}
