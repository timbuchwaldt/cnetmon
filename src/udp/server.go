package udp

import (
	"cnetmon/metrics"
	"cnetmon/utils"

	"net"
	"time"
)

func StartServer(m *metrics.Metrics) {
	udpServer, err := net.ListenPacket("udp", ":7788")
	utils.CheckErrorFatal(err)

	defer udpServer.Close()

	for {
		buf := make([]byte, 1024)
		i, addr, err := udpServer.ReadFrom(buf)
		if err != nil {
			continue
		}

		if i == 0 {
			continue
		}

		if addr != nil && i > 0 {
			udpServer.SetWriteDeadline(time.Now().Add(2 * time.Second))
			udpServer.WriteTo([]byte("yo"), addr)
		}

	}
}
