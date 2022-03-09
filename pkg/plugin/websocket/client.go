package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"net"
	"time"
)

type Client struct {
	Conn
	Token     string // SA登录token
	PluginID  string // 插件id
	Formatted bool   // 是否格式化输出
}

type Attribute struct {
	InstanceID int         `json:"instance_id"`
	Attribute  string      `json:"attribute"`
	Val        interface{} `json:"val"`
}

type MsgGetAttribute struct {
	ID       int    `json:"id"`
	Domain   string `json:"domain"`
	Service  string `json:"service"`
	Identity string `json:"identity"`
}

type MsgSetAttribute struct {
	MsgGetAttribute
	ServiceData `json:"service_data"`
}

type ServiceData struct {
	Attributes []Attribute `json:"attributes"`
}

type MsgDiscover struct {
	Domain  string `json:"domain"`
	ID      int    `json:"id"`
	Service string `json:"service"`
}

// GetClient 获取客户端
func GetClient(c Client) (Client, error) {
	var err error
	addr := net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: 37965,
	}
	rowQuery := fmt.Sprintf("token=%s", c.Token)
	c.Conn, err = GetConn(addr.String(), rowQuery)

	return c, err
}

// GetAttributes 获取设备属性
func (c *Client) GetAttributes(identify string) (err error) {
	m := MsgGetAttribute{
		ID:       1,
		Domain:   c.PluginID,
		Service:  "get_attributes",
		Identity: identify,
	}
	msg, err := json.Marshal(m)
	if err != nil {
		return
	}

	defer c.Close()

	if err = c.Write(string(msg)); err != nil {
		return
	}
	resp, err := c.Read()
	if err != nil {
		return
	}
	err = c.printMessage(resp)

	return
}

// SetAttributes 修改设备属性
func (c *Client) SetAttributes(identify string, attr ...Attribute) (err error) {
	m := MsgSetAttribute{
		MsgGetAttribute: MsgGetAttribute{
			ID:       2,
			Domain:   c.PluginID,
			Service:  "set_attributes",
			Identity: identify,
		},
		ServiceData: ServiceData{Attributes: attr},
	}
	msg, err := json.Marshal(m)
	if err != nil {
		return
	}

	defer c.Close()

	if err = c.Write(string(msg)); err != nil {
		return
	}

	for i := 1; i < len(attr)*2; i++ {
		resp, err := c.Read()
		if err != nil {
			return err
		}
		if err = c.printMessage(resp); err != nil {
			return err
		}
		// 修改失败, 则结束等待
		var result map[string]interface{}
		if err = json.Unmarshal(resp, &result); err != nil {
			return err
		}
		if _, ok := result["success"]; ok {
			if !result["success"].(bool) {
				break
			}
		}
	}

	return
}

// Discover 发现特定品牌设备
func (c *Client) Discover() (err error) {
	type discoverResult struct {
		ID     int64  `json:"id"`
		Type   string `json:"type"`
		Result struct {
			Device plugin.DiscoverResponse `json:"device"`
		} `json:"result"`
		Error   string `json:"error,omitempty"`
		Success bool   `json:"success"`
	}

	msg := MsgDiscover{
		Domain:  "plugin",
		ID:      1634210518525,
		Service: "discover",
	}
	msgByte, err := json.Marshal(msg)
	if err != nil {
		return
	}
	if err = c.Write(string(msgByte)); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ch := make(chan error)
	go func() {
		for {
			resp, err := c.Read()
			if err != nil {
				ch <- err
			}
			var result discoverResult
			if err = json.Unmarshal(resp, &result); err != nil {
				ch <- err
			}
			err = c.printMessage(resp)
		}
	}()

	select {
	case <-ctx.Done():
	case <-ch:
	}

	return
}

// printMessage 打印消息
func (c *Client) printMessage(msg []byte) (err error) {
	var response bytes.Buffer
	if c.Formatted {
		err = json.Indent(&response, msg, "", "    ")
	} else {
		err = json.Compact(&response, msg)
	}
	logger.Debugf("read:", response.String())

	return
}
