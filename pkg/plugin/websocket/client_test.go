package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/zhiting-tech/smartassistant/pkg/event"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

const localSAURL = "0.0.0.0:37965"

const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjozNTUxLCJleHAiOjMxNTM2MDAwMCwiYXJlYV9pZCI6NzY1NjAwMTM3NDA1MDQ0NjAsImFjY2Vzc19jcmVhdGVfYXQiOjE2NDg2MDQ5OTksImNsaWVudF9pZCI6Ijg5MGMzYWVmLWU5M2MtNDc5MS05Y2JjLWQ3NzM1OWMyODJjZSIsInNjb3BlIjoiYWxsIn0.87RRqVH3vibeJYAColDmlzg8zSzJjsef2EEqLhhsKqc"
const pluginID = "meihuizhiju.meihuizhijuchajianbao"
const identity = "84f703a5ead5"

var client = NewClient(localSAURL, token)

func init() {
	if err := client.Connect(); err != nil {
		logrus.Println(err.Error())
	}
	time.Sleep(time.Second)
	logrus.SetLevel(logrus.DebugLevel)
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
	power := sdk.SetAttribute{IID: identity, AID: 9, Val: "on"}
	// brightness := sdk.SetAttribute{IID: "0x0000000017995bc5", AID: 2, Val: 10}

	err := client.SetAttributes(pluginID, identity, power /*, brightness*/)
	if err != nil {
		logger.Panic(err)
		return
	}
	time.Sleep(time.Second * 10)
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

func TestListen(t *testing.T) {
	client.Subscribe(string(event.OnlineStatus))
	time.Sleep(time.Second * 50)
}

func TestConnect(t *testing.T) {
	tm, err := client.addDevice(pluginID, identity)
	if err != nil {
		logrus.Println(err)
		logger.Panic(err)
		return
	}
	logger.Println(tm)
}

func TestDisconnect(t *testing.T) {
	tm, err := client.deleteDevice(pluginID, identity)
	if err != nil {
		logger.Panic(err)
		return
	}
	logger.Println(tm)
}

func TestPermitJoin(t *testing.T) {

	attr := sdk.SetAttribute{
		IID: identity,
		AID: 6,
		Val: 20,
	}
	err := client.SetAttributes(pluginID, identity, attr)
	if err != nil {
		logger.Panic(err)
		return
	}
}

func TestGateways(t *testing.T) {

	err := client.Gateways(pluginID, "DoorSensor-EF-3.0")
	if err != nil {
		logger.Panic(err)
		return
	}
}

func TestSubDevices(t *testing.T) {

	err := client.SubDevices(pluginID, identity)
	if err != nil {
		logger.Panic(err)
		return
	}
}

func TestDeviceStates(t *testing.T) {

	err := client.DeviceStates(pluginID, "5c0272fffee7c66f")
	if err != nil {
		logger.Panic(err)
		return
	}
}

func TestAddSubDevice(t *testing.T) {
	// permit join
	TestPermitJoin(t)

	// reset sensor

	// wait device increase
	time.Sleep(time.Second * 20)
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
