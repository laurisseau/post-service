// metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP metrics
    HttpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    HttpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    HttpRequestInFlight = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "http_requests_in_flight",
            Help: "Current number of HTTP requests being processed",
        },
    )

    // Database metrics
    DatabaseQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "database_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"query_type", "table"},
    )

    DatabaseErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "database_errors_total",
            Help: "Total number of database errors",
        },
        []string{"query_type", "error_type"},
    )

    // Business metrics
    PostsCreatedTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "posts_created_total",
            Help: "Total number of posts created",
        },
    )

    PostsUpdatedTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "posts_updated_total",
            Help: "Total number of posts updated",
        },
    )

    PostsDeletedTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "posts_deleted_total",
            Help: "Total number of posts deleted",
        },
    )

    ActivePostsCount = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_posts_count",
            Help: "Current number of active posts",
        },
    )
)

// Initialize metrics
func Init() {

    // Register custom metrics if needed
    prometheus.MustRegister(HttpRequestsTotal, HttpRequestDuration, HttpRequestInFlight)
    prometheus.MustRegister(DatabaseQueryDuration, DatabaseErrorsTotal)
    prometheus.MustRegister(PostsCreatedTotal, PostsUpdatedTotal, PostsDeletedTotal, ActivePostsCount)

}