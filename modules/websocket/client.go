package websocket

import (
	"context"
	"encoding/json"
	errors2 "errors"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"

	ws "github.com/gorilla/websocket"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	status2 "github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type client struct {
	key    string
	areaID uint64
	conn   *ws.Conn
	send   chan []byte
	bucket *bucket
	ginCtx *gin.Context

	subscribers []Subscriber
}

func (cli *client) SubscribeTopic(topic string, id int64) {

	fn := func(msg interface{}) error {
		logger.Debugf("func topic: %s", topic)
		m := msg.(*Message)
		m.ID = id
		cli.SendMsg(m)
		return nil
	}
	logger.Debugf("append subscriber")
	cli.subscribers = append(cli.subscribers, cli.bucket.Subscribe(topic, fn))
}
func (cli *client) SendMsg(msg *Message) {

	data, _ := json.Marshal(msg)
	select {
	case cli.send <- data:
	}
}

type ActionWrap struct {
	Cmd      string `json:"cmd"`
	Name     string `json:"name"`
	IsPermit bool   `json:"is_permit"`
}

type DeviceWrap struct {
	Cmd      string `json:"cmd"`
	Name     string `json:"name"`
	IsPermit bool   `json:"is_permit"`
}

// 解析 WebSocket 消息，并且调用业务逻辑
func (cli *client) handleWsMessage(data []byte, user *session.User) (err error) {
	var req Request
	if err = json.Unmarshal(data, &req); err != nil {
		return
	}
	req.ginCtx = cli.ginCtx
	req.User = user

	// 请参考 docs/guide/web-socket-api.md 中的定义
	// 订阅消息
	if req.Service == serviceSubscribeEvent {
		return cli.handleSubscribeEvent(req)
	}
	// 发现插件设备
	if req.Service == serviceDiscover { // 写死的发现命令，优先级最高，忽略 domain，发送给所有插件
		return cli.discover(req, user)
	}
	beginTime := time.Now().Format("2006-01-02 15:04:05 -0700 MST")
	resp := cli.handleCallService(req) // 通过插件服务和设备通信
	cli.SendMsg(resp)
	endTime := time.Now().Format("2006-01-02 15:04:05 -0700 MST")
	msg, _ := json.Marshal(resp)
	logger.Debugf("request: %s, response msg: %s, begin time: %s, end time: %s",
		string(data), string(msg), beginTime, endTime)
	return
}

type DiscoverResponse struct {
	Device plugin.DiscoverResponse `json:"device"`
}

func (cli *client) discover(req Request, user *session.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var discovered sync.Map
	ch := plugin.GetGlobalClient().DevicesDiscover(ctx)
	for r := range ch {

		y := plugin.Identify{
			PluginID: r.PluginID,
			IID:      r.IID,
			AreaID:   user.AreaID,
		}
		// 过滤相同的设备
		if _, loaded := discovered.LoadOrStore(y.ID(), struct{}{}); loaded {
			continue
		}
		_, err := entity.GetPluginDevice(user.AreaID, r.PluginID, r.IID)
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			r.LogoURL = plugin.DeviceLogoURL(cli.ginCtx.Request, r.PluginID, r.Model, r.Type)
			resp := NewResponse(req.ID)
			resp.Success = true
			resp.Data = DiscoverResponse{Device: r}
			cli.SendMsg(resp)
		}
	}
	return nil
}

type subscribeEventReq struct {
	IID      string `json:"iid"`
	PluginID string `json:"plugin_id"`
}

func (cli *client) handleSubscribeEvent(req Request) (err error) {
	if req.Event == "" {
		return errors.New(status2.WebsocketEventRequired)
	}
	logger.Debugf("handler subscribe event:%s", req.Event)

	// 解析请求消息
	var data subscribeEventReq
	if req.Data != nil {
		json.Unmarshal(req.Data, &data)
	}

	topic := fmt.Sprintf("%d/%s", req.User.AreaID, req.Event)
	if data.PluginID != "" {
		topic = fmt.Sprintf("%s/%s", topic, data.PluginID)
		if data.IID != "" {
			topic = fmt.Sprintf("%s/%s", topic, data.IID)
		}
	}
	topic = fmt.Sprintf("%s*", topic)

	logger.Debugf("subscribe topic: %s", topic)
	// 当前 cli 订阅该 topic
	cli.SubscribeTopic(topic, req.ID)

	resp := NewResponse(req.ID)
	resp.Success = true
	cli.SendMsg(resp)
	return
}

func (cli *client) handleCallService(req Request) (resp *Message) {

	var err error
	resp = NewResponse(req.ID)
	defer func() {
		if r := recover(); r != nil {
			err = errors2.New(fmt.Sprintf("handleCallService err: %v", r))
		}
		if err != nil {
			s := status.Convert(err)
			resp.Error.Code = s.Code()
			resp.Error.Message = s.Message()
		} else {
			resp.Success = true
		}
	}()
	if req.Service == "" {
		err = errors.New(status2.WebsocketDomainRequired)
		return
	}

	callFunc, ok := callFunctions[req.Service]
	if !ok {
		err = errors.New(status2.WebsocketCommandNotFound)
		return
	}
	resp.Data, err = callFunc(req)
	return
}

func (cli *client) Close() error {
	close(cli.send)
	for _, s := range cli.subscribers {
		s.Unsubscribe()
	}
	return cli.conn.Close()
}

func (cli *client) close() {
	cli.bucket.unregister <- cli
}

// readWS
func (cli *client) readWS(user *session.User) {
	defer cli.close()
	for {
		t, data, err := cli.conn.ReadMessage()
		if err != nil {
			return
		}
		if t == ws.CloseMessage {
			return
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error(r)
				}
			}()
			if err = cli.handleWsMessage(data, user); err != nil {
				logger.Errorf("handle websocket message error: %s, request: %s", err.Error(), string(data))
			}
		}()
	}
}

// writeWS
func (cli *client) writeWS() {
	defer cli.close()

	for {
		select {
		case msg, ok := <-cli.send:
			if !ok {
				_ = cli.conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}
			_ = cli.conn.WriteMessage(ws.TextMessage, msg)
		}
	}
}
