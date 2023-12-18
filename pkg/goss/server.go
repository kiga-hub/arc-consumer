package goss

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kiga-hub/arc/logging"
	microComponent "github.com/kiga-hub/arc/micro/component"
)

// Handler -
type Handler interface {
	Start(context.Context)
	Sync(uint64) error
}

// KeyCache -
type KeyCache struct {
	sensorID string
	last     int64
	lastPkg  int64
}

// Server -
type Server struct {
	cache         *sync.Map
	maxTime       int64
	gossipKVCache *microComponent.GossipKVCacheComponent
	logger        logging.ILogger
}

// New -
func New(opts ...Option) Handler {
	srv := loadOptions(opts...)

	if srv.gossipKVCache == nil {
		return nil
	}

	srv.cache = new(sync.Map)
	srv.maxTime = 60

	return srv
}

// Start -
func (s *Server) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second * time.Duration(s.maxTime))
	defer ticker.Stop()

	s.logger.Infow("goss kvcache start")

	for {
		select {
		case <-ticker.C:
			cur := time.Now().Unix()
			sensors := []string{}
			s.cache.Range(func(key, value interface{}) bool {
				c := value.(*KeyCache)
				if cur-atomic.LoadInt64(&c.lastPkg) > 5 {
					s.cache.Delete(key)
					return true
				}
				if cur-c.last > s.maxTime {
					sensors = append(sensors, c.sensorID)
					c.last = cur
				}
				return true
			})
			if len(sensors) > 0 {
				err := s.gossipKVCache.HaveSensorIDs(sensors)
				if err != nil {
					s.logger.Error(err)
				}
			}
		case <-ctx.Done():
			s.logger.Infow("goss kvcache stop")
			return
		}
	}
}

// Sync -
func (s *Server) Sync(sensorid uint64) error {
	if v, ok := s.cache.Load(sensorid); ok {
		kc := v.(*KeyCache)
		atomic.StoreInt64(&kc.lastPkg, time.Now().Unix())
		return nil
	}
	c := &KeyCache{
		sensorID: fmt.Sprintf("%012X", sensorid),
		last:     time.Now().Unix(),
		lastPkg:  time.Now().Unix(),
	}
	s.logger.Debugw("gossipKVCache", "time", time.Unix(c.last, 0).Format("2006-01-02 15:04:05"))
	if err := s.gossipKVCache.HaveSensorIDs([]string{c.sensorID}); err != nil {
		c.last = 0
		return err
	}
	s.cache.Store(sensorid, c)
	return nil
}
