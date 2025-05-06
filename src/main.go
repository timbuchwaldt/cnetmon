package main

import (
	"cnetmon/dns"
	"cnetmon/generic_client"
	"cnetmon/k8s"
	"cnetmon/metrics"
	"cnetmon/structs"
	"cnetmon/tcp"
	"cnetmon/udp"
	"cnetmon/utils"
	"os"

	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "go.uber.org/automaxprocs"
)

var resolveDNSServicesLock sync.Mutex
var resolveDNSServices []structs.Target

var resolveK8SLock sync.Mutex
var resolveK8S []structs.Target

func main() {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		log.Fatal().Msg("NODE_NAME environment variable is not set")
	}
	metricsLabels := prometheus.Labels{"src_node": nodeName}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	m := metrics.NewMetrics()
	log.Info().Msg("Starting continuous network monitoring cnetmon")

	go tcp.StartServer(m)
	go udp.StartServer(m)

	go generic_client.Connect(&resolveK8S, &resolveK8SLock, m, utils.Merge(metricsLabels, prometheus.Labels{"mode": "k8s", "protocol": "tcp"}), tcp.Connect)
	go generic_client.Connect(&resolveDNSServices, &resolveDNSServicesLock, m, utils.Merge(metricsLabels, prometheus.Labels{"mode": "dns", "protocol": "tcp"}), tcp.Connect)

	go generic_client.Connect(&resolveK8S, &resolveK8SLock, m, utils.Merge(metricsLabels, prometheus.Labels{"mode": "k8s", "protocol": "udp"}), udp.Connect)
	go generic_client.Connect(&resolveDNSServices, &resolveDNSServicesLock, m, utils.Merge(metricsLabels, prometheus.Labels{"mode": "dns", "protocol": "udp"}), udp.Connect)

	go k8s.UpdateServiceK8S(&resolveK8SLock, &resolveK8S, m)
	go dns.UpdateServiceDNS(&resolveDNSServicesLock, &resolveDNSServices, m)

	go tcp.PersistentConnectionManager(&resolveK8S, metricsLabels, &resolveK8SLock, m)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2808", nil)
}
