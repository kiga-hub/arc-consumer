package gnet

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/utils"
	"github.com/panjf2000/gnet/pkg/ringbuffer"
)

const (
	// AlignType - 对齐
	AlignType = 1
	// BrokenType - 断包
	BrokenType = 2
	// DropType - 丢包
	DropType = 3
	// CorrType - 矫正
	CorrType = 4
	// ResetType - 重发
	ResetType = 5
)

// FrameBuff - 缓存包结构信息
type FrameBuff struct {
	seq      int64
	realtime int64
	length   int
}

// Sensor - 传感器结构
type Sensor struct {
	// 传感器属性
	id      uint64 // 编号
	sid     string // 字符串编号
	addr    string // 地址
	holding bool

	// 时间对齐
	lastSequence    int64 // 上一个包序号
	aligntime       int64 // 上一个包理论时间点
	realtime        int64 // 上一个包实际时间点
	fillCount       int   // 补时间偏差计数
	offsetThreshold int   // 偏差最最大阈值

	// 日志
	logger logging.ILogger

	// 缓存
	buff      *ringbuffer.RingBuffer
	buffArray []FrameBuff
	redo      int
	reqtime   int64
	req       int64
	lasttime  int64

	// 采样计算
	sample      bool
	sampleCount int
}

// Nexttime - 获取时间戳
func (s *Sensor) Nexttime(id uint64, seq int64, fill bool, now, dataTime int64) (int64, int) {
	alignTime := s.aligntime + dataTime // 对齐时间(um)
	realTime := now - dataTime          // 实际时间(um)

	// 第一个包
	if s.aligntime == 0 {
		s.logger.Infow(fmt.Sprintf("first seq %d", seq), "sensor", s.sid)
		s.reset(seq, realTime)
		return s.aligntime, ResetType
	}

	// 包序号判断
	if seq != s.lastSequence+1 { // 包号不连续，重置时间
		s.logger.Infow(fmt.Sprintf("timestamp align seq %d != %d + 1", seq, s.lastSequence),
			"sensor", s.sid,
			"realtime", time.Unix(realTime/1e6, (realTime%1e6)*1e3).Format("2006-01-02 15:04:05.999999"),
			"aligntime", time.Unix(alignTime/1e6, (alignTime%1e6)*1e3).Format("2006-01-02 15:04:05.999999"),
		)
		var ctype = BrokenType
		if seq <= s.lastSequence {
			ctype = ResetType
		}
		s.reset(seq, realTime)
		return s.aligntime, ctype
	}

	// 如果阈值为不大于0，直接进行时间对齐
	if s.offsetThreshold <= 0 {
		s.inc(realTime, dataTime, fill)
		return s.aligntime, AlignType
	}

	// 对齐时间阈值判断
	if alignTime < realTime-int64(s.offsetThreshold) {
		// 包接收累计偏慢
		s.logger.Infow("offset max threshold reset time",
			"sensor", s.sid,
			"seq", seq,
			"realtime", time.Unix(realTime/1e6, (realTime%1e6)*1e3).Format("2006-01-02 15:04:05.999999"),
			"aligntime", time.Unix(alignTime/1e6, (alignTime%1e6)*1e3).Format("2006-01-02 15:04:05.999999"),
			"duration", dataTime,
			"threshold", s.offsetThreshold,
		)
		s.reset(seq, realTime)
		return s.aligntime, CorrType
	}
	if alignTime > realTime+int64(s.offsetThreshold) {
		// 包接收累计偏快
		s.lastSequence++
		s.logger.Infow("offset max threshold drop package",
			"client", s.sid,
			"seq", seq,
			"realtime", time.Unix(realTime/1e6, (realTime%1e6)*1e3).Format("2006-01-02 15:04:05.999999"),
			"aligntime", time.Unix(alignTime/1e6, (alignTime%1e6)*1e3).Format("2006-01-02 15:04:05.999999"),
			"duration", dataTime,
			"threshold", s.offsetThreshold,
		)
		return s.aligntime, DropType
	}

	// 时间对齐
	s.inc(realTime, dataTime, fill)
	return s.aligntime, AlignType
}

// 时间重置，用实际时间作为包时间
func (s *Sensor) reset(seq int64, realtime int64) {
	s.lastSequence = seq
	s.aligntime = realtime
	s.realtime = realtime
	s.fillCount = 0
}

// 时间对齐，用上一个包时间，加包数据时间，作为这个包时间
func (s *Sensor) inc(realtime, duration int64, fill bool) {
	s.lastSequence++
	s.aligntime += duration
	s.realtime = realtime

	// 补时间判断, 48k 采样数据，包数据不能整除，需要每3个包补一个时间
	if fill {
		s.fillCount++
		if s.fillCount >= 3 {
			s.fillCount = 0
			s.aligntime++
		}
	} else {
		s.fillCount = 0
	}
}

// 补包请求
func (s *Sensor) createFrame(seq uint64) []byte {
	buff := make([]byte, 19)
	buff[0] = 0xFC
	buff[1] = 0x01
	buff[2] = 0x0E
	buff[3] = byte((s.id >> 40) & 0xFF)
	buff[4] = byte((s.id >> 32) & 0xFF)
	buff[5] = byte((s.id >> 24) & 0xFF)
	buff[6] = byte((s.id >> 16) & 0xFF)
	buff[7] = byte((s.id >> 8) & 0xFF)
	buff[8] = byte(s.id & 0xFF)
	binary.BigEndian.PutUint64(buff[9:], seq)
	binary.BigEndian.PutUint16(buff[17:], utils.CheckSum(buff[3:17]))
	return buff
}
