# plugin-sdk

### proto gen

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go

go get -u github.com/golang/protobuf/{proto,protoc-gen-go}

cd pkg/plugin/sdk/proto/
protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative  v2/plugin.proto
```

### 插件实现

1) 获取sdk

```shell
go get github.com/zhiting-tech/smartassistant
```

2) 定义设备

```go
package plugin

import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/instance"

type Device struct {
	Light instance.LightBulb
	Info0 instance.Info
	// 根据实际设备功能组合定义
}

func NewDevice() *Device {
	return &Device{
		Light: instance.NewLightBulb(),
		Info0: instance.NewInfo(),
	}
}
```

3) 实现设备接口

```go

package plugin

import (
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/instance"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

type Device struct {
	Light instance.LightBulb
	Info0 instance.Info
	// 根据实际设备功能组合定义
}

func NewDevice() *Device {
	return &Device{
		Light: instance.NewLightBulb(),
		Info0: instance.NewInfo(),
	}
}

func (d *Device) Info() plugin.DeviceInfo {
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

func (d *Device) GetChannel() plugin.WatchChan {
	// 返回WatchChan频道，用于状态变更推送
	panic("implement me")
}

```

4) 初始化和运行

```go
package main

import (
	"log"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
	sdk "github.com/zhiting-tech/smartassistant/pkg/server/sdk"
)

func main() {
	p := plugin.NewPluginServer("demo")
	go func() {
		// 发现设备
		d := NewDevice()
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