package generic_client

import (
	"cnetmon/metrics"
	"cnetmon/structs"
	"sync"
	"time"
)

func Connect(outsideAddresses *[]structs.Target, mutex *sync.Mutex, m *metrics.Metrics, labels []string, function func(structs.Target, *metrics.Metrics, []string, *sync.WaitGroup)) {

	for {
		// be ultra-cautious here so we are 100% sure the slice isn't just updated.
		mutex.Lock()
		addresses := make([]structs.Target, len(*outsideAddresses))
		copy(addresses, *outsideAddresses)
		mutex.Unlock()

		// allow async checking and waiting for all to finish
		var wg = &sync.WaitGroup{}

		for _, addr := range addresses {
			wg.Add(1)
			go function(addr, m, labels, wg)
		}
		wg.Wait()

		time.Sleep(1 * time.Second)
	}
}
