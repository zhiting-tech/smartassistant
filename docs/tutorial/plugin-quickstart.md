# 开发您的第一个插件

此文档描述如何开发一个简单插件，面向插件开发者。

开发前先阅读插件设计概要：[插件系统设计技术概要](../guide/plugin-module.md)

### 插件实现

1) 获取sdk

```shell
    go get github.com/zhiting-tech/smartassistant
```

2) 定义协议

```go
package plugin

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

// 定义描述协议的结构
type ProtocolDevice struct {
	light *definer.BaseService
	info *definer.BaseService
	
	// 插件所支持的协议
	pc   protocol
}

func NewDevice(pc  protocol) sdk.Device {
	return &ProtocolDevice{
		id:           pc.GetID(),
		model:        pc.GetModel(),
		manufacturer: pc.GetManufacturer(),
		pc:           pc,
	}
}

```

3) 定义协议所需属性和信息

```go
import "github.com/zhiting-tech/smartassistant/pkg/thingmodel"

// 定义属性或协议信息
// 通过实现thingmodel.IAttribute的接口，以便sdk调用
type OnOff struct {
	pd *ProtocolDevice
}

func (l OnOff) Set(val interface{}) error {
	pwrState := map[]interface{}{
		"pwr": val,
    }
	resp, err := l.pd.pc.SetState(pwrState)
	if err != nil {
		return err
	}
	// 设置属性完成后需要，通知到 smartassistant
	return l.pd.Switch.Notify(thingmodel.OnOff, val)
}

func (l OnOff) Get() (interface{}, error) {
	resp, err := l.pd.pc.GetState()
	if err != nil {
		return nil, err
	}
	pwr, ok := resp["pwr"]
	if !ok {
		return nil, fmt.Errorf("on off get error is state nil")
	}
	return pwr, nil
}


type Model struct {
	pd *ProtocolDevice
}

func (l Model) Get() (interface{}, error) {
	return l.pd.model, nil
}

func (l Model) Set(interface{}) error {
	return nil
}

type Manufacturer struct {
	pd *ProtocolDevice
}

func (l Manufacturer) Get() (interface{}, error) {
	return l.pd.manufacturer, nil
}

func (l Manufacturer) Set(interface{}) error {
	return nil
}

type Identify struct {
	pd *ProtocolDevice
}

func (l Identify) Get() (interface{}, error) {
	return l.pd.id, nil
}

func (l Identify) Set(interface{}) error {
	return nil
}

```
4) 实现协议接口

```go

package plugin

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

// 定义描述协议的结构
type ProtocolDevice struct {
	Switch *definer.BaseService
	info *definer.BaseService
	
	pc   protocol   // 所描述的协议
}

func NewDevice(pc protocol) sdk.Device {
	return &ProtocolDevice{
		id:           pc.GetID(),
		model:        pc.GetModel(),
		manufacturer: pc.GetManufacturer(),
		pc:           pc,
    }
}

func (pd *ProtocolDevice) Info() sdk.DeviceInfo {
	// 该方法返回设备的主要信息
	return sdk.DeviceInfo{
		IID:          pd.id,
		Model:        pd.model,
		Manufacturer: pd.manufacturer,
	}
}

func (pd *ProtocolDevice) Define(def *definer.Definer) {
	// 设置符合该协议设备的属性和相关配置（比如设备id、型号、厂商等，以及设备的属性）
	
	// 对每个属性和配置都可以有权限
	thingmodel.OnOff.WithPermissions(
		thingmodel.AttributePermissionWrite,
		thingmodel.AttributePermissionRead,
		thingmodel.AttributePermissionNotify,
	)
	pd.Switch = def.Instance(pd.id).NewSwitch()
	pd.Switch.Enable(thingmodel.OnOff, OnOff{pd})

	pd.info = def.Instance(pd.id).NewInfo()
	pd.info.Enable(thingmodel.Model, Model{pd})
	pd.info.Enable(thingmodel.Manufacturer, Manufacturer{pd})
	pd.info.Enable(thingmodel.Identify, Identify{pd})
	return
}

func (pd *ProtocolDevice) Connect() error {
	// 提供给sdk主动进行设备tcp连接
	return nil
}

func (pd *ProtocolDevice) Disconnect() error {
	// 提供给sdk主动断开设备tcp连接
	return nil
}

func (pd *ProtocolDevice) Online(iid string) bool {
	// sdk 调用该接口检测设备是否在线
	return true
}

func Discover(ctx context.Context, devices chan<- sdk.Device) {
    // 这里需要实现一个发现设备的方法,给sdk调用
	discoverer := NewDiscoverer()
	go discoverer.Run()
	defer discoverer.Close()

	for {
		select {
		case d, ok := <-discoverer.C:
			if !ok {
				return
			}
			l := NewDevice(d)
			devices <- l
		}
	}
	return
}

```

5) 初始化和运行

```go
package main

import (
	"log"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
)

func main() {
	p := server.NewPluginServer(Discover)
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
