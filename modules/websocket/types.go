package websocket

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"google.golang.org/grpc/codes"
)

type MsgType string
type EventType string

// websocket API命令
const (
	// serviceDiscover 发现设备
	serviceDiscover = "discover"

	// serviceGetAttributes 获取设备所有属性
	serviceGetAttributes = "get_attributes"
	// serviceSetAttributes 设置设备属性
	serviceSetAttributes = "set_attributes"
	// serviceUpdateThingModel
	serviceUpdateThingModel = "update_thing_model"
	// serviceConnect 连接（认证、配对）
	serviceConnect = "connect"
	// serviceDisconnect 断开连接（取消配对）
	serviceDisconnect = "disconnect"
	// serviceOTA 查看设备是否有固件更新
	serviceCheckUpdate = "check_update"
	// serviceOTA 更新设备固件
	serviceOTA = "ota"
)

// 消息类型
const (
	MsgTypeResponse MsgType = "response"
	MsgTypeEvent    MsgType = "event"
)

// callService websocket命令结构体
type callService struct {
	Domain      string          // 不为空时等于插件唯一标识
	ID          int             // 请求ID，由客户端生成
	Service     string          // 对应的websocket命令
	ServiceData json.RawMessage `json:"service_data"` // 具体的参数
	Identity    string          // 插件中设备唯一标识

	callUser session.User // 发起请求的用户信息
}

type Result map[string]interface{}

type CallFunc func(service callService) (Result, error)

var callFunctions = make(map[string]CallFunc)

func RegisterCallFunc(cmd string, callFunc CallFunc) {
	if _, ok := callFunctions[cmd]; ok {
		logrus.Panic("call cmd already exist")
	}
	callFunctions[cmd] = callFunc
}

type event struct {
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
}

type callResponse struct {
	ID      int                    `json:"id"`
	Error   Error                  `json:"error,omitempty"`
	Result  map[string]interface{} `json:"result"`
	Success bool                   `json:"success"`
}

type Error struct {
	Code    codes.Code `json:"code"`
	Message string     `json:"message"`
}

func (cr *callResponse) AddResult(key string, value interface{}) {
	if cr.Result == nil {
		cr.Result = make(map[string]interface{})
	}
	cr.Result[key] = value
}

type Message struct {
	*callResponse
	*event
	Type MsgType `json:"type"`
}

func NewResponse(id int) *Message {
	return &Message{
		callResponse: &callResponse{
			ID: id,
		},
		Type: MsgTypeResponse,
	}
}

func NewEvent(eventType string) *Message {
	return &Message{
		event: &event{
			EventType: eventType,
			Data:      make(map[string]interface{}),
		},
		Type: MsgTypeEvent,
	}
}
