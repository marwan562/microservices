package infrastructure

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PaymentRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "payment_requests_total",
		Help: "Total number of payment requests (create/confirm).",
	}, []string{"operation", "status"})

	PaymentLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "payment_latency_seconds",
		Help:    "Latency of payment operations.",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})
)
