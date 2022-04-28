package websocket

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
	"testing"
	"time"
)

const localSAURL = "http://0.0.0.0:37965"

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjg2NDAwLCJhcmVhX2lkIjoyMzA2NTQ3MTM4OTU1MjI3NiwiYWNjZXNzX2NyZWF0ZV9hdCI6MTY0MjQ5NTEwNywiY2xpZW50X2lkIjoiNTc2ZjRlNDctODQ4OS00NTk5LWFjNDgtOTU5NjgyNzNiNmIwIiwic2NvcGUiOiJ1c2VyLGFyZWEsZGV2aWNlIn0.61ZucZuL0fEjEftvrmU6PLTItGXiEiiI_OPWHlehaGk"
const pluginID = "zhiting"
const identity = "84f703a5ead5"

var client = NewClient(localSAURL, token)

func init() {
	if err := client.Connect(); err != nil {
		logrus.Println(err.Error())
	}
	time.Sleep(time.Second)
}

// TestGetAttributes 获取设备属性
func TestGetAttributes(t *testing.T) {

	tm, err := client.GetInstances(pluginID, identity)
	if err != nil {
		logger.Panic(err)
		return
	}
	data, _ := json.Marshal(tm)
	logrus.Println(string(data))
}

// TestSetAttributes 设置设备属性
func TestSetAttributes(t *testing.T) {
	power := sdk.SetAttribute{IID: "0x0000000012ed37c8", AID: 1, Val: "off"}
	brightness := sdk.SetAttribute{IID: "0x0000000017995bc5", AID: 2, Val: 10}

	err := client.SetAttributes(pluginID, identity, power, brightness)
	if err != nil {
		logger.Panic(err)
		return
	}
}

// TestDiscover 发现设备
func TestDiscover(t *testing.T) {
	devices, err := client.Discover()
	if err != nil {
		logger.Panic(err)
		return
	}
	for _, d := range devices {
		logrus.Println(d)
	}
}

// TestCheck 测试插件包基础功能
func TestCheck(t *testing.T) {
	// TODO 服务是否在etc注册成功

	// 发现设备
	logger.Println("发现设备...")
	ds, err := client.Discover()
	if err != nil {
		logger.Error(err)
		return
	}
	if len(ds) == 0 {
		logger.Warn("没有发现未添加过的设备")
	} else {
		// 添加设备
		logger.Println("添加设备...")
		if err = client.addDevices(ds...); err != nil {
			return
		}
	}

	// 获取设备列表
	logger.Println("获取设备列表...")
	devices, err := client.getDevices()
	if err != nil {
		return
	}
	for _, d := range devices {
		logger.Println(d.PluginID, d.IID)
	}

	// 获取并修改属性
	logger.Println("获取属性...")
	for _, d := range devices {
		tm, err := client.GetInstances(pluginID, d.IID)
		if err != nil {
			logger.Error(err.Error())
		}

		logger.Println("修改属性...")
		for _, ins := range tm.Instances {
			for _, srv := range ins.Services {
				if srv.Type == "info" {
					continue
				}
				for _, attr := range srv.Attributes {
					if attr.AID == 0 {
						continue
					}

					var val interface{}
					switch attr.Type {
					case thingmodel.OnOff.Type:
						if attr.Val == "on" {
							val = "off"
						} else {
							val = "on"
						}
					case thingmodel.Brightness.Type:
						if v, ok := attr.Val.(int); ok && v > 50 {
							val = 1
						} else {
							val = 100
						}
					}
					setAttr := sdk.SetAttribute{IID: ins.IID, AID: attr.AID, Val: val}
					err = client.SetAttributes(pluginID, d.IID, setAttr)
					if err != nil {
						logger.Error(err.Error())
					}
				}
			}
		}
	}

	// 删除设备
	logger.Println("删除设备...")
	if err = client.deleteDevices(ds...); err != nil {
		return
	}

	return
}
