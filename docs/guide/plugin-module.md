## 插件系统设计技术概要

### 简述

- 当前所说的插件仅指`设备类插件`，插件为SA提供额外的设备发现和控制功能；
- 插件通过实现定义的grpc接口，以grpc服务的形式运行，提供接口给SA调用
- 插件同时需要http服务提供h5页面及静态文件

### SA中插件的工作流程

#### 插件部署流程

1) 插件开发者将开发好的插件服务编译成docker镜像提供给SA
2) SA根据插件的镜像地址判断本地是否已经拉取或更新
3) 用户安装插件后，SA根据镜像运行起容器，插件往注册中心注册服务
4) SA通过服务发现发现新的插件服务

#### 插件使用流程

1) 用户在界面上发现设备时对所有插件服务调用Discover接口，插件根据实现的接口返回所发现的设备
2) 用户添加设备并标记设备对应的插件
3) 用户请求设备的H5地址，进去插件自定义页面
4) 通过交互发起自定义指令给SA，SA将指令转发给对应的插件服务

### 接口

- 文件http服务 sdk提供了方便的方法进行静态文件挂载和自定义api接口实现

  ```go
  package main
  
  import "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
  
  func main() {
      ps := server.NewPluginServer()
      ps.HtmlRouter.Static("", "./html")
  
      apiGroup := ps.Router.Group("/api/")
      apiGroup.GET("")
  }
  ```

- grpc服务， 通过实现protobuf定义的grpc接口来实现插件服务：
    ```proto
    syntax = "proto3";
    package proto;
    option go_package = "../proto";
  
    service Plugin {
    // Discover 发现时设备
    rpc Discover (empty) returns (stream device);
    rpc StateChange (empty) returns (stream state);
    rpc HealthCheck (healthCheckReq) returns (healthCheckResp);
    // GetAttributes TODO 考虑删除该接口，仅通过Connect获取模型，并通过回调更新属性
    rpc GetAttributes (GetAttributesReq) returns (GetAttributesResp);
    rpc SetAttributes (SetAttributesReq) returns (SetAttributesResp);
  
    rpc Connect (AuthReq) returns (GetAttributesResp);
    rpc Disconnect (AuthReq) returns (empty);
    }
  
    message AuthReq {
        string identity = 1;
        map<string, string> params = 2;
    }
  
    message ExecuteReq {
        string identity = 1;
        string cmd = 2;
        bytes data = 3;
    }
    message ExecuteResp {
        bool success = 1;
        string error = 2;
        bytes data = 3;
    }
    message GetAttributesReq {
        string identity = 1;
    }
    
    message GetAttributesResp {
        bool success = 1;
        string error = 2;
        repeated Instance instances = 3;
    }
    message Instance {
        string identity = 1;
        int32 instance_id = 2;
        bytes attributes = 3;
        string type = 4;
    }
    
    message SetAttributesReq {
        string identity = 1;
        bytes data = 2;
    }
    
    message SetAttributesResp {
        bool success = 1;
        string error = 2;
    }
    
    message Action {
        string identity = 1;
        int32 instance_id = 2;
       bytes attributes = 3;
    }
    
    message device {
        string identity = 1;
        string model = 2;
        string manufacturer = 3;
    bool authRequired = 4;
    }
    
    message empty {
    }
    
    message state {
        string identity = 1;
        int32 instance_id = 2;
        bytes attributes = 3;
    }
  
    message healthCheckReq {
    string identity = 1;
    }
  
    message healthCheckResp {
    string identity = 1;
    bool online = 2;
    }
    ```

注：grpc接口是通用的定义，SDK对接口实现了封装，开发者使用SDK时不需要关心，仅需要实现设备类型即可。

### sdk

为了方便开发者快速开发插件以及统一接口，我们提供sdk规范了接口以及预定义了设备模型，以下为sdk实现功能：

- 插件服务注册
- http服务
- grpc服务以及接口封装（包括设备属性获取、属性设置、消息通知等）
- 预定义模型

### 设备模型设计

#### 背景

云对云接入时，需要对第三方云的命令进行解析，并通过SA对插件发起命令。

这就要求插件实现的命令必须要有统一的规范和标准，这样第三方就可以通过这个标准来控制SA的所有支持的设备。

同时也能方便SA更好的通过统一的接口以及命令来管理设备。

#### 模型设计

SDK预定义设备类型以及属性，开发者通过引入设备类型实现相关功能。

SDK通过反射获取设备的所有属性，将属性与命令做好对应关系，这样可以使得无论设备是什么形态，都能有统一的接口以及命令进行控制。

[模型定义](device-thing-model.md)

操作某个属性时，根据属性的tag对命令中的值进行解析和校验 模型例子如下：

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

### 示例项目

- [demo-plugin](../../pkg/plugin/sdk/demo)

### 快速开发

- [开发指南](device-plugin.md)
- [快速开始](../tutorial/plugin-quickstart.md)