package k8s

import (
	"cnetmon/metrics"
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

func UpdateServiceK8S(lock *sync.Mutex, services *[]string, m *metrics.Metrics) {
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
		*services = []string{}

		for _, p := range items.Items {
			*services = append(*services, p.Status.PodIP)
		}
		lock.Unlock()

		m.ResolvedK8SHosts.Set(float64(len(items.Items)))
		time.Sleep(30 * time.Second)
	}
}
