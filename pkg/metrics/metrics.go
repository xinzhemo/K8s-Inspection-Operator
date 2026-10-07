package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	InspectionTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "inspection_total",
			Help: "集群巡检次数",
		},
	)
	AnomalyDetectedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "anomaly_detected_total",
			Help: "集群巡检异常总数",
		},
		[]string{"namespace", "severity"},
	)
)

func init() {
	metrics.Registry.MustRegister(InspectionTotal)
	metrics.Registry.MustRegister(AnomalyDetectedTotal)
}
