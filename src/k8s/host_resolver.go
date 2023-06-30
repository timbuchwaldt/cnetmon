package k8s

import (
	"cnetmon/metrics"
	"cnetmon/structs"
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func UpdateServiceK8S(lock *sync.Mutex, services *[]structs.Target, m *metrics.Metrics) {
	logger := log.With().Str("component", "updateServiceK8S").Logger()
	ctx := context.Background()
	config := ctrl.GetConfigOrDie()
	clientset := kubernetes.NewForConfigOrDie(config)

	for {
		start := time.Now()
		items, err := clientset.CoreV1().Pods("default").List(ctx, v1.ListOptions{
			LabelSelector: "name=cnetmon",
		})
		m.ResolutionTiming.WithLabelValues("k8s").Observe(float64(time.Since(start).Milliseconds()))

		if err != nil {
			logger.Error().Err(err)
		}
		lock.Lock()
		*services = []structs.Target{}

		for _, p := range items.Items {
			if p.Status.PodIP == "" {
				// This can be null if the IP hasn't been assigned yet -> Hopefully it appears in the next iteration
				break
			}
			t := structs.Target{
				NodeName: p.Spec.NodeName,
				IP:       p.Status.PodIP,
			}
			*services = append(*services, t)
		}
		lock.Unlock()

		m.ResolvedK8SHosts.Set(float64(len(items.Items)))
		time.Sleep(30 * time.Second)
	}
}
