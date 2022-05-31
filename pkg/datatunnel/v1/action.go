package datatunnel

import (
	"encoding/json"
	"errors"
)

const (
	AuthAction   = "Auth"   // 认证,由SA发起
	CreateAction = "Create" // 认证通过, 创建数据通道客户端,由SC发起
	NewAction    = "New"    // 新建一个TCP连接,由SC发起
	ErrorAction  = "Error"  // 错误事件
)

// CreateActionData 新建时间数据格式
type CreateActionData struct {
	Port        int    `json:"port"`
	ServiceName string `json:"service_name"`
}

// AuthActionData 认证事件数据格式
type AuthActionData struct {
	SAID        string `json:"sa_id"`
	Key         string `json:"key"`
	ServiceName string `json:"service_name"`
}

// NewActionData 新建TCP连接数据格式
type NewActionData struct {
	ServiceName   string `json:"service_name"`
	ConnectionKey string `json:"key"`
}

// 获取事件数据
func GetActionData(value string, data interface{}) (err error) {
	if data == nil {
		err = errors.New("pointer is nil")
		return
	}

	err = json.Unmarshal([]byte(value), data)
	return
}
