package internals

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var constLabels = map[string]string{"app": "chart-sync"}

// buckets duration - 10 minutes
var CustomBuckets = []float64{1, 5, 15, 30, 60, 120, 180, 240, 300, 600}

var (
	ChartVersionsProcessed = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "chart_versions_processed_seconds",
			Help:        "Total number of chart versions processed successfully",
			ConstLabels: constLabels,
			Buckets:     CustomBuckets,
		},
		[]string{"repo_type", "chart_name", "status"})

	RepoSyncDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "repo_sync_duration_seconds",
			Help:        "Time taken to sync an entire repository",
			ConstLabels: constLabels,
			Buckets:     CustomBuckets,
		},
		[]string{"repo_type", "repo_name", "error_type"})
)
