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
	connections map[string]PersistentConnection
}

type PersistentConnection struct {
	c         chan ConnectionMessage
	target    structs.Target
	completed bool
}

type ConnectionMessage struct {
	command string
}

func IsClosed(ch <-chan ConnectionMessage) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func PersistentConnectionManager(outsideAddresses *[]structs.Target, labels prometheus.Labels, mutex *sync.Mutex, m *metrics.Metrics) {
	var pcs = PersistentConnections{
		connections: map[string]PersistentConnection{},
	}

	for {
		mutex.Lock()
		addresses := make([]structs.Target, len(*outsideAddresses))
		copy(addresses, *outsideAddresses)
		mutex.Unlock()

		newConnections := map[string]PersistentConnection{}
		for _, c := range pcs.connections {
			if !c.completed {
				newConnections[c.target.IP] = c
			} else {
				log.Info().Str("remoteIP", c.target.IP).Msg("Removing persistent connection")
			}
		}
		pcs.connections = newConnections

		for _, addr := range addresses {
			_, contains := pcs.connections[addr.IP]
			if !contains {
				log.Info().Str("remoteIP", addr.IP).Msg("Creating persistent connection")
				pcs.connections[addr.IP] = CreatePersistentConnection(addr, labels, m)
			}
		}

		for _, pc := range pcs.connections {
			if IsClosed(pc.c) {
				pc.completed = true
			} else {
				if !utils.IPTargetInSlice(pc.target, addresses) {
					pc.c <- ConnectionMessage{command: "disconnect"}
				}
				pc.c <- ConnectionMessage{command: "ping"}
			}
		}
		time.Sleep(1 * time.Second)
	}

}

func CreatePersistentConnection(target structs.Target, labels prometheus.Labels, m *metrics.Metrics) PersistentConnection {
	pc := PersistentConnection{
		c:      make(chan ConnectionMessage, 30),
		target: target,
	}

	go HandlePersistentConnection(pc, labels, m)
	return pc
}

func HandlePersistentConnection(pc PersistentConnection, labels prometheus.Labels, m *metrics.Metrics) {
	start := time.Now()
	lt, err := m.PersistentLifetime.CurryWith(utils.Merge(labels, prometheus.Labels{"direction": "client", "dst_node": pc.target.NodeName, "dst_pod_ip": pc.target.IP}))
	if err != nil {
		log.Error().Err(err).Msg("Can't label")
		pc.completed = true
		return
	}
	pt, err := m.PingTiming.CurryWith(utils.Merge(labels, prometheus.Labels{"dst_node": pc.target.NodeName, "dst_pod_ip": pc.target.IP}))
	if err != nil {
		log.Error().Err(err).Msg("Can't label")
		pc.completed = true
		return
	}

	connLogger := log.With().Str("component", "PersistentConnection").Str("remoteIP", pc.target.IP).Logger()
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.Dial("tcp", pc.target.IP+":7777")
	if err != nil {

		log.Error().Err(err).Msg("Can't connect")
		pc.completed = true
		return
	}

	defer conn.Close()

	for {
		msg, ok := <-pc.c
		if ok {
			if msg.command == "disconnect" {
				connLogger.Error().Msg("Closing connection")
				pc.completed = true
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
					pc.completed = true
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
