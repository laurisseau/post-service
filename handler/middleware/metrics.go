// middleware/metrics.go
package middleware

import (
    "strconv"
    "time"
	"github.com/laurisseau/post-service/metrics"
    "github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Increment in-flight requests
        metrics.HttpRequestInFlight.Inc()
        defer metrics.HttpRequestInFlight.Dec()

        start := time.Now()

        // Process request
        c.Next()

        // Record request duration
        duration := time.Since(start).Seconds()
        endpoint := c.FullPath()
        if endpoint == "" {
            endpoint = c.Request.URL.Path
        }

        metrics.HttpRequestDuration.WithLabelValues(
            c.Request.Method,
            endpoint,
        ).Observe(duration)

        // Record request count
        metrics.HttpRequestsTotal.WithLabelValues(
            c.Request.Method,
            endpoint,
            strconv.Itoa(c.Writer.Status()),
        ).Inc()
    }
}