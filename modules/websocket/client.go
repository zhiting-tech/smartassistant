package websocket

import (
	"context"
	"encoding/json"
	errors2 "errors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gorm.io/gorm"
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
	var cs callService
	if err = json.Unmarshal(data, &cs); err != nil {
		return
	}

	logger.Debugf("domain:%s,service:%s,data:%s\n", cs.Domain, cs.Service, string(cs.ServiceData))

	// 请参考 docs/guide/web-socket-api.md 中的定义
	// 如果消息类型持续增多，请拆分
	if cs.Service == serviceDiscover { // 写死的发现命令，优先级最高，忽略 domain，发送给所有插件
		return cli.discover(cs, user)
	}

	cs.callUser = *user
	return cli.handleCallService(cs) // 通过插件服务和设备通信
}

func (cli *client) discover(cs callService, user *session.User) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ch := plugin.GetGlobalClient().DevicesDiscover(ctx)
	for result := range ch {
		resp := NewResponse(cs.ID)
		resp.Success = true
		_, err = entity.GetPluginDevice(user.AreaID, result.PluginID, result.Identity)
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			d := entity.Device{
				Identity:     result.Identity,
				Model:        result.Model,
				Manufacturer: result.Manufacturer,
				PluginID:     result.PluginID,
			}
			result.LogoURL = plugin.DeviceLogoURL(cli.ginCtx.Request, d)
			resp.AddResult("device", result)
			msg, _ := json.Marshal(resp)
			cli.send <- msg
		}
	}
	return

}

func (cli *client) handleCallService(cs callService) (err error) {

	resp := NewResponse(cs.ID)
	defer func() {
		if err != nil {
			s := status.Convert(err)
			resp.Error.Code = s.Code()
			resp.Error.Message = s.Message()
		} else {
			resp.Success = true
		}
		msg, _ := json.Marshal(resp)
		cli.send <- msg
		logger.Debugf("cs: %v, response msg: %s", cs, string(msg))
	}()

	callFunc, ok := callFunctions[cs.Service]
	if !ok {
		return
	}
	resp.Result, err = callFunc(cs)
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
