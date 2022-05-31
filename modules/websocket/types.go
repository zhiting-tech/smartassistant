package websocket

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"

	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type ServiceType string

// websocket API命令
const (
	// serviceDiscover 发现设备
	serviceDiscover ServiceType = "discover"

	serviceSubscribeEvent ServiceType = "subscribe_event"

	// serviceGetInstances 获取设备物模型
	serviceGetInstances ServiceType = "get_instances"
	// serviceSetAttributes 设置设备属性
	serviceSetAttributes ServiceType = "set_attributes"
	// serviceConnect 连接（认证、配对）
	serviceConnect ServiceType = "connect"
	// serviceDisconnect 断开连接（取消配对）
	serviceDisconnect ServiceType = "disconnect"
	// serviceOTA 查看设备是否有固件更新
	serviceCheckUpdate ServiceType = "check_update"
	// serviceOTA 更新设备固件
	serviceOTA ServiceType = "ota"
	// serviceListGateways
	serviceListGateways ServiceType = "list_gateways"
	// serviceDeviceStates 设备的日志
	serviceDeviceStates ServiceType = "device_states"
	// serviceSubDevices 子设备列表
	serviceSubDevices ServiceType = "sub_devices"
)

type MsgType string

// 消息类型
const (
	MsgTypeResponse MsgType = "response"
	MsgTypeEvent    MsgType = "event"
)

type Request struct {
	ID      int64           `json:"id"`      // 请求ID，由客户端生成
	Domain  string          `json:"domain"`  // 不为空时等于插件唯一标识
	Service ServiceType     `json:"service"` // 对应的websocket命令
	Data    json.RawMessage `json:"data"`    // 具体的参数

	Event string `json:"event,omitempty"` // server为 subscribe_event 时可选

	ginCtx *gin.Context
	user   *session.User // 发起请求的用户信息
}

type CallFunc func(req Request) (interface{}, error)

var callFunctions = make(map[ServiceType]CallFunc)

func RegisterCallFunc(cmd ServiceType, callFunc CallFunc) {
	if _, ok := callFunctions[cmd]; ok {
		logger.Panic("call cmd already exist")
	}
	callFunctions[cmd] = callFunc
}

type Data interface {
}

type Response struct {
	Error   Error `json:"error,omitempty"`
	Success bool  `json:"success"`
}

type Error struct {
	Code    codes.Code `json:"code"`
	Message string     `json:"message"`
}

// Message 服务端响应的消息
type Message struct {
	ID int64 `json:"id"`
	*Response
	EventType string      `json:"event_type,omitempty"`
	Data      interface{} `json:"data"`
	Type      MsgType     `json:"type"`
}

func NewResponse(id int64) *Message {
	return &Message{
		ID:       id,
		Response: &Response{},
		Type:     MsgTypeResponse,
	}
}

func NewEvent(eventType string) *Message {
	return &Message{
		EventType: eventType,
		Type:      MsgTypeEvent,
	}
}
