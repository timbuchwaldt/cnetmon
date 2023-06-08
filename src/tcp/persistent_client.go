package tcp

import (
	"cnetmon/metrics"
	"cnetmon/utils"
	"fmt"
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
	addr      string
	completed bool
}

type ConnectionMessage struct {
	command string
}

func PersistentConnectionManager(outsideAddresses *[]string, mutex *sync.Mutex, m *metrics.Metrics) {
	var pcs = PersistentConnections{
		connections: map[string]PersistentConnection{},
	}

	for {
		mutex.Lock()
		addresses := make([]string, len(*outsideAddresses))
		copy(addresses, *outsideAddresses)
		mutex.Unlock()
		fmt.Println(addresses)

		newConnections := map[string]PersistentConnection{}
		for _, c := range pcs.connections {
			if !c.completed {
				newConnections[c.addr] = c
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
			if !utils.StringInSlice(pc.addr, addresses) {
				pc.c <- ConnectionMessage{command: "disconnect"}
			}
			pc.c <- ConnectionMessage{command: "ping"}
		}
		time.Sleep(1 * time.Second)
	}

}

func CreatePersistentConnection(addr string, m *metrics.Metrics) PersistentConnection {
	pc := PersistentConnection{
		c:    make(chan ConnectionMessage),
		addr: addr,
	}

	go HandlePersistentConnection(pc, m)
	return pc
}

func HandlePersistentConnection(pc PersistentConnection, m *metrics.Metrics) {
	//tcpAddr, err := net.ResolveTCPAddr("tcp", addr+":7777")

	//if err != nil {
	//		log.Error().Err(err).Msg("Can't resolve")
	//		return
	//	}
	//labels := append(inLabels, tcpAddr.IP.String())

	start := time.Now()
	lt, err := m.PersistentLifetime.CurryWith(prometheus.Labels{"direction": "client", "hostname": pc.addr})
	pt, err := m.PingTiming.CurryWith(prometheus.Labels{"hostname": pc.addr})
	if err != nil {

		log.Error().Err(err).Msg("Can't label")
		return
	}
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.Dial("tcp", pc.addr+":7777")
	if err != nil {

		log.Error().Err(err).Msg("Can't connect")
		return
	}

	defer conn.Close()

	for {
		msg, ok := <-pc.c
		if ok {
			if msg.command == "disconnect" {
				fmt.Println("Closing connection")
				pc.completed = true
				conn.Close()
				return
			}
			if msg.command == "ping" {
				lt.WithLabelValues().Set(float64(time.Since(start).Seconds()))
				fmt.Println("In Handle Persistent Connection ping", pc)
				pingStart := time.Now()
				conn.Write([]byte("ping"))

				reply := make([]byte, 128)

				conn.SetDeadline(time.Now().Add(2 * time.Second))
				_, err = conn.Read(reply)
				pt.WithLabelValues().Observe(float64(time.Since(pingStart).Milliseconds()))
				if err != nil {
					log.Error().Err(err).Msg("Can't read reply")
					conn.Close()
					return
				}
				conn.SetDeadline(time.Now().Add(30 * time.Second))
			}
		} else {
			fmt.Println("Not OK - weird.")
			pc.completed = true
			break
		}
	}

}
