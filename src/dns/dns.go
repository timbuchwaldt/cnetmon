package dns

import (
	"cnetmon/metrics"
	"cnetmon/structs"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func UpdateServiceDNS(lock *sync.Mutex, services *[]structs.Target, m *metrics.Metrics) {
	logger := log.With().Str("component", "updateServiceDNS").Logger()

	for {

		startDNS := time.Now()
		_, addrs, err := net.LookupSRV("cnetmon", "tcp", "cnetmon-tcp")
		m.ResolutionTiming.WithLabelValues("dns").Observe(float64(time.Since(startDNS).Milliseconds()))

		if err != nil {
			logger.Error().Err(err).Msg("Error resolving DNS")
		}

		// mutex lock here so we can work in peace
		*services = []structs.Target{}
		for _, s := range addrs {
			logger.Debug().Msg(s.Target)
			t := structs.Target{
				NodeName: "unknown",
				IP:       s.Target,
			}
			lock.Lock()
			*services = append(*services, t)
			lock.Unlock()

		}
		m.ResolvedHeadlessServiceHosts.Set(float64(len(*services)))

		time.Sleep(30 * time.Second)
	}
}
