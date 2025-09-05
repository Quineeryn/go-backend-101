package httpx

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"route", "method", "status"},
	)
	httpLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_latency_seconds",
			Help:    "Latency of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route", "method"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests, httpLatency)
}

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		method := c.Request.Method
		status := fmt.Sprint(c.Writer.Status())
		httpRequests.WithLabelValues(route, method, status).Inc()
		httpLatency.WithLabelValues(route, method).Observe(time.Since(start).Seconds())
	}
}
