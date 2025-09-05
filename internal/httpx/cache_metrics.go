package httpx

import "github.com/prometheus/client_golang/prometheus"

var (
	CacheHit = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "cache_hit_total", Help: "Total cache hits"},
		[]string{"resource"},
	)
	CacheMiss = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "cache_miss_total", Help: "Total cache misses"},
		[]string{"resource"},
	)
)

func init() {
	prometheus.MustRegister(CacheHit, CacheMiss)
}
