package httpproxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/kiga-hub/arc/httplib"
	"github.com/kiga-hub/arc/logging"
)

// Handler -  代理接口
type Handler interface {
	Start(context.Context)
	Write(uint64, []byte) (int, error)
	Stop()
}

// Buffer -
type Buffer struct {
	sid  string
	data []byte
}

// HTTPTransport -
var HTTPTransport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second, // 连接超时时间
		KeepAlive: 60 * time.Second, // 保持长连接的时间
	}).DialContext, // 设置连接的参数
	MaxIdleConns:          500,              // 最大空闲连接
	IdleConnTimeout:       60 * time.Second, // 空闲连接的超时时间
	ExpectContinueTimeout: 30 * time.Second, // 等待服务第一个响应的超时时间
	MaxIdleConnsPerHost:   100,              // 每个host保持的空闲连接数
}

// Server - 代理结构
type Server struct {
	client    http.Client
	conn      *websocket.Conn
	logger    logging.ILogger
	config    *Config
	sensorids *sync.Map
	metrics   *Metrics
	queue     chan *Buffer
	closeChan chan struct{}
}

// New - 创建代理服务
// @param opts Option 设置选项的函数，可变参数
// @return Handler 代理接口
func New(opts ...Option) (Handler, error) {
	srv := loadOptions(opts...)
	if !srv.config.Enable {
		return nil, nil
	}

	spew.Dump(srv.config)

	srv.closeChan = make(chan struct{})
	srv.queue = make(chan *Buffer, srv.config.Size)
	srv.client = http.Client{Transport: HTTPTransport} // 初始化一个带有transport的http的client
	srv.sensorids = new(sync.Map)

	srv.metrics = NewMetrics()
	return srv, nil
}

// Write - 代理转发数据
// @param data []byte 二进制数据
// @return int 发送数据长度
// @return error 错误信息
func (s *Server) Write(id uint64, data []byte) (int, error) {
	v, ok := s.sensorids.Load(id)
	if !ok {
		return 0, nil
	}

	s.queue <- &Buffer{
		sid:  v.(string),
		data: data,
	}
	s.metrics.queuePkts.Inc()

	return len(data), nil
}

// Stop - 关闭代理服务
func (s *Server) Stop() {
	close(s.closeChan)
}

// Start - 开启代理服务
// @param ctx context 代理上下文信息
// @return error 错误信息
func (s *Server) Start(ctx context.Context) {
	url := fmt.Sprintf("ws://%s/ws/upload", s.config.Parent)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	pingWait := 0
	cleanChannel := false

	s.logger.Infow("http proxy service start")

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.closeChan:
			return
		case <-ticker.C:
			// 获取传感器ID信息
			s.GetSensorids()

			// 取消清数据
			cleanChannel = false

			// 18s 没有数据 发送ping信息
			pingWait++
			if pingWait < 18 {
				continue
			}
			pingWait = 0
			if s.conn == nil {
				continue
			}
			// 出现ping信息
			if err := s.conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
				continue
			}
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				s.logger.Errorw(fmt.Sprintf("ping:%s", err.Error()))
				s.conn.Close()
				s.conn = nil
			}
		case b := <-s.queue:
			s.metrics.queuePkts.Dec() // 队列计数
			if cleanChannel {
				continue
			}
			// 发送包
			if err := s.DoBytesWebsocket(url, b.data); err != nil {
				s.logger.Error(err)
				cleanChannel = true
				continue
			}
			s.metrics.packets.WithLabelValues(b.sid).Inc()
			// 发送数据，不需要发送ping
			pingWait = 0
		}
	}
}

// ConnWebsocket -
func (s *Server) ConnWebsocket(url string) error {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	// 连接超时1s
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ws, _, err := websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		return err
	}
	s.conn = ws
	return nil
}

// DoBytesWebsocket -
func (s *Server) DoBytesWebsocket(url string, data []byte) error {
	if s.conn == nil {
		if err := s.ConnWebsocket(url); err != nil {
			return err
		}
	}

	// 发送超时1s
	if err := s.conn.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	}
	if err := s.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		s.conn.Close()
		s.conn = nil
		return err
	}

	return nil
}

// GetSensorids -
func (s *Server) GetSensorids() {
	url := fmt.Sprintf("http://%s/ws/sensorids", s.config.Parent)
	req := httplib.Get(url)
	var resp []uint64
	if err := req.ToJSON(&resp); err != nil {
		s.logger.Error(err)
	}
	proxy := false
	if len(resp) > 0 {
		proxy = true
	}
	s.sensorids.Range(func(key, value any) bool {
		idx := -1
		for i, id := range resp {
			if id == key.(uint64) {
				idx = i
				break
			}
		}
		if idx == -1 {
			s.sensorids.Delete(key)
		} else {
			resp = append(resp[:idx], resp[idx+1:]...)
		}
		return true
	})
	for _, id := range resp {
		if _, ok := s.sensorids.Load(id); !ok {
			s.sensorids.Store(id, fmt.Sprintf("%012X", id))
		}
	}
	if !proxy && s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}
