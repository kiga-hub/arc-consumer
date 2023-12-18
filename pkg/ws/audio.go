package ws

import (
	"github.com/kiga-hub/arc/protocols"
)

// Audio - 原始数据
type Audio struct {
	Time int64  `json:"time"`
	PCM  []byte `json:"-"`
}

// AudioData - 音频原始数据头结构
type AudioData struct {
	CollectorID string  `json:"collectorid"`
	Samplerate  int     `json:"samplerate"`
	Bits        int     `json:"bits"`
	Channel     int     `json:"channel"`
	IsSend      bool    `json:"-"`
	Data        []Audio `json:"data"`
}

// Get - 获取音频原始数据websocket数据结构
func (ad *AudioData) Get(timestamp int64, sa *protocols.SegmentArc) {
	ad.Data[0].Time = timestamp / 1e3
	ad.Data[0].PCM = sa.Data
}
