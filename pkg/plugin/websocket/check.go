package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/zhiting-tech/smartassistant/modules/api/device"
	response2 "github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Check 检查插件包基础功能
func Check(client Client) (err error) {
	// 获取连接
	log.Println("发起连接...")
	c, err := GetClient(client)
	if err != nil {
		return
	}
	defer c.Close()

	// TODO 服务是否在etc注册成功

	// 发现设备
	log.Println("发现设备...")
	devices, err := checkDiscover(c)
	if err != nil {
		return
	}

	// 添加设备
	log.Println("添加设备...")
	err = addDevices(c, devices)
	if err != nil {
		return
	}

	// 获取设备列表
	log.Println("获取设备列表...")
	identities, err := getDevices(c)
	if err != nil {
		return
	}

	// 获取属性
	log.Println("获取属性...")
	if err = checkGetAttribute(c, identities); err != nil {
		return
	}
	// 修改属性
	log.Println("修改属性...")
	if err = checkSetAttribute(c, identities); err != nil {
		return
	}

	return
}

// checkDiscover 检查插件发现功能
func checkDiscover(c Client) (devices []plugin.DiscoverResponse, err error) {
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

			if result.Success {
				if result.Result.Device.PluginID == c.PluginID {
					err = c.printMessage(resp)
					if err != nil {
						ch <- err
					}
					devices = append(devices, result.Result.Device)
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
	case <-ch:
	}

	return
}

// addDevices 添加设备
func addDevices(c Client, devices []plugin.DiscoverResponse) (err error) {
	type deviceAddReq struct {
		Device plugin.DiscoverResponse `json:"device"`
	}

	url := "http://0.0.0.0:9020/api/devices"

	for _, d := range devices {
		devicesByte, err := json.Marshal(deviceAddReq{Device: d})
		if err != nil {
			return err
		}
		body := bytes.NewReader(devicesByte)
		req, err := http.NewRequest("POST", url, body)
		if err != nil {
			return err
		}
		req.Header.Add("smart-assistant-token", c.Token)
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		resp, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		var baseResp response2.BaseResponse
		if err = json.Unmarshal(resp, &baseResp); err != nil {
			return err
		}
		if baseResp.Status == 0 {
			log.Println(d.PluginID, d.Identity, ":success")
		} else {
			log.Println(d.PluginID, d.Identity, ":fail", "err:", baseResp.Reason)
		}

		response.Body.Close()
	}

	return
}

// getDevices 获取设备列表
func getDevices(c Client) (identities []string, err error) {
	type BaseResponse struct {
		errors.Code
		Data struct {
			Devices []device.Device `json:"devices"`
		} `json:"data,omitempty"`
	}

	url := "http://0.0.0.0:9020/api/devices"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Add("smart-assistant-token", c.Token)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	resp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	var baseResp BaseResponse
	if err = json.Unmarshal(resp, &baseResp); err != nil {
		return
	}
	for _, v := range baseResp.Data.Devices {
		identities = append(identities, v.Identity)
		log.Println(v.PluginID, v.Identity)
	}

	return
}

// checkGetAttribute 测试获取设备属性
func checkGetAttribute(c Client, identities []string) (err error) {

	for _, identity := range identities {
		mGet := MsgGetAttribute{
			ID:       1,
			Domain:   c.PluginID,
			Service:  "get_attributes",
			Identity: identity,
		}
		msgGet, err := json.Marshal(mGet)
		if err != nil {
			return err
		}

		if err = c.Write(string(msgGet)); err != nil {
			return err
		}
		resp, err := c.Read()
		if err != nil {
			return err
		}

		var result map[string]interface{}
		if err = json.Unmarshal(resp, &result); err != nil {
			return err
		}
		if _, ok := result["success"]; ok {
			if !result["success"].(bool) {
				log.Println(c.PluginID, identity, ": fail, err:", result["error"].(string))
			} else {
				log.Println(c.PluginID, identity, ": success")
			}
		}
	}

	return
}

// checkSetAttribute 测试设置设备属性
func checkSetAttribute(c Client, identities []string) (err error) {
	for _, identity := range identities {
		var attr []Attribute
		power := Attribute{InstanceID: 1, Attribute: "power", Val: "on"}
		attr = append(attr, power)
		mSet := MsgSetAttribute{
			MsgGetAttribute: MsgGetAttribute{
				ID:       2,
				Domain:   c.PluginID,
				Service:  "get_attributes",
				Identity: identity,
			},
			ServiceData: ServiceData{Attributes: attr},
		}
		msgSet, err := json.Marshal(mSet)
		if err != nil {
			return err
		}

		if err = c.Write(string(msgSet)); err != nil {
			return err
		}

		for i := 1; i < len(attr)*2; i++ {
			resp, err := c.Read()
			if err != nil {
				return err
			}
			var result map[string]interface{}
			if err = json.Unmarshal(resp, &result); err != nil {
				return err
			}
			if _, ok := result["success"]; !ok {
				continue
			}
			if !result["success"].(bool) {
				log.Println(c.PluginID, identity, ": fail, err:", result["error"].(string))
			} else {
				log.Println(c.PluginID, identity, ": success")
			}
		}
	}

	return
}
