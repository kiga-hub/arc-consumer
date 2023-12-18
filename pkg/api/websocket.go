package api

import (
	"bytes"
	"encoding/binary"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kiga-hub/arc/protocols"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有的CORS 跨域请求，正式环境可以关闭
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// websocket接口, inswarm配置为false时调用
// @param c echo.Context echo框架对象
// @return error 错误信息
func (s *Server) collectorWs(c echo.Context) error {
	// 应答客户端告知升级连接为websocket
	conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		return c.JSON(http.StatusBadRequest, nil)
	}

	// 注册连接
	s.ws.Register(conn)
	return c.JSON(http.StatusOK, nil)
}

// inswarm is true
// @param conn websocket.Conn websocket连接
// @param mt int
// @param msg []byte 消息数据
// @return error 错误信息
func (s *Server) handleWs(conn *websocket.Conn, mt int, msg []byte) error {
	// 注册连接
	client := s.ws.Register(conn)
	if mt == websocket.TextMessage {
		if err := client.MessageData(msg); err != nil {
			return err
		}

	}
	<-client.Done()
	return nil
}

func (s *Server) collectorWssensorids(c echo.Context) error {
	return c.JSON(http.StatusOK, s.ws.GetSensorids())
}

func (s *Server) collectorDataUpload(c echo.Context) error {
	// 应答客户端告知升级连接为websocket
	conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		return c.JSON(http.StatusBadRequest, nil)
	}

	defer conn.Close()
	frameBuff := protocols.NewDefaultFrame()
	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
			break
		}
		if len(data) <= protocols.DefaultHeadLength || msgType != websocket.BinaryMessage {
			s.logger.Errorf("websocket param error: %d:%v", msgType, data)
			continue
		}

		// check head
		if !bytes.Equal(data[0:4], protocols.Head[:]) {
			s.logger.Error("not find header")
			continue
		}

		// check version
		if data[4] != 1 {
			s.logger.Error("version=!1 ignore this package")
			continue
		}
		// check size
		fSize := binary.BigEndian.Uint32(data[5:9])
		if len(data) != protocols.DefaultHeadLength+int(fSize) {
			s.logger.Errorf("bad size [%d]", fSize)
			continue
		}
		// check packet end
		if data[len(data)-1] != protocols.End {
			s.logger.Errorf("bad end [%02X]", data[len(data)-1])
			continue
		}
		// 解包
		if err := frameBuff.Decode(data); err != nil {
			s.logger.Errorw(err.Error())
			continue
		}

		// 获取传感器ID
		var sensorid uint64
		for _, b := range frameBuff.ID {
			sensorid <<= 8
			sensorid += uint64(b)
		}

		// 将数据送到websocket
		if s.ws != nil {
			if err := s.ws.Write(sensorid, frameBuff); err != nil {
				s.logger.Errorw(err.Error())
			}
		}
	}

	return c.JSON(http.StatusOK, nil)
}
