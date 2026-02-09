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
	PodQueryTime       time.Duration
	NodeQueryTime      time.Duration
	ImageAnalysisTime  time.Duration
	TotalTime          time.Duration
	ImagesProcessed    int
	ImagesFailed       int
	ImagesInaccessible int
	CacheHits          int
	CacheMisses        int
}
