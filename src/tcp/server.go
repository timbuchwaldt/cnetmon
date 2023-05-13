package tcp

import (
	"net"
	"strings"

	"cnetmon/metrics"
	"cnetmon/utils"

	"github.com/rs/zerolog/log"
)

func StartServer(m *metrics.Metrics) {
	service := ":7777"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	utils.CheckErrorFatal(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	utils.CheckErrorFatal(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleTCPClient(conn, m)
	}
}

func handleTCPClient(conn net.Conn, m *metrics.Metrics) {
	connLogger := log.With().Str("component", "handleTCPClient").Str("remoteIP", conn.RemoteAddr().String()).Logger()
	// connLogger.Trace().Msg("New TCP client connected")
	defer conn.Close()

	request := make([]byte, 128)
	for {
		i, err := conn.Read(request)
		if err != nil {
			connLogger.Info().Err(err).Msg("Socket read error")
			break
		}
		if i == 0 {
			connLogger.Info().Msg("0 bytes read, closing connection")
			break
		} else if strings.TrimSpace(string(request[:i])) == "ping" {
			conn.Write([]byte("pong\n"))
		} else if strings.TrimSpace(string(request[:i])) == "bye" {
			break
		} else {
			/// connLogger.Trace().Msg("unknown request, echoing")
			conn.Write(request[:i])
		}
	}
}
