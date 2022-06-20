package websocket

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/event"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
)

func NewWebSocketServer() *Server {
	return &Server{
		bucket: newBucket(),
	}
}

// Server WebSocket服务端
type Server struct {
	bucket *bucket
}

func (s *Server) AcceptWebSocket(c *gin.Context) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

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

func (s *Server) Run(ctx context.Context) {
	logger.Info("starting websocket server")
	go s.bucket.run()
	<-ctx.Done()
	s.bucket.stop()
	logger.Warning("websocket server stopped")
}

type AttrChangeEvent struct {
	PluginID string                 `json:"plugin_id"`
	Attr     definer.AttributeEvent `json:"attr"`
}

type DeviceIncreaseEvent struct {
	Device entity.Device `json:"device"`
}

// MulticastMsg 设备状态、数量改变，会多播给所有订阅了主题的客户端，并且触发场景
func (s *Server) MulticastMsg(em event.EventMessage) error {
	ev := NewEvent(string(em.EventType))
	areaID := em.AreaID
	topic := fmt.Sprintf("%d/%s", areaID, em.EventType)
	switch em.EventType {
	case event.OnlineStatus, event.ThingModelChange:
		ev.Data = em.Param
		pluginID := em.Param["plugin_id"]
		iid := em.Param["iid"]
		topic = fmt.Sprintf("%d/%s/%s/%s", areaID, em.EventType, pluginID, iid)
	case event.DeviceDecrease:
	case event.DeviceIncrease:
		ev.Data = em.Param
		v, ok := em.Param["device"]
		if ok {
			d, ok := v.(entity.Device)
			if ok && d.PluginID != "" {
				topic = fmt.Sprintf("%d/%s/%s", areaID, em.EventType, d.PluginID)
			}
		}
	case event.AttributeChange:
		deviceID := em.GetDeviceID()
		d, err := entity.GetDeviceByID(deviceID)
		if err != nil {
			logger.Error(err)
			return err
		}
		attr := em.GetAttr()
		if attr == nil {
			logger.Warn("attr is nil")
			return nil
		}

		ev.Data = AttrChangeEvent{
			PluginID: d.PluginID,
			Attr:     *attr,
		}
		topic = fmt.Sprintf("%d/%s/%s/%s", areaID, em.EventType, d.PluginID, d.IID)
	}

	logger.Debugf("multicast topic: %s msg %v", topic, em)

	s.bucket.Publish(topic, ev)
	return nil
}
