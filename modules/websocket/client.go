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
}

func (cli *client) Close() error {
	close(cli.send)
	return cli.conn.Close()
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

	logger.Debugf("domain:%s,service:%s,data:%s\n", req.Domain, req.Service, string(req.Data))

	// 请参考 docs/guide/web-socket-api.md 中的定义
	// 如果消息类型持续增多，请拆分
	if req.Service == serviceDiscover { // 写死的发现命令，优先级最高，忽略 domain，发送给所有插件
		return cli.discover(req, user)
	}

	req.user = *user
	return cli.handleCallService(req) // 通过插件服务和设备通信
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
		resp := NewResponse(req.ID)
		resp.Success = true
		_, err := entity.GetPluginDevice(user.AreaID, r.PluginID, r.IID)
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			r.LogoURL = plugin.DeviceLogoURL(cli.ginCtx.Request, r.PluginID, r.Model)
			resp.Data = DiscoverResponse{Device: r}
			msg, _ := json.Marshal(resp)
			cli.send <- msg
		}
	}
	return nil
}

func (cli *client) handleCallService(req Request) (err error) {

	resp := NewResponse(req.ID)
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
		msg, _ := json.Marshal(resp)
		cli.send <- msg
		logger.Debugf("req: %v, response msg: %s", req, string(msg))
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

// readWS
func (cli *client) readWS(user *session.User) {
	defer func() { cli.bucket.unregister <- cli }()

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
			if err := cli.handleWsMessage(data, user); err != nil {
				logger.Errorf("handle websocket message error: %s, request: %s", err.Error(), string(data))
			}
		}()
	}
}

// writeWS
func (cli *client) writeWS() {
	defer func() { cli.bucket.unregister <- cli }()

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
