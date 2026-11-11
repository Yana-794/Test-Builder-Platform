package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TestsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "tests_created_total",
			Help: "Total tests created",
		},
	)

	RateLimitedRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limited_requests_total",
			Help: "Requests rejected by rate limiter",
		},
		[]string{"scope"}, 
	)
)