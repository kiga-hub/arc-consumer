package ws

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/protocols"
)

// Handler -
type Handler interface {
	Start(context.Context)
	Write(uint64, *protocols.Frame) error
	Register(*websocket.Conn) *Client
	GetSensorids() Uint64Slice
}

// Server websocket 管理结构
type Server struct {
	clients   *sync.Map
	maxConnID int64
	mutex     sync.Mutex
	timeout   int64
	config    *Config
	logger    logging.ILogger
	metrics   *Metrics
}

// New - 初始化创建管理结构
func New(opts ...Option) (Handler, error) {
	srv := loadOptions(opts...)

	// 判断配置是否开启这个模块
	if !srv.config.Enable {
		return nil, nil
	}

	spew.Dump(srv.config)

	srv.clients = new(sync.Map)
	srv.maxConnID = 0
	srv.timeout = 1000
	srv.metrics = NewMetrics()
	return srv, nil
}

// Start - 启动
func (s *Server) Start(ctx context.Context) {
	s.logger.Info("websocket server start")
}

// Write - 将数据包数据广播到各个websocket数据连接中
func (s *Server) Write(id uint64, frame *protocols.Frame) error {
	s.clients.Range(func(key, value interface{}) bool {
		c := value.(*Client)
		select {
		case <-c.Done():
			c.logger.Debugw("websocket disconnect")
			s.clients.Delete(key)
			s.metrics.connections.Dec()
			return true
		default:
			// 检查是否需要转发这个传感器的数据
			v, ok := c.data.Load(id)
			if !ok {
				return true
			}

			// 发送数据
			c.Send(v.(*ClientBuffer), frame)
		}
		return true
	})
	return nil
}

// Register - 基于已有的连接创建
func (s *Server) Register(conn *websocket.Conn) *Client {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// wsConn := NewWebsocket(s.logger, conn)
	srv := &Client{
		ws:             conn,
		connectedAt:    time.Now().Unix(),
		outChan:        make(chan *Message, 1024*1024),
		data:           new(sync.Map),
		audioMaxSize:   1024 * 1024,
		vibrateMaxSize: 1024 * 1024,
		fill:           s.config.Fill,
		fillEnable:     s.config.FillEnable,
		logger:         s.logger,
		closeChan:      make(chan struct{}),
	}
	srv.Start(context.Background())

	s.logger.Debugf("websocket connected")

	// 记录所有连接，为了广播数据
	s.clients.Store(s.maxConnID, srv)
	s.maxConnID++
	s.metrics.connections.Inc()
	return srv
}

// Uint64Slice -
type Uint64Slice []uint64

// Len -
func (s Uint64Slice) Len() int { return len(s) }

// Less -
func (s Uint64Slice) Less(i, j int) bool { return s[i] < s[j] }

// Swap -
func (s Uint64Slice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetSensorids -
func (s *Server) GetSensorids() Uint64Slice {
	result := make(Uint64Slice, 0)
	s.clients.Range(func(key, value interface{}) bool {
		c := value.(*Client)
		if c.data == nil {
			return true
		}

		if c.isClosed {
			return true
		}

		c.data.Range(func(key, value interface{}) bool {
			result = append(result, key.(uint64))
			return true
		})

		return true
	})

	sort.Sort(result)

	i := 0
	var j int
	for {
		if i >= len(result)-1 {
			break
		}

		for j = i + 1; j < len(result) && result[i] == result[j]; j++ {
		}
		result = append(result[:i+1], result[j:]...)
		i++
	}

	return result
}
