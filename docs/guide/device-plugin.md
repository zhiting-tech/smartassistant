## 设备类插件开发指南

开发前先阅读插件设计概要：[插件系统设计技术概要](plugin-module.md)

使用 [plugin-sdk](../../pkg/plugin/sdk) 可以忽略不重要的逻辑，快速实现插件

### 插件实现

1) 获取sdk

```shell
    go get github.com/zhiting-tech/smartassistant
```

2) 定义设备

sdk中提供了预定义的设备模型，使用模型可以方便SA有效进行管理和控制

请参考[设备模型](device-thing-model.md)

```go
package plugin

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/instance"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
)

type Device struct {
	Light instance.LightBulb
	Info0 instance.Info
	// 根据实际设备功能组合定义
}

func NewDevice() *Device {

	// 定义属性
	lightBulb := instance.LightBulb{
		Power:      attribute.NewPower(),
		ColorTemp:  instance.NewColorTemp(), // 可选字段需要初始化才能使用
		Brightness: nil,                      // 可选字段不初始化则不从接口中显示
	}

	// 定义设备基础属性
	info := instance.Info{
		Name:         attribute.NewName(),
		Identity:     attribute.NewIdentity(),
		Model:        attribute.NewModel(),
		Manufacturer: attribute.NewManufacturer(),
		Version:      attribute.NewVersion(),
	}
	return &Device{
		Light: lightBulb,
		Info0: info,
	}
}
```

3) 实现设备接口 定义好设备之后，需要为设备实现如下几个方法：

```go
package main

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

type Device interface {
	Identity() string             // 获取设备唯一值
	Info() server.DeviceInfo      // 设备详情
	Setup() error                 // 初始化设备属性
	Update() error                // 更新设备所有属性值
	Close() error                 // 回收所有资源
	GetChannel() server.WatchChan // 返回通知channel
}
```

实现如下：

```go

package plugin

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/instance"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

type Device struct {
	Light instance.LightBulb
	Info0 instance.Info
	// 根据实际设备功能组合定义
}

func NewDevice() *Device {
	// 定义属性
	lightBulb := instance.LightBulb{
		Power:      attribute.NewPower(),
		ColorTemp:  attribute.NewColorTemp(), // 可选字段需要初始化才能使
		Brightness: nil,                      // 可选字段不初始化则不从接口中显示
	}

	info := instance.Info{
		Name:         attribute.NewName(),
		Identity:     attribute.NewIdentity(),
		Model:        attribute.NewModel(),
		Manufacturer: attribute.NewManufacturer(),
		Version:      attribute.NewVersion(),
	}
	return &Device{
		Light: lightBulb,
		Info0: info,
	}
}

func (d *Device) Info() server.DeviceInfo {
	// 该方法返回设备的主要信息
	panic("implement me")
}

func (d *Device) Setup() error {
	// 设置设备的属性和相关配置（比如设备id、型号、厂商等，以及设备的属性更新触发函数）
	panic("implement me")
}

func (d *Device) Update() error {
	// 该方法在获取设备所有属性值时调用，通过调用attribute.SetBool()等方法更新
	panic("implement me")
}

func (d *Device) Close() error {
	// 自定义退出相关资源的回收
	panic("implement me")
}

func (d *Device) GetChannel() server.WatchChan {
	// 返回WatchChan频道，用于状态变更推送
	panic("implement me")
}

```

4) 初始化和运行

定义好设备和实现方法后，运行插件服务（包括grpc和http服务）

```go
package main

import (
	"log"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
	sdk "github.com/zhiting-tech/smartassistant/pkg/server/sdk"
)

func main() {
	p := server.NewPluginServer() // 插件服务名
	go func() {
		// 发现设备，并将设备添加到manager中
		d := NewDevice()
		p.Manager.AddDevice(d)
	}()
	err := sdk.Run(p)
	if err != nil {
		log.Panicln(err)
	}
}
```

这样服务就会运行起来，并通过SA的etcd地址0.0.0.0:2379注册插件服务， SA会通过etcd发现插件服务并且建立通道开始通信并且转发请求和命令

### 快速开始

[快速开始](../tutorial/plugin-quickstart.md)