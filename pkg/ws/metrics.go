package ws

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics -
type Metrics struct {
	connections prometheus.Gauge // 连接数量
}

// NewMetrics -
func NewMetrics() *Metrics {
	srv := &Metrics{}
	srv.connections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "ws_connections",
		Help:      "Number of websocket connections",
	})
	prometheus.MustRegister(srv.connections)
	return srv
}
