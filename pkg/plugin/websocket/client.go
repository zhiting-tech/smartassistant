package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	websocket2 "github.com/zhiting-tech/smartassistant/modules/websocket"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"

	"github.com/gorilla/websocket"

	"github.com/zhiting-tech/smartassistant/modules/api/device"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

func NewClient(addr, token string) *Client {
	return &Client{
		Addr:      addr,
		Token:     token,
		Formatted: false,
	}
}

type Client struct {
	Conn      *websocket.Conn
	Addr      string
	Token     string // SA登录token
	Formatted bool   // 是否格式化输出
	ch        chan []byte
	requests  sync.Map
}

func (c *Client) listen() {
	c.ch = make(chan []byte, 10)
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			logger.Error(err.Error())
			time.Sleep(time.Second)
			continue
		}

		go func(msg []byte) {
			resp := websocket2.NewResponse(1)
			if err = json.Unmarshal(msg, &resp); err != nil {
				// logrus.Error(err)
				return
			}
			v, loaded := c.requests.Load(resp.ID)
			if loaded {
				v.(chan websocket2.Message) <- *resp
			}
		}(msg)
	}
}

// Request 请求消息
func (c *Client) Request(req websocket2.Request) (response websocket2.Message, err error) {
	req.ID = time.Now().UnixNano()
	msg, err := json.Marshal(req)
	if err != nil {
		return
	}

	timeout := time.NewTimer(time.Second * 10)
	ch := make(chan websocket2.Message)
	c.requests.Store(req.ID, ch)
	if err = c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return
	}

	for {
		select {
		case <-timeout.C:
			err = fmt.Errorf("request timeout: %s", string(msg))
			return
		case response = <-ch:
			c.requests.Delete(req.ID)
			return
		}
	}
}

// discover 发现设备
func (c *Client) discover(ctx context.Context) (devices []plugin.DiscoverResponse, err error) {

	req := websocket2.Request{
		ID:      time.Now().UnixNano(),
		Service: "discover",
	}
	msg, err := json.Marshal(req)
	if err != nil {
		return
	}

	if err = c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		return
	}

	resp := make(chan websocket2.Message, 256)
	c.requests.Store(req.ID, resp)
	for {
		select {
		case d := <-resp:
			if d.Data != nil {
				var exist bool
				var result websocket2.DiscoverResponse

				respJson, _ := json.Marshal(d.Data)
				json.Unmarshal(respJson, &result)

				if !exist {
					devices = append(devices, result.Device)
				}
			}
		case <-ctx.Done():
			c.requests.Delete(req.ID)
			close(resp)
			return
		}
	}
}

func (c *Client) Connect() (err error) {
	rowQuery := fmt.Sprintf("token=%s", c.Token)

	u := url.URL{Scheme: "ws", Host: c.Addr, Path: "/ws", RawQuery: rowQuery, ForceQuery: true}
	logger.Printf("connecting to %s", u.String())
	c.Conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return
	}

	go c.listen()
	return
}

// GetInstances 获取设备属性
func (c *Client) GetInstances(pluginID, identify string) (thingModel thingmodel.ThingModel, err error) {

	data, _ := json.Marshal(websocket2.DeviceHandleParams{IID: identify})
	req := websocket2.Request{
		ID:      1,
		Domain:  pluginID,
		Service: "get_instances",
		Data:    data,
	}
	var resp websocket2.Message
	if resp, err = c.Request(req); err != nil {
		return
	}
	d, _ := json.Marshal(resp.Data)
	json.Unmarshal(d, &thingModel)

	return
}

// SetAttributes 修改设备属性
func (c *Client) SetAttributes(pluginID, identify string, attr ...sdk.SetAttribute) (err error) {

	data, _ := json.Marshal(sdk.SetRequest{Attributes: attr})
	req := websocket2.Request{
		ID:      2,
		Domain:  pluginID,
		Service: "set_attributes",
		Data:    data,
	}
	logger.Println(string(data))
	if _, err = c.Request(req); err != nil {
		return
	}

	return
}

func (c *Client) Gateways(pluginID, model string) (err error) {
	gatewaysReq := struct {
		Model string `json:"model"`
	}{Model: model}
	data, _ := json.Marshal(gatewaysReq)

	req := websocket2.Request{
		ID:      2,
		Domain:  pluginID,
		Service: "list_gateways",
		Data:    data,
	}
	if _, err = c.Request(req); err != nil {
		return
	}

	return
}

func (c *Client) SubDevices(pluginID, identify string) (err error) {
	data, _ := json.Marshal(websocket2.DeviceHandleParams{IID: identify})

	req := websocket2.Request{
		ID:      2,
		Domain:  pluginID,
		Service: "sub_devices",
		Data:    data,
	}
	if _, err = c.Request(req); err != nil {
		return
	}

	return
}

func (c *Client) DeviceStates(pluginID, identify string) (err error) {
	data, _ := json.Marshal(websocket2.DeviceHandleParams{IID: identify})

	req := websocket2.Request{
		ID:      2,
		Domain:  pluginID,
		Service: "device_states",
		Data:    data,
	}
	if _, err = c.Request(req); err != nil {
		return
	}

	return
}

type subscribeReq struct {
	PluginID string `json:"plugin_id,omitempty"`
	IID      string `json:"iid,omitempty"`
}

func (c *Client) Subscribe(event string) (err error) {
	data, _ := json.Marshal(subscribeReq{ /*IID: "0x00000000157b4d9c666",*/ PluginID: "yeelight"})

	req := websocket2.Request{
		ID:      2,
		Service: "subscribe_event",
		Event:   event,
		Data:    data,
	}
	if _, err = c.Request(req); err != nil {
		return
	}

	return
}

// Discover 发现特定品牌设备
func (c *Client) Discover() (devices []plugin.DiscoverResponse, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if devices, err = c.discover(ctx); err != nil {
		return
	}

	return
}

// printMessage 打印消息
func (c *Client) printMessage(msg []byte) (err error) {
	var dst bytes.Buffer
	if c.Formatted {
		err = json.Indent(&dst, msg, "", "    ")
	} else {
		err = json.Compact(&dst, msg)
	}
	logger.Println("read:", dst.String())

	return
}

// addDevices 添加设备
func (c *Client) addDevices(devices ...plugin.DiscoverResponse) (err error) {

	for _, d := range devices {
		_, err = c.addDevice(d.PluginID, d.IID)
	}
	return
}

// addDevice 添加设备
func (c *Client) addDevice(pluginID string, iid string) (thingModel thingmodel.ThingModel, err error) {
	data, _ := json.Marshal(websocket2.DeviceHandleParams{IID: iid})
	req := websocket2.Request{
		ID:      1,
		Domain:  pluginID,
		Service: "connect",
		Data:    data,
	}
	if _, err = c.Request(req); err != nil {
		return
	}

	return
}

// deleteDevices 删除设备
func (c *Client) deleteDevices(devices ...plugin.DiscoverResponse) (err error) {
	for _, d := range devices {
		_, err = c.deleteDevice(d.PluginID, d.IID)
	}
	return
}

// deleteDevice 删除设备
func (c *Client) deleteDevice(pluginID string, iid string) (thingModel thingmodel.ThingModel, err error) {
	data, _ := json.Marshal(websocket2.DeviceHandleParams{IID: iid})
	req := websocket2.Request{
		ID:      1,
		Domain:  pluginID,
		Service: "disconnect",
		Data:    data,
	}
	if _, err = c.Request(req); err != nil {
		return
	}

	return
}

// getDevices 获取设备列表
func (c *Client) getDevices() (devices []device.Device, err error) {
	type BaseResponse struct {
		errors.Code
		Data struct {
			Devices []device.Device `json:"devices"`
		} `json:"data,omitempty"`
	}

	api := fmt.Sprintf("%s/api/devices", c.Addr)
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return
	}
	req.Header.Add("smart-assistant-token", c.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var baseResp BaseResponse
	if err = json.Unmarshal(data, &baseResp); err != nil {
		return
	}

	return baseResp.Data.Devices, nil
}
