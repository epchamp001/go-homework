package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	// инфраструктура
	HTTPRequestTotal    *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsErrors  *prometheus.CounterVec

	// бизнес
	OrdersAccepted prometheus.Counter
	OrdersIssued   prometheus.Counter
	OrdersReturned prometheus.Counter

	// кэш
	CacheEvictions prometheus.Counter
}

func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		HTTPRequestTotal: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"handler", "method", "code"},
		),

		HTTPRequestDuration: promauto.With(reg).NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds.",
				Buckets: prometheus.ExponentialBuckets(0.05, 1.7, 8),
			},
			[]string{"handler", "method", "code"},
		),

		HTTPRequestsErrors: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_errors_total",
				Help: "Total number of HTTP requests resulting in error codes (>=400).",
			},
			[]string{"handler", "method", "code"},
		),

		OrdersAccepted: promauto.With(reg).NewCounter(
			prometheus.CounterOpts{
				Name: "pvz_orders_accepted_total",
				Help: "Number of orders accepted from couriers.",
			}),
		OrdersIssued: promauto.With(reg).NewCounter(
			prometheus.CounterOpts{
				Name: "pvz_orders_issued_total",
				Help: "Number of orders issued to clients.",
			}),
		OrdersReturned: promauto.With(reg).NewCounter(
			prometheus.CounterOpts{
				Name: "pvz_orders_returned_total",
				Help: "Number of orders returned (courier or client).",
			}),
		CacheEvictions: promauto.With(reg).NewCounter(
			prometheus.CounterOpts{
				Name: "pvz_cache_evictions_total",
				Help: "Keys removed from in-memory cache (TTL or capacity).",
			}),
	}
	return m
}

func (m *Metrics) IncOrdersAccepted() {
	m.OrdersAccepted.Inc()
}

func (m *Metrics) IncOrdersIssued() {
	m.OrdersIssued.Inc()
}

func (m *Metrics) IncOrdersReturned() {
	m.OrdersReturned.Inc()
}
