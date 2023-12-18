package gnet

import (
	"encoding/binary"
	"math"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics -
type Metrics struct {
	connections prometheus.Gauge       // 连接
	sensors     *prometheus.CounterVec // 传感器
	packets     *prometheus.CounterVec
	bytes       *prometheus.CounterVec
	timeOffset  *prometheus.GaugeVec
	queuePkts   *prometheus.GaugeVec
	correct     *prometheus.CounterVec
	resend      *prometheus.CounterVec
	dt          *prometheus.GaugeVec
	da          *prometheus.GaugeVec
	dv          *prometheus.GaugeVec
}

// NewMetrics - 初始化统计
func NewMetrics() *Metrics {
	srv := &Metrics{}
	srv.connections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "connections",
		Help:      "gnet number of connections",
	})
	prometheus.MustRegister(srv.connections)

	srv.packets = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "packets",
		Help:      "gnet number of packets",
	}, []string{
		"sensorid",
	})
	prometheus.MustRegister(srv.packets)

	srv.bytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "bytes",
		Help:      "gnet number of bytes",
	}, []string{
		"sensorid",
		"view",
	})
	prometheus.MustRegister(srv.bytes)

	srv.timeOffset = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "aligntime_offset",
		Help:      "gnet alignment time offset in milliseconds",
	}, []string{
		"sensorid",
	})
	prometheus.MustRegister(srv.timeOffset)

	srv.sensors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "sensors",
		Help:      "gnet sensor connection information",
	}, []string{
		"sensorid",
		"addr",
		"hardware_version",
		"firmware_version",
	})
	prometheus.MustRegister(srv.sensors)

	srv.queuePkts = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "queue_pkts",
		Help:      "gnet number of queue packets",
	}, []string{
		"sensorid",
	})
	prometheus.MustRegister(srv.queuePkts)

	srv.correct = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "correct_counter",
		Help:      "gnet number of time alignment corrections",
	}, []string{
		"sensorid",
		"type",
	})
	prometheus.MustRegister(srv.correct)

	srv.resend = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "resend_counter",
		Help:      "get number of time alignment corrections",
	}, []string{
		"sensorid",
		"type",
	})
	prometheus.MustRegister(srv.resend)

	srv.dt = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "data_t",
		Help:      "temperature sampling value",
	}, []string{
		"sensorid",
	})
	prometheus.MustRegister(srv.dt)

	srv.da = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "data_audio_db",
		Help:      "audio db value",
	}, []string{
		"sensorid",
	})
	prometheus.MustRegister(srv.da)

	srv.dv = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "arc",
		Subsystem: "arc-consumer",
		Name:      "data_vibrate_offset",
		Help:      "vibrate offset value",
	}, []string{
		"sensorid",
		"axis",
	})
	prometheus.MustRegister(srv.dv)

	return srv
}

// GetAudioDB -
func GetAudioDB(pcm []byte) float64 {
	var sum float64
	for i := 0; i < len(pcm); i += 2 {
		value := float64(int16(binary.LittleEndian.Uint16(pcm[i:])))
		sum += math.Abs(value)
	}
	sum /= float64(len(pcm) / 2)
	if sum > 0 {
		return 20.0 * math.Log10(sum)
	}
	return 0
}

// GetVibrateOffset -
func GetVibrateOffset(data []byte) (int64, int64, int64) {
	var x int64
	var y int64
	var z int64
	for i := 0; i < len(data); i += 6 {
		x += int64(int16(binary.LittleEndian.Uint16(data[i:])))
		y += int64(int16(binary.LittleEndian.Uint16(data[i+2:])))
		z += int64(int16(binary.LittleEndian.Uint16(data[i+4:])))
	}
	return x / int64(len(data)/6), y / int64(len(data)/6), z / int64(len(data)/6)
}
