package websocket

import (
	"log"
	"testing"
)

var client = Client{
	Token:     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjMxNTM2MDAwMCwiYXJlYV9pZCI6MSwiYWNjZXNzX2NyZWF0ZV9hdCI6MTYzNjUzODkxNiwiY2xpZW50X2lkIjoiMzE1ZDRjNWYtNjE0YS00ZjhmLWIyNTgtM2M3ZWU2MjcyMTZkIiwic2NvcGUiOiJhcmVhLGRldmljZSx1c2VyIn0.k38jZ9_9HqCtsE22FJbkNEGKG5Gtf0zi_c1esQwZKfE",
	PluginID:  "tplink-kasa",
	Formatted: false,
}

var identity = "800676CFB5349FDFC894257EA85BFF2F1D81C8A0"

// TestGetAttributes 获取设备属性
func TestGetAttributes(t *testing.T) {
	c, err := GetClient(client)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = c.GetAttributes(identity)
	if err != nil {
		log.Fatal(err)
		return
	}
}

// TestSetAttributes 设置设备属性
func TestSetAttributes(t *testing.T) {
	c, err := GetClient(client)
	if err != nil {
		log.Fatal(err)
		return
	}

	power := Attribute{InstanceID: 1, Attribute: "power", Val: "off"}
	brightness := Attribute{InstanceID: 1, Attribute: "brightness", Val: 10}

	err = c.SetAttributes(identity, power, brightness)
	if err != nil {
		log.Fatal(err)
		return
	}
}

// TestDiscover 发现设备
func TestDiscover(t *testing.T) {
	c, err := GetClient(client)
	err = c.Discover()
	if err != nil {
		log.Fatal(err)
		return
	}
}

// TestCheck 测试插件包基础功能
func TestCheck(t *testing.T) {
	err := Check(client)
	if err != nil {
		log.Fatal(err)
	}
}
