# 开发您的第一个插件

此文档描述如何开发一个简单插件，面向插件开发者。

开发前先阅读插件设计概要：[插件系统设计技术概要](../guide/plugin-module.md)

### 插件实现

1) 获取sdk

```shell
    go get github.com/zhiting-tech/smartassistant
```

2) 定义设备

```go
package plugin

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/instance"
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
		Power:     attribute.NewPower(),
		ColorTemp: attribute.NewColorTemp(), // 根据需要初始化可选字段
	}

	info := instance.Info{
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

3) 实现设备接口

```go

package plugin

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/instance"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

type Device struct {
	LightBulb instance.LightBulb
	Info0     instance.Info
	// 根据实际设备功能组合定义
	identity string
	ch       server.WatchChan
}

func NewDevice(identity string) *Device {
	// 定义设备属性
	lightBulb := instance.LightBulb{
		Power:      attribute.NewPower(),
		ColorTemp:  instance.NewColorTemp(),
		Brightness: instance.NewBrightness(),
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
		LightBulb: lightBulb,
		Info0:     info,
		identity:  identity,
		ch:        make(chan server.Notification, 5),
	}
}

func (d *Device) Info() server.DeviceInfo {
	// 该方法返回设备的主要信息
	return d.identity
}

func (d *Device) update(attr string) attribute.UpdateFunc {
	return func(val interface{}) error {
		switch attr {
		case "power":
			d.LightBulb.Power.SetString(val.(string))
		case "brightness":
			d.LightBulb.Brightness.SetInt(val.(int))
		case "color_temp":
			d.LightBulb.ColorTemp.SetInt(val.(int))
		}

		n := server.Notification{
			Identity:   d.identity,
			InstanceID: 1,
			Attr:       attr,
			Val:        val,
		}
		select {
		case d.ch <- n:
		default:
		}

		return nil
	}
}

func (d *Device) Setup() error {
	// 设置设备的属性和相关配置（比如设备id、型号、厂商等，以及设备的属性更新触发函数）
	d.Info0.Identity.SetString("123456")
	d.Info0.Model.SetString("model")
	d.Info0.Manufacturer.SetString("manufacturer")

	d.LightBulb.Brightness.SetRange(1, 100)
	d.LightBulb.ColorTemp.SetRange(1000, 5000)
	
	// 给属性设置更新函数，在执行命名时，该函数会被执行
	d.LightBulb.Power.SetUpdateFunc(d.update("power"))
	d.LightBulb.Brightness.SetUpdateFunc(d.update("brightness"))
	d.LightBulb.ColorTemp.SetUpdateFunc(d.update("color_temp"))
	return nil
}

func (d *Device) Update() error {
	// 该方法在获取设备所有属性值时调用，通过调用attribute.SetBool()等方法更新
	// d.LightBulb.Power.SetString("on")
	// d.LightBulb.Brightness.SetInt(100)
	// d.LightBulb.ColorTemp.SetInt(2000)
	return nil
}

func (d *Device) Close() error {
	// 自定义退出相关资源的回收
	close(d.ch)
	return nil
}

func (d *Device) GetChannel() server.WatchChan {
	// 返回WatchChan频道，用于状态变更推送
	return d.ch
}

```

4) 初始化和运行

```go
package main

import (
	"log"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
	"github.com/zhiting-tech/smartassistant/pkg/server/sdk"
)

func main() {
	p := server.NewPluginServer()
	go func() {
		// 发现设备
		d := NewDevice("abcdefg")
		p.Manager.AddDevice(d)
	}()
	err := sdk.Run(p)
	if err != nil {
		log.Panicln(err)
	}
}
```

### 镜像编译和部署

暂时仅支持以镜像方式安装插件，调试正常后，编译成镜像提供给SA

- Dockerfile示例参考

```dockerfile
FROM golang:1.16-alpine as builder
RUN apk add build-base
COPY . /app
WORKDIR /app
RUN go env -w GOPROXY="goproxy.cn,direct"
RUN go build -ldflags="-w -s" -o demo-plugin

FROM alpine
WORKDIR /app
COPY --from=builder /app/demo-plugin /app/demo-plugin

# static file
COPY ./html ./html
ENTRYPOINT ["/app/demo-plugin"]

```

- 编译镜像

```shell
docker build -f your_plugin_Dockerfile -t your_plugin_name
```

- 运行插件

```shell
docker run -net=host your_plugin_name

//注意：-net=host 参数只有linux环境才有用。
```

- 更多

[设备类插件开发指南](../guide/device-plugin.md)

### Demo

[demo-plugin](../../examples/plugin-demo) :
通过上文的插件实现教程实现的示例插件；这是一个模拟设备写的一个简单插件服务，不依赖硬件，实现了核心插件的功能
