package ws

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/protocols"
)

const (
	// 允许等待的写入时间
	writeWait = 60 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 20 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	// maxMessageSize = 512
)

// Message - 客户端读写消息
type Message struct {
	// websocket.TextMessage 消息类型
	Type int
	Data []byte
}

// Client - websocket连接管理结构
type Client struct {
	ws             *websocket.Conn // 底层websocket
	outChan        chan *Message   // 写队列
	connectedAt    int64           // 连接连接时间戳（秒）
	clientids      []string        // 请求数据的采集设备ID
	interval       int64           // 数据采样间隔
	request        *ClientRequest  // 请求参数
	data           *sync.Map       // 数据缓存
	audioMaxSize   int             // 音频最大缓存
	vibrateMaxSize int             // 振动缓存
	fill           int             // 自动填充上一个值个数
	fillEnable     bool            // 全局配置
	fillStatus     bool
	logger         logging.ILogger
	isClosed       bool
	closeChan      chan struct{}
	mutex          sync.Mutex // 避免重复关闭管道,加锁处理
}

// ClientBuffer - 单个采集设备数据缓存
type ClientBuffer struct {
	CollectorID string
	ID          uint64
	Time        int64 // 当前数据时间
	Last        int64 //  上一个包处理的当前时间
	resp        ClientResponse
	Status      bool   // 数据状态
	fillCount   int    // 数据填补次数
	ASize       int    // 音频大小
	Audio       []byte // 音频缓存
	VSize       int    // 振动大小
	VibrateX    []byte
	VibrateY    []byte
	VibrateZ    []byte
	mutex       sync.Mutex // 避免重复关闭管道,加锁处理
}

// ClientRequest - 请求数据参数
type ClientRequest struct {
	Collectorids string `json:"collectorids"` // 采集设备ID列表，逗号分隔
	Interval     int64  `json:"interval"`     // 采样间隔，毫秒
	Temperature  bool   `json:"temperature"`  // 请求温度数据
	VibrateVatd  bool   `json:"vibrate_vatd"` // 请求振动强度波动数据
	Vibrate      bool   `json:"vibrate"`      // 请求原始振动数据
	AudioVatd    bool   `json:"audio_vatd"`   // 请求音频强度波动数据
	AudioSpark   bool   `json:"audio_spark"`  // 请求音频电火花波动数据
	Audio        bool   `json:"audio"`        // 请求原始音频数据
	Fill         bool   `json:"fill"`         // 请求没有数据补默认数据
}

// ClientResponse - 返回数据
type ClientResponse struct {
	AudioEngine *AudioEngineData `json:"audio_engine,omitempty"` // 音频经过算法后数据
	Audio       *AudioData       `json:"audio,omitempty"`        // 原始音频
}

// Done - 客户端连接结束，通知
func (c *Client) Done() <-chan struct{} {
	return c.closeChan
}

// Start - 启动维护连接收发数据
func (c *Client) Start(ctx context.Context) {
	go c.writeLoop()
	go c.readLoop()
}

// Send - 发送数据
func (c *Client) Send(cbuff *ClientBuffer, frame *protocols.Frame) {
	req := c.request
	if req == nil {
		return
	}

	// 获取数据包各个段数据
	var err error
	var sa *protocols.SegmentArc
	for _, stype := range frame.DataGroup.STypes {
		switch stype {
		case protocols.STypeArc:
			if sa, err = frame.DataGroup.GetArcSegment(); err != nil {
				c.logger.Error(err)
				return
			}
		}
	}

	cbuff.mutex.Lock()
	defer cbuff.mutex.Unlock()

	cbuff.Last = time.Now().UnixNano() / 1e3 // 处理包时间点（微秒）

	// 判断发送音频数据，音频数据每包都发送，不需要采样
	if req.Audio && sa != nil {
		cbuff.resp.Audio.Get(frame.Timestamp, sa)
		if err := c.sendAudioData(cbuff.resp.Audio); err != nil {
			c.logger.Errorw(err.Error())
		}
	}

	// 发送当前采样点数据
	if cbuff.Status && (req.Temperature || req.AudioSpark || req.AudioVatd || req.VibrateVatd) {
		c.sendBuffData(req, cbuff)
		cbuff.ASize = 0
		cbuff.VSize = 0
	}

}

// 读客户端发送过来的参数
func (c *Client) readLoop() {
	defer c.Close()
	for {
		msgType, data, err := c.ws.ReadMessage()
		if err != nil {
			websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
			return
		}
		if len(data) == 0 || msgType == websocket.BinaryMessage {
			c.logger.Errorf("websocket param error: %d:%v", msgType, data)
			continue
		}
		// 参数处理
		if err := c.MessageData(data); err != nil {
			c.logger.Errorf("websocket param %s: %v", err, string(data))
		}
	}
}

// 发送消息给客户端
func (c *Client) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	defer c.Close()
	for {
		select {
		// 取一个应答
		case msg := <-c.outChan:
			if err := c.ws.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
				c.logger.Error(err)
				return
			}
			// 写给websocket
			if err := c.ws.WriteMessage(msg.Type, msg.Data); err != nil {
				c.logger.Errorw(fmt.Sprintf("发送消息给客户端发生错误:%s", err.Error()))
				// 切断服务
				return
			}
		case <-c.closeChan:
			// 获取到关闭通知
			return
		case <-ticker.C:
			// 出现超时情况
			if err := c.ws.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Errorw(fmt.Sprintf("ping:%s", err.Error()))
				return
			}
		}
	}
}

// Write - 写入消息到队列中
func (c *Client) Write(Type int, data []byte) error {
	select {
	case c.outChan <- &Message{Type, data}:
	case <-c.closeChan:
		return errors.New("ws write - 连接已经关闭")
	}
	return nil
}

// 发送音频数据
func (c *Client) sendAudioData(data *AudioData) error {
	return c.sendData(&ClientResponse{Audio: data})
}

// 发送温度、震动、强度值数据
func (c *Client) sendBuffData(req *ClientRequest, buff *ClientBuffer) {
	err := c.sendData(&ClientResponse{
		AudioEngine: buff.resp.AudioEngine,
	})
	if err != nil {
		c.logger.Infow(err.Error(), "client", buff.CollectorID)
	}
}

func (c *Client) sendData(data *ClientResponse) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if data.Audio != nil {
		// 发送音频头数据
		if !data.Audio.IsSend {
			data.Audio.IsSend = true
			if err := c.Write(websocket.TextMessage, d); err != nil {
				return err
			}
		}
		// 发送websocket 二进制数据包
		if err := c.Write(websocket.BinaryMessage, data.Audio.Data[0].PCM); err != nil {
			return err
		}
	} else {
		// 发送websocket json数据包
		if err := c.Write(websocket.TextMessage, d); err != nil {
			return err
		}
	}
	return nil
}

func sliceRemoveDuplicates(slice []string) []string {
	sort.Strings(slice)
	i := 0
	var j int
	for {
		if i >= len(slice)-1 {
			break
		}

		for j = i + 1; j < len(slice) && slice[i] == slice[j]; j++ {
		}
		slice = append(slice[:i+1], slice[j:]...)
		i++
	}
	return slice
}

// MessageData - websocket client 参数数据解析
func (c *Client) MessageData(data []byte) error {
	var param ClientRequest
	if err := json.Unmarshal(data, &param); err != nil {
		return err
	}

	c.logger.Debugw("web socket request", "param", param)

	// 清空上次参数缓存
	c.data.Range(func(key, value interface{}) bool {
		c.data.Delete(key)
		return true
	})

	// 解析 collectorids, 分配缓存
	c.clientids = sliceRemoveDuplicates(strings.Split(strings.ToUpper(param.Collectorids), ","))

	// 每个传感器ID分配缓存
	for _, collid := range c.clientids {
		// 字符串ID转16进制
		var id uint64
		if collid != "" {
			storekeybyte, err := hex.DecodeString(collid)
			if err != nil || len(storekeybyte) != 6 {
				continue
			}

			for _, b := range storekeybyte {
				id <<= 8
				id += uint64(b)
			}
		}

		cbuff := &ClientBuffer{
			CollectorID: collid,
			ID:          id,
		}

		// 音频引擎算法结果缓存
		if param.AudioVatd || param.AudioSpark {
			cbuff.Audio = make([]byte, c.audioMaxSize)
			cbuff.resp.AudioEngine = &AudioEngineData{
				CollectorID: collid,
				Data:        make([]AudioEngine, 1),
			}
			if param.AudioVatd {
				cbuff.resp.AudioEngine.Data[0].NewVATD()
			}
			if param.AudioSpark {
				cbuff.resp.AudioEngine.Data[0].NewSpark()
			}
		}

		// 音频原始数据缓存
		if param.Audio {
			cbuff.resp.Audio = &AudioData{
				CollectorID: collid,
				Data:        make([]Audio, 1),
			}
		}

		c.data.Store(id, cbuff)
	}

	c.interval = param.Interval * 1000 // 毫秒转微秒
	c.request = &param

	return nil
}

// Close -
func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.isClosed {
		c.ws.Close()
		close(c.closeChan)
		c.isClosed = true
	}
}
