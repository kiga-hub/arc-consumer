package grpc

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/kiga-hub/arc/logging"
	proto "github.com/kiga-hub/arc/protobuf/pb"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	// DialTimeout the timeout of create connection
	DialTimeout = 5 * time.Second

	// BackoffMaxDelay provided maximum delay when backing off after failed connection attempts.
	BackoffMaxDelay = 3 * time.Second

	// KeepAliveTime is the duration of time after which if the client doesn't see
	// any activity it pings the server to see if the transport is still alive.
	KeepAliveTime = time.Second // time.Duration(10) *

	// KeepAliveTimeout is the duration of time for which the client waits after having
	// pinged for keepalive check and if no activity is seen even after that the connection
	// is closed.
	KeepAliveTimeout = time.Second // time.Duration(3) *

	// InitialWindowSize we set it 1GB is to provide system's throughput.
	InitialWindowSize = 1 << 30

	// InitialConnWindowSize we set it 1GB is to provide system's throughput.
	InitialConnWindowSize = 1 << 30

	// MaxSendMsgSize set max gRPC request message size sent to server.
	// If any request message size is larger than current value, an error will be reported from gRPC.
	MaxSendMsgSize = 4 << 30

	// MaxRecvMsgSize set max gRPC arc-consumer message size arc-consumer from server.
	// If any message size is larger than current value, an error will be reported from gRPC.
	MaxRecvMsgSize = 4 << 30
)

// Handler - grpc接口定义
type Handler interface {
	Start(ctx context.Context)
	Write(uint64, string, []byte) error
	Stop()
	SetMask(uint64)
	ReConnect()
	Disconnect()
}

// Conn - 连接
type Conn struct {
	conn       *grpc.ClientConn
	grpcclient proto.FrameDataClient
	grpcstream proto.FrameData_FrameDataCallbackClient
	valid      bool
	reconn     bool
}

// Server -
type Server struct {
	pools     *sync.Map
	mask      uint64
	logger    logging.ILogger
	config    *Config
	running   *atomic.Bool
	closeChan chan struct{}
}

// New - 初始化grpc服务
// @param opts Option 设置选项的函数，可变参数
// @return Handler grpc处理器结构
func New(opts ...Option) Handler {
	srv := loadOptions(opts...)
	// 判断没有开启grpc服务直接返回
	if !srv.config.Enable {
		return nil
	}

	srv.pools = new(sync.Map)
	srv.running = atomic.NewBool(false)
	srv.closeChan = make(chan struct{})

	spew.Dump(srv.config)

	return srv
}

// ReConnect -
func (s *Server) ReConnect() {
	s.config = GetConfig()
	s.pools.Range(func(key, value interface{}) bool {
		value.(*Conn).reconn = true
		return true
	})
	s.running.Store(true)
}

// Disconnect -
func (s *Server) Disconnect() {
	s.running.Store(false)
}

// SetMask -
func (s *Server) SetMask(mask uint64) {
	s.mask = mask
}

// Start target server of grpc
// @return err 错误信息
func (s *Server) Start(ctx context.Context) {
	s.logger.Infow("grpc service start")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.closeChan:
			return
		case <-ticker.C:
			if !s.running.Load() {
				continue
			}
			var connections int
			s.pools.Range(func(key, value interface{}) bool {
				conn := value.(*Conn)
				if conn.valid {
					connections++
					return true
				}
				if conn.reconn {
					return true
				}

				grpcstream, err := conn.grpcclient.FrameDataCallback(context.Background())
				if err != nil {
					return false
				}

				conn.grpcstream = grpcstream
				conn.valid = true
				connections++
				s.logger.Infow("grpc replay", "mask", key, "addr", s.config.Server)
				return true
			})
		}
	}
}

// Dial return a grpc connection with defined configurations.
// @return grpc.ClientConn grpc客户端连接
// @return err 错误信息
// Dial return a grpc connection with defined configurations.
func (s *Server) factory() (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DialTimeout)
	defer cancel()
	return grpc.DialContext(ctx, s.config.Server,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// grpc.WithBackoffMaxDelay(BackoffMaxDelay),
		grpc.WithInitialWindowSize(InitialWindowSize),
		grpc.WithInitialConnWindowSize(InitialConnWindowSize),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(MaxSendMsgSize)),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxRecvMsgSize)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                KeepAliveTime,
			Timeout:             KeepAliveTimeout,
			PermitWithoutStream: true,
		}))
}

// Write - 调用grpc服务发送数据
// @param id uint64 偏移时间段
// @param data []byte 二进制数据包
// @return err 错误信息
func (s *Server) Write(id uint64, sid string, value []byte) (err error) {
	if !s.running.Load() {
		return nil
	}

	v, ok := s.pools.Load(id & s.mask)
	if !ok {
		conn := &Conn{}
		s.pools.Store(id&s.mask, conn)

		c, err := s.factory()
		if err != nil {
			return err
		}
		conn.conn = c
		conn.grpcclient = proto.NewFrameDataClient(c)
		grpcstream, err := conn.grpcclient.FrameDataCallback(context.Background())
		if err != nil {
			return err
		}
		conn.grpcstream = grpcstream
		conn.valid = true
		s.logger.Infow("grpc connnect", "mask", id&s.mask, "addr", s.config.Server)
		v = conn
	}

	p := v.(*Conn)

	// 需要重连
	if p.reconn {
		if p.grpcstream != nil {
			_, err = p.grpcstream.CloseAndRecv()
			if err != nil {
				s.logger.Errorf("CloseAndRecv() error: %v", err)
			}
		}
		p.conn.Close()
		s.pools.Delete(id & s.mask)
		s.logger.Infow("grpc reconnnect", "mask", id&s.mask, "addr", s.config.Server)
		return nil
	}

	// 连接不可用判断
	if !p.valid || p.grpcstream == nil {
		return nil
	}

	// 准备数据
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, id)

	request := proto.FrameDataRequest{
		Key:   key[2:],
		Value: value,
	}

	// fmt.Printf("\tgrpc sending %d, goroutine %d\n", len(value), runtime.NumGoroutine())
	// KeepAliveTime = 10s 在40s时，会发送失败，现在设置60s，在240s，发送失败
	if err := p.grpcstream.Send(&request); err != nil {
		if err != io.EOF {
			s.logger.Infow("send error", "err", err, "mask", id&s.mask)
		}
		// 注意： 这里不能用 CloseSend , 不然每次新建连接时，goroutine会不断增加，每次增加1个
		if _, cerr := p.grpcstream.CloseAndRecv(); cerr != nil {
			if err != io.EOF {
				s.logger.Infow("CloseAndRecv error", "err", err, "mask", id&s.mask)
			}
		}
		p.grpcstream, err = p.grpcclient.FrameDataCallback(context.Background())
		if err != nil {
			p.valid = false
			return fmt.Errorf("frameDataCallback %v", err)
		}
		if err := p.grpcstream.Send(&request); err != nil {
			p.valid = false
			return fmt.Errorf("send %v", err)
		}
	}

	return nil
}

// Stop - 停止grpc服务
func (s *Server) Stop() {
	if !s.running.Load() {
		return
	}
	s.running.Store(false)
	s.pools.Range(func(key, value interface{}) bool {
		p := value.(*Conn)
		if p.grpcstream != nil {
			resp, err := p.grpcstream.CloseAndRecv()
			if err != nil {
				s.logger.Infow("Stop Connected", "error", err)
			}
			if !resp.Successed {
				s.logger.Infow("gRPC Connected Fail", "success", resp.Successed)
			}
		}
		s.pools.Delete(key)
		return true
	})
	s.logger.Infow("grpc service stop")
}
