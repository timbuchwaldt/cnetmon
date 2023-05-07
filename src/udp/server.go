package udp

import (
	"log"
	"net"
	"time"
)

func StartServer() {
	udpServer, err := net.ListenPacket("udp", ":7788")
	if err != nil {
		log.Fatal(err)
	}
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
