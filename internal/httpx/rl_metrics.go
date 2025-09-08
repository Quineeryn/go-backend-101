package httpx

import "github.com/prometheus/client_golang/prometheus"

var RateLimitExceeded = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "rate_limit_exceeded_total",
		Help: "Total requests rejected by rate limiter",
	},
	[]string{"route"},
)

func init() {
	prometheus.MustRegister(RateLimitExceeded)
}
