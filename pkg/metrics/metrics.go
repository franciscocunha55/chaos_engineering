package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var ChaosPodsDeletedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "chaos_pods_deleted_total",
		Help: "Total number of pods deleted by chaos engineering tests",
	},
	[]string{"namespace"},
)

func Register() error{
	if err := prometheus.Register(ChaosPodsDeletedCounter); err != nil {
		return fmt.Errorf("failed to register chaos_pods_deleted_total: %s", err)
	}
	return nil
}