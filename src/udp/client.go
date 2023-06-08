package udp

import (
	"cnetmon/metrics"
	"net"
	"sync"
	"time"
)

func Connect(addr string, m *metrics.Metrics, inLabels []string, wg *sync.WaitGroup) {

	defer wg.Done()
	udpServer, err := net.ResolveUDPAddr("udp", addr+":7788")

	if err != nil {
		println("ResolveUDPAddr failed:", err.Error())
		// os.Exit(1)
		return
	}
	labels := append(inLabels, udpServer.IP.String())

	start := time.Now()
	conn, err := net.DialUDP("udp", nil, udpServer)
	if err != nil {
		println("Listen failed:", err.Error())
		// os.Exit(1)
		return
	}

	// close the connection
	_, err = conn.Write([]byte("This is a UDP message"))
	if err != nil {
		println("Write data failed:", err.Error())
		//os.Exit(1)
		return
	}

	conn.SetDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		println("read data failed:", err.Error())
		//os.Exit(1)
		return
	}
	conn.Close()
	m.Timing.WithLabelValues(labels...).Observe(float64(time.Since(start).Milliseconds()))
}
