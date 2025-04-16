package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "server_active_connections",
		Help: "Number of active connections",
	})

	TotalConnections = promauto.NewCounter(prometheus.CounterOpts{
		Name: "server_total_connections",
		Help: "Total number of connections",
	})

	PoWChallengesGenerated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "server_pow_challenges_generated",
		Help: "Number of PoW challenges generated",
	})

	PoWChallengesVerified = promauto.NewCounter(prometheus.CounterOpts{
		Name: "server_pow_challenges_verified",
		Help: "Number of PoW challenges verified",
	})

	PoWVerificationFailures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "server_pow_verification_failures",
		Help: "Number of PoW verification failures",
	})

	RateLimitHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "server_rate_limit_hits",
		Help: "Number of rate limit hits",
	})

	WorkerPoolSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "server_worker_pool_size",
		Help: "Current size of the worker pool",
	})

	WorkerPoolQueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "server_worker_pool_queue_size",
		Help: "Current size of the worker pool queue",
	})

	ResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "server_response_time_seconds",
		Help:    "Response time in seconds",
		Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
	})
)
