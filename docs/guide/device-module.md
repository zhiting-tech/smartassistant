# 设备

对智汀家庭云即smart-assistant（以下简称SA）的设备模块的说明。

## 品牌

品牌指的是智能设备的品牌，SA通过插件的形式对该品牌下的设备进行发现控制。理论上来说一个品牌对应一个插件服务。您可以通过“我的-支持品牌”查看SA支持的品牌。关于插件服务的详细信息可以参考[plugin](./device-plugin.md)

## 设备的相关操作

在SA中是通过一个个命令对设备进行操作的，如果您想使用这些命令操作某一品牌的设备，首先应该安装该品牌的插件。在SA中安装、更新、 移除插件。请参考[plugin](./device-plugin.md)
SA处理设备命令的流程:客户端通过websocket消息的形式将对应的操作命令发送给SA，SA通过grpc的方式将消息转发给插件服务，插件
服务处理后，将处理的结果通过grpc的方式发送给SA，SA将处理结果以websocket消息返回给客户端。

### 设备的发现与添加

* 发现设备 发现设备需向SA发送以下格式的websocket消息，字段说明: domain: 插件名称；service：设备命令。

```json
{
  "domain": "",
  "id": 1,
  "service": "discover"
}
```

成功后SA会返回以下消息

```json
{
  "id": 1,
  "success": true,
  "result": {
    "device": {
      "id": "21394770a79648e6a3416239e1ebecb9",
      "address": "192.168.0.195:55443",
      "identity": "0x0000000012ed37c8",
      "name": "yeelight_ceiling17_0x0000000012ed37c8",
      "model": "ceiling17",
      "sw_version": "5",
      "manufacturer": "yeelight",
      "power": "on",
      "bright": 80,
      "color_mode": 2,
      "ct": 3017,
      "rgb": 0,
      "hue": 0,
      "sat": 0
    }
  }
}
```

manufacturer之后的字段为设备属性，取决于设备的类型

* 添加设备 将发现设备操作获取的设备主要信息通过添加设备接口以下列格式发送到SA。如果添加的设备为SA，则type为smart_assistant

```json
{
  "device": {
    "name": "nisi dolore eu est",
    "brand_id": "commodo es",
    "address": "pariatur sint",
    "identity": "velit ut ad",
    "model": "proident veniam",
    "type": "nisi Lorem in officia irure",
    "sw_version": "qui ut",
    "manufacturer": "aute Lorem pariatur volu",
    "plugin_id": "dolore reprehenderit"
  }
}
```

SA会将该设备持久化保存在数据库中，之后便可通过插件控制设备。

### 设备控制与设备信息

客户端同样是以websocket消息的形式将命令发送给SA。因为不同类型的设备的命令不一定相同，所以这里只以yeelight灯进行示例展示。更多类型设备的
消息格式请阅读[websocket-api.md](./web-socket-api.md)

设备信息

```json
{
  "domain": "yeelight",
  "id": 1,
  "service": "state",
  "service_data": {
    "device_id": "device_id"
  }
}
```

```json
{
  "id": 1,
  "result": {
    "state": {
      "power": "on/off",
      "brightness": 55,
      "color_temp": 4000
    }
  },
  "success": true
}
```

开关

```json
{
  "domain": "yeelight",
  "id": 1,
  "service": "switch",
  "service_data": {
    "device_id": "device_id",
    "power": "on/off/toggle"
  }
}
```

设置亮度

```json
{
  "domain": "yeelight",
  "id": 1,
  "service": "set_bright",
  "service_data": {
    "device_id": "device_id",
    "brightness": 100
  }
}
```

设置色温

```json
{
  "domain": "yeelight",
  "id": 1,
  "service": "set_color_temp",
  "service_data": {
    "device_id": "device_id",
    "color_temp": 100
  }
}
```

## 设备的权限

SA会从插件的安装目录[插件安装目录](../../static/plugins)读取每一个插件的config.yaml文件以获得该设备具有的操作功能。具体方法可以查看
[获取设备的操作功能](../../internal/orm/device.go)device.go文件中的GetDeviceActions()方法。SA为设备的每一个功能操作设置了权限
控制，这意味着您可能只能控制某个设备的一种或多种功能。关于权限的详细信息，您可以阅读[权限](./user-module.md)。您可以通过获取用户权限接口 来查看您拥有的设备控制权限。

