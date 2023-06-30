package tcp

import (
	"cnetmon/metrics"
	"cnetmon/structs"
	"cnetmon/utils"
	"net"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type PersistentConnections struct {
	connections map[structs.Target]PersistentConnection
}

type PersistentConnection struct {
	c         chan ConnectionMessage
	target    structs.Target
	completed bool
}

type ConnectionMessage struct {
	command string
}

func PersistentConnectionManager(outsideAddresses *[]structs.Target, mutex *sync.Mutex, m *metrics.Metrics) {
	var pcs = PersistentConnections{
		connections: map[structs.Target]PersistentConnection{},
	}

	for {
		mutex.Lock()
		addresses := make([]structs.Target, len(*outsideAddresses))
		copy(addresses, *outsideAddresses)
		mutex.Unlock()

		newConnections := map[structs.Target]PersistentConnection{}
		for _, c := range pcs.connections {
			if !c.completed {
				newConnections[c.target] = c
			}
		}
		pcs.connections = newConnections

		for _, addr := range addresses {
			_, contains := pcs.connections[addr]
			if !contains {
				pcs.connections[addr] = CreatePersistentConnection(addr, m)
			}
		}

		for _, pc := range pcs.connections {
			if !utils.IPTargetInSlice(pc.target, addresses) {
				pc.c <- ConnectionMessage{command: "disconnect"}
			}
			pc.c <- ConnectionMessage{command: "ping"}
		}
		time.Sleep(1 * time.Second)
	}

}

func CreatePersistentConnection(target structs.Target, m *metrics.Metrics) PersistentConnection {
	pc := PersistentConnection{
		c:      make(chan ConnectionMessage),
		target: target,
	}

	go HandlePersistentConnection(pc, m)
	return pc
}

func HandlePersistentConnection(pc PersistentConnection, m *metrics.Metrics) {
	start := time.Now()
	lt, err := m.PersistentLifetime.CurryWith(prometheus.Labels{"direction": "client", "node_name": pc.target.NodeName, "pod_ip": pc.target.IP})
	if err != nil {
		log.Error().Err(err).Msg("Can't label")
		return
	}
	pt, err := m.PingTiming.CurryWith(prometheus.Labels{"node_name": pc.target.NodeName, "pod_ip": pc.target.IP})
	if err != nil {
		log.Error().Err(err).Msg("Can't label")
		return
	}

	connLogger := log.With().Str("component", "PersistentConnection").Str("remoteIP", pc.target.IP).Logger()
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.Dial("tcp", pc.target.IP+":7777")
	if err != nil {

		log.Error().Err(err).Msg("Can't connect")
		return
	}

	defer conn.Close()

	for {
		msg, ok := <-pc.c
		if ok {
			if msg.command == "disconnect" {
				connLogger.Error().Msg("Closing connection")
				pc.completed = true
				conn.Close()
				return
			}
			if msg.command == "ping" {
				lt.WithLabelValues().Set(float64(time.Since(start).Seconds()))
				connLogger.Trace().Msg("In Handle Persistent Connection ping")
				pingStart := time.Now()
				conn.Write([]byte("ping"))

				reply := make([]byte, 128)

				conn.SetDeadline(time.Now().Add(2 * time.Second))
				_, err = conn.Read(reply)
				pt.WithLabelValues().Observe(float64(time.Since(pingStart).Milliseconds()))
				if err != nil {
					connLogger.Error().Err(err).Msg("Can't read reply")
					conn.Close()
					return
				}
				conn.SetDeadline(time.Now().Add(30 * time.Second))
			}
		} else {
			connLogger.Warn().Msg("Not OK - weird.")
			pc.completed = true
			break
		}
	}

}
