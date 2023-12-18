package httpproxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics -
type Metrics struct {
	queuePkts prometheus.Gauge // 队列包
	packets   *prometheus.CounterVec
}

// NewMetrics - 注册代理配置数量连接
// @return *Metrics 代理配置数量结构
func NewMetrics() *Metrics {
	srv := &Metrics{}
	srv.queuePkts = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "http_proxy_queue_pkts",
		Help:      "Number of queue packets",
	})
	prometheus.MustRegister(srv.queuePkts)

	srv.packets = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "http_proxy_packets",
		Help:      "Number of packets",
	}, []string{
		"sensorid",
	})
	prometheus.MustRegister(srv.packets)
	return srv
}
