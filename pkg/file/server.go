package file

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/protocols"
)

// Handler -
type Handler interface {
	Start(context.Context)
	Write(uint64, *protocols.Frame) error
	Stop()
}

// Server - 文件模块管理结构
type Server struct {
	aBuffer *sync.Map
	config  *Config
	logger  logging.ILogger
}

// New - 创建
// @param opts Option 设置选项的函数，可变参数
// @return Handler 文件存储对象
// @return error 错误信息
func New(opts ...Option) (Handler, error) {
	srv := loadOptions(opts...)

	if !srv.config.AEnable && !srv.config.VEnable && !srv.config.TEnable && !srv.config.MVEnable {
		return nil, nil
	}

	spew.Dump(srv.config)

	if _, err := os.Stat(srv.config.Dirent); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(srv.config.Dirent, os.ModePerm); err != nil {
				return nil, fmt.Errorf("mkdir %s", srv.config.Dirent)
			}
		}
	}

	if srv.config.AEnable {
		srv.aBuffer = new(sync.Map)
	}

	return srv, nil
}

// Start - 启动
// @param ctx Context 文件存储上下文信息
func (s *Server) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(s.config.Timeout))
	defer func() {
		ticker.Stop()
	}()

	s.logger.Infow("file service start", "dirent", s.config.Dirent)

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			s.logger.Infow("file service stop")
			return
		}

		// 定时检查接收数据超时刷新缓存数据存储
		if s.aBuffer != nil {
			s.aBuffer.Range(func(key, value interface{}) bool {
				b := value.(*AudioWriter)
				if ok, err := b.Flush(int64(s.config.Timeout) * 1000); err != nil {
					s.logger.Errorw(err.Error(), "client", fmt.Sprintf("%012X", key.(uint64)))
				} else if ok {
					s.aBuffer.Delete(key)
				}
				return true
			})
		}
	}
}

// Stop - 关闭
func (s *Server) Stop() {
	if s.aBuffer != nil {
		s.aBuffer.Range(func(key, value interface{}) bool {
			b := value.(*AudioWriter)
			if _, err := b.Flush(0); err != nil {
				s.logger.Errorw(err.Error(), "client", fmt.Sprintf("%012X", key.(uint64)))
			}
			s.aBuffer.Delete(key)
			return true
		})
	}
	s.logger.Infow("file service close")
}

// Write - 写入文件存储
// @param id uint64 设备ID号
// @param frame *protocols.Frame 数据包对象
// @return error 错误信息
func (s *Server) Write(id uint64, frame *protocols.Frame) error {
	for _, stype := range frame.DataGroup.STypes {
		switch stype {
		case protocols.STypeArc:
			if s.aBuffer == nil {
				continue
			}
			sa, err := frame.DataGroup.GetArcSegment()
			if err != nil {
				return err
			}
			// 音频数据写文件
			if _, err := s.AWrite(id, frame.Timestamp, true, true, sa); err != nil {
				return err
			}
		}
	}
	return nil
}
