package grpc

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics -
type Metrics struct {
	connections prometheus.Gauge // 连接数量
	packets     *prometheus.CounterVec
}

// NewMetrics - 注册代理配置数量连接
// @return *Metrics 代理配置数量结构
func NewMetrics() *Metrics {
	srv := &Metrics{}
	srv.connections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "grpc_connections",
		Help:      "Grpc currently available connections in the pool",
	})
	prometheus.MustRegister(srv.connections)

	srv.packets = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "grpc_packet",
		Help:      "Grpc number of packets",
	}, []string{
		"sensorid",
	})
	prometheus.MustRegister(srv.packets)
	return srv
}