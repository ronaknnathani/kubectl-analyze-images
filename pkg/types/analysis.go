package types

import (
	"time"
)

// AnalysisConfig holds configuration for image analysis
type AnalysisConfig struct {
	PodPageSize int64 // Number of pods to fetch per page
}

// DefaultAnalysisConfig returns default configuration
func DefaultAnalysisConfig() *AnalysisConfig {
	return &AnalysisConfig{
		PodPageSize: 500,
	}
}

// PerformanceMetrics holds timing and performance data
type PerformanceMetrics struct {
	PodQueryTime       time.Duration `json:"podQueryTime"`
	NodeQueryTime      time.Duration `json:"nodeQueryTime"`
	ImageAnalysisTime  time.Duration `json:"imageAnalysisTime"`
	TotalTime          time.Duration `json:"totalTime"`
	ImagesProcessed    int           `json:"imagesProcessed"`
	ImagesFailed       int           `json:"imagesFailed"`
	ImagesInaccessible int           `json:"imagesInaccessible"`
}
