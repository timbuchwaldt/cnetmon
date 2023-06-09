package tcp

import (
	"cnetmon/metrics"
	"cnetmon/structs"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func Connect(target structs.Target, m *metrics.Metrics, inLabels []string, wg *sync.WaitGroup) {
	defer wg.Done()
	tcpAddr, err := net.ResolveTCPAddr("tcp", target.IP+":7777")

	if err != nil {
		log.Error().Err(err).Msg("Can't resolve")
		return
	}
	labels := append(append(inLabels, target.NodeName), tcpAddr.IP.String())

	start := time.Now()
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.Dial("tcp", target.IP+":7777")
	if err != nil {

		log.Error().Err(err).Msg("Can't connect")
		return
	}
	conn.Write([]byte("ping"))

	reply := make([]byte, 128)

	conn.SetDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(reply)
	if err != nil {
		log.Error().Err(err).Msg("Can't read reply")
		return
	}
	m.Timing.WithLabelValues(labels...).Observe(float64(time.Since(start).Milliseconds()))

	conn.Write([]byte("bye"))

	conn.Close()
}
