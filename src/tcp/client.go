package tcp

import (
	"cnetmon/metrics"
	"cnetmon/structs"
	"cnetmon/utils"
	"math"
	"net"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

func Connect(target structs.Target, m *metrics.Metrics, labels prometheus.Labels, wg *sync.WaitGroup) {
	defer wg.Done()
	tcpAddr, err := net.ResolveTCPAddr("tcp", target.IP+":7777")

	if err != nil {
		log.Error().Err(err).Msg("Can't resolve")
		return
	}

	metric := m.Timing.With(utils.Merge(labels, prometheus.Labels{"dst_node": target.NodeName, "dst_pod_ip": tcpAddr.IP.String()}))

	start := time.Now()
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.Dial("tcp", target.IP+":7777")
	if err != nil {

		log.Error().Err(err).Msg("Can't connect")
		metric.Observe(math.Inf(1))
		return
	}
	conn.Write([]byte("ping"))

	reply := make([]byte, 128)

	conn.SetDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(reply)
	if err != nil {
		log.Error().Err(err).Msg("Can't read reply")
		metric.Observe(math.Inf(1))
		return
	}
	metric.Observe(float64(time.Since(start).Milliseconds()))

	conn.Write([]byte("bye"))

	conn.Close()
}
