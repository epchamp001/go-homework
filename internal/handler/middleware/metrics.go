package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"pvz-cli/internal/metrics"
	"time"
)

func MetricMiddleware(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		handler := c.FullPath()
		if handler == "" {
			handler = c.Request.URL.Path
		}
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		code := fmt.Sprintf("%d", status)
		labels := prometheus.Labels{
			"handler": handler,
			"method":  method,
			"code":    code,
		}

		m.HTTPRequestTotal.With(labels).Inc()
		m.HTTPRequestDuration.With(labels).
			Observe(time.Since(start).Seconds())
		if status >= 400 {
			m.HTTPRequestsErrors.With(labels).Inc()
		}
	}
}

func MetricsEndpoint(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
