package main

import (
	"cnetmon/dns"
	"cnetmon/generic_client"
	"cnetmon/k8s"
	"cnetmon/metrics"
	"cnetmon/structs"
	"cnetmon/tcp"
	"cnetmon/udp"

	"net/http"
	"sync"

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
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	m := metrics.NewMetrics()
	log.Info().Msg("Starting continuous network monitoring cnetmon")

	go tcp.StartServer(m)
	go udp.StartServer(m)

	go generic_client.Connect(&resolveK8S, &resolveK8SLock, m, []string{"k8s", "tcp"}, tcp.Connect)
	go generic_client.Connect(&resolveDNSServices, &resolveDNSServicesLock, m, []string{"dns", "tcp"}, tcp.Connect)

	go generic_client.Connect(&resolveK8S, &resolveK8SLock, m, []string{"k8s", "udp"}, udp.Connect)
	go generic_client.Connect(&resolveDNSServices, &resolveDNSServicesLock, m, []string{"dns", "udp"}, udp.Connect)

	go k8s.UpdateServiceK8S(&resolveK8SLock, &resolveK8S, m)
	go dns.UpdateServiceDNS(&resolveDNSServicesLock, &resolveDNSServices, m)

	go tcp.PersistentConnectionManager(&resolveK8S, &resolveK8SLock, m)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2808", nil)
}
