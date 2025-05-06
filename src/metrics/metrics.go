package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	ResolutionTiming             *prometheus.HistogramVec
	ResolvedHeadlessServiceHosts prometheus.Gauge
	ResolvedK8SHosts             prometheus.Gauge
	Timing                       *prometheus.HistogramVec
	PersistentLifetime           *prometheus.GaugeVec
	PingTiming                   *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	m := &Metrics{

		ResolutionTiming: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cnetmon_resolution_timing_milliseconds",
			Help:    "Time the pod resolution takes",
			Buckets: prometheus.ExponentialBuckets(0.125, 2, 16),
		},
			[]string{"mode"},
		),
		ResolvedHeadlessServiceHosts: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cnetmon_resolved_headless_service_hosts",
			Help: "Number of hosts resolved via headless service",
		}),
		ResolvedK8SHosts: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "cnetmon_resolved_k8s_hosts",
			Help: "Number of hosts resolved via kubernetes API",
		}),
		Timing: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cnetmon_timing_milliseconds",
			Help:    "Time the connect check takes",
			Buckets: prometheus.ExponentialBuckets(0.125, 2, 16),
		},
			[]string{"protocol", "mode", "src_node", "dst_node", "dst_pod_ip"},
		),
		PersistentLifetime: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "cnetmon_persistent_connection_lifetime",
			Help: "Time in seconds a persistent connection is open",
		}, []string{"direction", "src_node", "dst_node", "dst_pod_ip"}),
		PingTiming: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "cnetmon_ping_timing_milliseconds",
			Help:    "Time in ms it takes to reply to a ping on a persistent TCP connection",
			Buckets: prometheus.ExponentialBuckets(0.0125, 2, 18),
		}, []string{"src_node", "dst_node", "dst_pod_ip"}),
	}
	return m
}
