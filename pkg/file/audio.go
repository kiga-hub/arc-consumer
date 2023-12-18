package file

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/kiga-hub/arc-consumer/pkg/wave"
	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/protocols"
)

// AudioWriter - 音频文件存储结构
type AudioWriter struct {
	mutex      sync.Mutex
	out        *wave.Writer
	path       string // audio file path
	id         uint64 // clientid -- subdir
	samplerate uint16 // current audio samplerate
	channel    int
	starttime  int64 // file data starttime
	lasttime   int64 // last write time ms
	timeoutMs  int64 // timeout writer new file
	threshold  int64 // file max duration
	logger     logging.ILogger
	dirent     string
	subdirent  string
	stype      byte
}

// write - 音频数据文件存储
// @param startTime int64 开始时间戳
// @param isSplit bool 是否包连续
// @param sa *protocols.SegmentAudio 音频数据结构
// @return int 音频数据存储长度
// @return error 错误信息
func (w *AudioWriter) write(startTime int64, isAlign bool, sa *protocols.SegmentArc) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 包数据对齐时间判断 , 如果数据时间不连续，需要写新文件
	if !isAlign {
		w.logger.Infow("audio file is interrupt flush", "filename", filepath.Base(w.path))
		if err := w.flush(); err != nil {
			return 0, err
		}
		if err := w.create(startTime); err != nil {
			return 0, err
		}
	}

	// 记录处理包时间，为之后超时判断
	w.lasttime = time.Now().UnixNano() / 1e3
	buff := sa.Data

	// 检测打开文件，如果没有创建
	if w.out == nil {
		if err := w.create(startTime); err != nil {
			return 0, err
		}
	}

	return w.out.Write(buff)
}

// flush - 刷新音频存储文件
// @return error 错误信息
func (w *AudioWriter) flush() error {
	if w.out == nil {
		return nil
	}

	// 需要将缓存刷新到磁盘，关闭文件，重命名
	w.out.Close()
	w.out = nil

	fileName := fmt.Sprintf("%s-%s.wav",
		fmt.Sprintf("%012X", w.id),
		time.Unix(w.starttime/1e6, 0).Format("20060102150405"),
	)

	switch w.stype {
	case protocols.STypeArc:
		fileName = fmt.Sprintf("%s-%s.wav", fmt.Sprintf("%012X", w.id), time.Unix(w.starttime/1e6, 0).Format("20060102150405"))
	}

	w.logger.Infow("create audio file", "filename", fileName)
	dirName, _ := filepath.Split(w.path)
	if err := os.Rename(w.path, path.Join(dirName, fileName)); err != nil {
		w.starttime = 0
		return err
	}

	w.starttime = 0
	return nil
}

// create - 根据时间戳创建音频存储文件
// @param starttime int64 开始时间戳
// @return error 错误信息
func (w *AudioWriter) create(starttime int64) error {
	// 获取时间字符串
	w.starttime = starttime
	curtime := time.Unix(w.starttime/1e6, 0).Format("20060102150405")

	// 拼接当前文件名
	dir := w.dirent
	switch w.subdirent {
	case "day":
		dir = path.Join(w.dirent, curtime[:8])
	case "month":
		dir = path.Join(w.dirent, curtime[:6])
	case "year":
		dir = path.Join(w.dirent, curtime[:4])
	}

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return fmt.Errorf("mkdir %s", dir)
			}
		}
	}
	w.path = path.Join(dir, fmt.Sprintf("%s-%s.current.wav",
		fmt.Sprintf("%012X", w.id),
		curtime,
	))

	// 打开文件
	out, err := wave.New(w.path, int(w.samplerate), 1)
	if err != nil {
		return fmt.Errorf("open %s %s", w.path, err.Error())
	}
	w.out = out

	return nil
}

// Flush - 超时刷新音频存储文件
// @param tum int64 超时文件刷新时间戳
// @return bool 是否刷新
// @return error 错误信息
func (w *AudioWriter) Flush(tum int64) (bool, error) {
	now := time.Now().UnixNano() / 1e3
	if now-w.lasttime < tum {
		return false, nil
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.logger.Infow("audio file timeout flush", "filename", filepath.Base(w.path), "timeout", tum)
	return true, w.flush()
}

// AWrite - 构造音频数据，文件存储
// @param id uint64 设备ID号
// @param timestamp int64 时间戳
// @param isSplit bool 是否包连续
// @param sa *protocols.SegmentAudioV2 音频数据
// @return int 存储音频数据长度
// @return error 错误信息
func (s *Server) AWrite(id uint64, timestamp int64, isAlign, isEnd bool, sa *protocols.SegmentArc) (int, error) {
	buff, _ := s.aBuffer.LoadOrStore(id,
		&AudioWriter{
			id:        id,
			timeoutMs: 1000,
			threshold: int64(s.config.DurationMin) * 60 * 1e6,
			logger:    s.logger,
			dirent:    path.Join(s.config.Dirent, "audio"),
			subdirent: s.config.SubDirent,
			stype:     protocols.STypeArc,
		},
	)
	abuff := buff.(*AudioWriter)
	n, err := abuff.write(timestamp, isAlign, sa)
	if err != nil {
		return n, err
	}
	if isEnd {
		s.logger.Infow("audio file is end flush", "filename", filepath.Base(abuff.path))
		if err := abuff.flush(); err != nil {
			return n, err
		}
		s.aBuffer.Delete(id)
	}
	return n, nil
}

// AV2Write - 构造音频数据，文件存储
func (s *Server) AV2Write(id uint64, timestamp int64, isAlign, isEnd bool, sa *protocols.SegmentArc) (int, error) {
	buff, _ := s.aBuffer.LoadOrStore(id,
		&AudioWriter{
			id:        id,
			timeoutMs: 1000,
			threshold: int64(s.config.DurationMin) * 60 * 1e6,
			logger:    s.logger,
			dirent:    path.Join(s.config.Dirent, "audio"),
			subdirent: s.config.SubDirent,
			stype:     protocols.STypeArc,
		},
	)
	abuff := buff.(*AudioWriter)

	n, err := abuff.write(timestamp, isAlign, &protocols.SegmentArc{
		Data: sa.Data,
	})
	if err != nil {
		return n, err
	}
	if isEnd {
		s.logger.Infow("audio file is end flush", "filename", filepath.Base(abuff.path))
		if err := abuff.flush(); err != nil {
			return n, err
		}
		s.aBuffer.Delete(id)
	}
	return n, nil
}
