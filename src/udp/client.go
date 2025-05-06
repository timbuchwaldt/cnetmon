package udp

import (
	"cnetmon/metrics"
	"cnetmon/structs"
	"cnetmon/utils"
	"net"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func Connect(target structs.Target, m *metrics.Metrics, labels prometheus.Labels, wg *sync.WaitGroup) {
	defer wg.Done()
	udpServer, err := net.ResolveUDPAddr("udp", target.IP+":7788")

	if err != nil {
		println("ResolveUDPAddr failed:", err.Error())
		// os.Exit(1)
		return
	}

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
	m.Timing.With(utils.Merge(labels, prometheus.Labels{"dst_node": target.NodeName, "dst_pod_ip": udpServer.IP.String()})).Observe(float64(time.Since(start).Milliseconds()))
}
