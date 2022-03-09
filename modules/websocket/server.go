package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	event2 "github.com/zhiting-tech/smartassistant/modules/event"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

// Server WebSocket服务端
type Server struct {
	bucket *bucket
}

func NewWebSocketServer() *Server {
	return &Server{
		bucket: newBucket(),
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) AcceptWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error(err)
		return
	}
	var (
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	user := session.Get(c)
	logger.Debugf("start websocket serve \"%s\" with \"%s\"", lAddr, rAddr)
	cli := &client{
		key:    uuid.New().String(),
		areaID: user.AreaID,
		conn:   conn,
		bucket: s.bucket,
		send:   make(chan []byte, 4),
		ginCtx: c,
	}

	s.bucket.register <- cli
	logger.Debug("new client Key：", cli.key)
	go cli.readWS(user)
	go cli.writeWS()
}

// SingleCast 发送单播消息
func (s *Server) SingleCast(cliID string, data []byte) error {
	cli := s.bucket.get(cliID)
	if cli == nil {
		return ErrClientNotFound
	}
	cli.send <- data
	return nil
}

func (s *Server) Broadcast(areaID uint64, data []byte) {
	s.bucket.broadcast <- broadcastData{
		AreaID: areaID,
		Data:   data,
	}
}

func (s *Server) Run(ctx context.Context) {
	logger.Info("starting websocket server")
	go s.bucket.run()
	<-ctx.Done()
	s.bucket.stop()
	logger.Warning("websocket server stopped")
}

// OnDeviceChange 设备状态,数量改变回调，会广播给所有客户端，并且触发场景
func (s *Server) OnDeviceChange(em event2.EventMessage) error {
	event := NewEvent(string(em.EventType))
	areaID := em.AreaID
	switch em.EventType {
	case event2.DeviceIncrease, event2.DeviceDecrease:
	case event2.AttributeChange:
		deviceID := em.GetDeviceID()
		d, err := entity.GetDeviceByID(deviceID)
		if err != nil {
			return err
		}
		attr := em.GetAttr()
		if attr == nil {
			logger.Warn("attr is nil")
			return nil
		}
		event.Data = map[string]interface{}{
			"identity":    d.Identity,
			"instance_id": attr.InstanceID,
			"attr":        attr.Attribute,
			"domain":      d.PluginID,
		}
	}
	data, _ := json.Marshal(event)
	s.Broadcast(areaID, data)
	return nil
}
