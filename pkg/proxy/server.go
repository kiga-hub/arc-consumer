package proxy

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/kiga-hub/arc/logging"
)

// Handler -  代理接口
type Handler interface {
	Start(context.Context)
	Write(uint64, string, []byte) (int, error)
	Stop()
}

// Conn - 连接
type Conn struct {
	conn  net.Conn
	valid bool
}

// Server - 代理结构
type Server struct {
	pools     *sync.Map
	mask      uint64
	logger    logging.ILogger
	config    *Config
	metrics   *Metrics
	closeChan chan struct{}
	isClose   bool
}

// factory - 获取一个连接
// @return net.Conn 连接对象
// @return error 错误信息
func (s *Server) factory() (net.Conn, error) {
	return net.DialTimeout(
		s.config.NetType,
		s.config.Parent,
		time.Duration(time.Millisecond*time.Duration(s.config.Timeout)),
	)
}

// New - 创建代理服务
// @param opts Option 设置选项的函数，可变参数
// @return Handler 代理接口
func New(mask uint64, opts ...Option) (Handler, error) {
	srv := loadOptions(opts...)
	if !srv.config.Enable {
		return nil, nil
	}

	// 打印配置
	spew.Dump(srv.config)

	// 初始化
	srv.pools = new(sync.Map)
	srv.mask = mask
	srv.closeChan = make(chan struct{})

	// metrics统计
	srv.metrics = NewMetrics()
	return srv, nil
}

// Write - 代理转发数据
// @param data []byte 二进制数据
// @return int 发送数据长度
// @return error 错误信息
func (s *Server) Write(id uint64, sid string, data []byte) (int, error) {

	v, ok := s.pools.Load(id & s.mask)
	if !ok {
		// 第一次写，初始化连接
		conn := &Conn{}
		s.pools.Store(id&s.mask, conn)

		// 创建网络连接
		p, err := s.factory()
		if err != nil {
			return 0, err
		}
		conn.conn = p
		conn.valid = true
		s.metrics.connections.Inc()
		v = conn
		s.logger.Info("proxy conn ok", "mask", id&s.mask)
	}

	p := v.(*Conn)

	// 判断连接无效，丢弃包
	if !p.valid {
		return 0, nil
	}

	// 设置写超时
	if err := p.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * time.Duration(s.config.Timeout))); err != nil {
		p.valid = false
		s.metrics.connections.Dec()
		if err := p.conn.Close(); err != nil {
			s.logger.Error(err)
		}
		return 0, err
	}

	// 写数据
	n, err := p.conn.Write(data)
	if err != nil {
		p.valid = false
		s.metrics.connections.Dec()
		if err := p.conn.Close(); err != nil {
			s.logger.Error(err)
		}
		return n, err
	}
	s.metrics.packets.WithLabelValues(sid).Inc()
	return n, nil
}

// Stop - 关闭代理服务
func (s *Server) Stop() {
	if !s.isClose {
		close(s.closeChan)
		s.isClose = true
		s.pools.Range(func(key, value interface{}) bool {
			value.(*Conn).conn.Close()
			s.logger.Infow("proxy conn close", "mask", key)
			s.pools.Delete(key)
			return true
		})
		s.logger.Infow("proxy service stop")
	}
}

// Start - 开启代理服务
// @param ctx context 代理上下文信息
// @return error 错误信息
func (s *Server) Start(ctx context.Context) {

	s.logger.Infow("proxy service start")

	for {
		if s.isClose {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(s.config.Interval))
		s.pools.Range(func(key, value interface{}) bool {
			conn := value.(*Conn)
			if conn.valid {
				return true
			}

			p, err := s.factory()
			if err != nil {
				s.logger.Warnw("proxy reconn failed", "mask", key, "msg", err)
				return false
			}

			conn.conn = p
			conn.valid = true
			s.metrics.connections.Inc()
			s.logger.Infow("proxy reconn ok", "mask", key)
			return false
		})
	}
}
