# WebSocket API

## 消息结构

通常，一个 WebSocket 消息格式如下：

| 字段      | 类型     | 描述                     | 所属消息类型                 |
|---------|--------|------------------------|------------------------|
| type    | string | 消息类型                   | response/event         |
| service | string | 请求服务类型                 | request                |
| event   | string | 事件类型                   | event/request          |
| data    | Object | 消息的自定义数据               | event/request/response |
| domain  | string | 请求消息中使用，除特殊请求外均为插件id   | request                |
| id      | int64  | 消息ID，请求消息时必填，响应与请求id一致 | request/response       |
| success | bool   | 响应时，返回是否成功             | response               |
| error   | string | 响应时，返回错误               | response               |

当前有三种消息：

- 客户端发起的请求消息，如下
    ```json
    {
        "id": 1,
        "domain": "example",
        "service": "example",
        "data": {
            "custom_field": 1
        }
    }
    ```
- 服务端对客户端请求的响应消息，如下
    ```json
    {
        "type": "response",
        "id": 1,
        "data": {
            "custom_field": 1
        },
        "success": true,
        "error":"invalid argument"
    }
    ```
- 服务端根据客户端订阅响应的事件消息，如下
    ```json
    {
        "type": "event",
        "event": "example",
        "data": {
            "custom_field": 1
        }
    }
    ```

## Event

事件消息等同于订阅请求的响应

### 订阅设备是否在线消息

- data.plugin_id: 插件id，可选，不填则订阅所有插件
- data.iid: 设备id，可选，不填则订阅所有设备

```json
{
  "id": 1,
  "service": "subscribe_event",
  "event": "online_status",
  "data": {
    "plugin_id": "zhiting",
    "iid": "device_iid"
  }
}
```

```json
{
  "id": 1,
  "success": true,
  "err": "event invalid"
}
```

```json
{
  "id": 1,
  "type": "event",
  "event": "online_status",
  "data": {
    "plugin_id": "zhiting",
    "iid": "2762071932",
    "online": true
  }
}
```

### 订阅插件设备状态变更

- data.plugin_id: 插件id，可选，不填则订阅所有插件
- data.iid: 设备id，可选，不填则订阅所有设备

订阅属性

```json
{
  "id": 1,
  "service": "subscribe_event",
  "event": "attribute_change",
  "data": {
    "plugin_id": "zhiting",
    "iid": "device_iid"
  }
}
```

```json
{
  "id": 1,
  "success": true,
  "err": "event invalid"
}
```

```json
{
  "id": 1,
  "type": "event",
  "event": "attribute_change",
  "data": {
    "plugin_id": "zhiting",
    "attr": {
      "iid": "2762071932",
      "aid": 1,
      "val": "on"
    }
  }
}
```

### 设备增加

```json
{
  "id": 1,
  "service": "subscribe_event",
  "event": "device_increase"
}
```

```json
{
  "id": 1,
  "success": true,
  "err": "event invalid"
}
```

```json
{
  "event_type": "device_increase",
  "data": {
    "device": {
      "id": 7,
      "name": "DoorSensor-EF-3.0",
      "plugin_id": "zhiting",
      "iid": "5c0272fffee7c66f",
      "model": "DoorSensor-EF-3.0",
      "manufacturer": "zhiting",
      "type": "window_door_sensor",
      "location_id": 0,
      "department_id": 0,
      "area_id": 59268327347862572
    }
  },
  "type": "event"
}
```

## 发现设备

### request

```json
{
  "id": 1,
  "service": "discover"
}
```

### response

服务器会分多次响应设备

- auth_required: 表示是否需要认证/配对,true时需要提供足够的信息才能连接设备
- auth_param.name: 参数名字
- auth_param.type: 参数类型，（string/int/bool/float/select）
- auth_param.required: 是否必须
- auth_param.default: 默认值，没有则不返回该字段
- auth_param.min: 参数int/float时限制最小值，没有则不返回
- auth_param.max: 参数int/float时限制最大值，没有则不返回
- auth_param.options: type为select时的可选值，不是则不返回
- auth_param.option.name: 可选值名字
- auth_param.option.val: 可选值的值

```json
{
  "id": 1,
  "type": "",
  "data": {
    "device": {
      "name": "zhiting_M1",
      "iid": "hijklmn",
      "model": "M1",
      "manufacturer": "zhiting",
      "plugin_id": "demo",
      "auth_required": true,
      "auth_params": [
        {
          "name": "pin",
          "type": "string",
          "required": true,
          "default": "",
          "min": 0,
          "max": 10,
          "options": [
            {
              "name": "A",
              "val": "a"
            }
          ]
        }
      ]
    }
  },
  "success": true
}
```

## 连接设备/添加设备

### request

- auth_params：跟据[发现的响应](#request)的参数将auth_param.name作为key，将用户输入或者选择作为val

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "connect",
  "data": {
    "iid": "2095030692",
    "auth_params": {
      "pin": "12345678",
      "username": "username",
      "password": "123456"
    }
  }
}
```

### response

```json
{
  "id": 1,
  "data": {
    "instances": [
      {
        "iid": "id111",
        "services": [
          {
            "type": "gateway",
            "attributes": [
              {
                "aid": 1,
                "type": "on_off",
                "val_type": "",
                "permission": 0,
                "val": null
              }
            ]
          }
        ]
      },
      {
        "iid": "id222",
        "services": [
          {
            "type": "light",
            "attributes": [
              {
                "aid": 1,
                "type": "on_off",
                "val_type": "",
                "permission": 0,
                "val": null
              },
              {
                "aid": 2,
                "type": "brightness",
                "val_type": "",
                "min": 1,
                "max": 100,
                "permission": 0,
                "val": null
              }
            ]
          }
        ]
      },
      {
        "iid": "id333",
        "services": [
          {
            "type": "switch",
            "attributes": [
              {
                "aid": 1,
                "type": "on_off",
                "val_type": "",
                "permission": 0,
                "val": null
              }
            ]
          },
          {
            "type": "switch",
            "attributes": [
              {
                "aid": 2,
                "type": "on_off",
                "val_type": "",
                "permission": 0,
                "val": null
              }
            ]
          },
          {
            "type": "switch",
            "attributes": [
              {
                "aid": 3,
                "type": "on_off",
                "val_type": "",
                "permission": 0,
                "val": null
              }
            ]
          }
        ]
      }
    ],
    "ota_support": true,
    "device": {
      "id": 1,
      "model": "model",
      "plugin_id": 1,
      "plugin_url": "http://127.0.0.1/index.html"
    }
  },
  "success": true
}
```

## 获取设备物模型

### request

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "get_instances",
  "data": {
    "iid": "2095030692"
  }
}
```

### response

```json
{
  "id": 1,
  "data": {
    "instances": [
      {
        "iid": "id111",
        "services": [
          {
            "type": "gateway",
            "attributes": [
              {
                "aid": 1,
                "type": "on_off",
                "val_type": "",
                "permission": 0,
                "val": null
              }
            ]
          }
        ]
      },
      {
        "iid": "id222",
        "services": [
          {
            "type": "light",
            "attributes": [
              {
                "aid": 1,
                "type": "on_off",
                "val_type": "",
                "permission": 0,
                "val": null
              },
              {
                "aid": 2,
                "type": "brightness",
                "val_type": "",
                "min": 1,
                "max": 100,
                "permission": 0,
                "val": null
              }
            ]
          }
        ]
      },
      {
        "iid": "id333",
        "services": [
          {
            "type": "switch",
            "attributes": [
              {
                "aid": 1,
                "type": "on_off",
                "val_type": "",
                "permission": 0,
                "val": null
              }
            ]
          },
          {
            "type": "switch",
            "attributes": [
              {
                "aid": 2,
                "type": "on_off",
                "val_type": "",
                "permission": 0,
                "val": null
              }
            ]
          },
          {
            "type": "switch",
            "attributes": [
              {
                "aid": 3,
                "type": "on_off",
                "val_type": "",
                "permission": 0,
                "val": null
              }
            ]
          }
        ]
      }
    ],
    "sync_data": "",
    "ota_support": true,
    "auth_required": true,
    "is_auth": true,
    "auth_params": [
      {
        "name": "pin",
        "type": "string",
        "required": true,
        "default": "",
        "min": 0,
        "max": 10,
        "options": [
          {
            "name": "A",
            "val": "a"
          }
        ]
      }
    ]
  },
  "success": true
}
```

## 设置设备属性

### request

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "set_attributes",
  "data": {
    "attributes": [
      {
        "iid": "2095030692",
        "aid": 1,
        "val": "on"
      }
    ]
  }
}

```

### response

```json
{
  "id": 1,
  "type": "response",
  "success": true,
  "error": "error"
}
```

## 检查设备是否有固件更新

### request

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "check_update",
  "data": {
    "iid": "2095030692"
  }
}
```

### response

```json
{
  "id": 1,
  "type": "response",
  "data": {
    "current_version": "1.0.4",
    "latest_firmware": {
      "version": "1.0.4",
      "url": "https://sc.zhitingtech.com:11110/download/ZT-SW3ZLW001W_v1.0.4.bin",
      "info": "1.0.4更新"
    },
    "update_available": false
  },
  "success": true
}
```

## 通过插件更新设备固件

### request

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "ota",
  "data": {
    "iid": "2095030692"
  }
}
```

### response

```json
{
  "id": 1,
  "type": "response",
  "data": null,
  "success": true
}
```

## 断开连接/删除设备

### request

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "disconnect",
  "data": {
    "iid": "2095030692"
  }
}

```

### response

```json
{
  "id": 1,
  "type": "response",
  "success": true,
  "error": "error"
}
```

## 设备状态变更（日志）

### request

- size: 分页大小
- attr_type: 属性类型，不传则返回所有属性
- index: 不传则从头获取，滚动加载时使用最后一个state的id
- start_at:开始时间，不传则没有限制
- end_at: 结束时间，不传则不限制

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "device_states",
  "data": {
    "iid": "2095030692",
    "attr_type": "on_off",
    "size": 20,
    "index": 10,
    "start_at": 1645152639,
    "end_at": 1645152639
  }
}
```

### response

```json
{
  "id": 1,
  "type": "response",
  "data": {
    "states": [
      {
        "id": 10,
        "timestamp": 1645152639,
        "type": "on_off",
        "val_type": "string",
        "val": "on"
      }
    ]
  },
  "success": true
}
```

## 网关列表

### request

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "list_gateways",
  "data": {
    "model": "model"
  }
}
```

### response

```json
{
  "id": 1,
  "type": "response",
  "data": {
    "gateways": [
      {
        "name": "gateway",
        "iid": "iid",
        "model": "gateway",
        "logo_url": "www.example.com/logo.png",
        "plugin_id": "example"
      }
    ],
    "support_gateways": [
      {
        "model": "gateway",
        "logo_url": "www.example.com/logo.png",
        "plugin_id": "example"
      }
    ]
  },
  "success": true
}
```

## 子设备列表

### request

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "sub_devices",
  "data": {
    "iid": "2095030692"
  }
}
```

### response

```json
{
  "id": 1,
  "type": "response",
  "data": {
    "devices": [
      {
        "name": "motion_sensor",
        "logo_url": "www.example.com/logo.png",
        "plugin_url": "www.example.com/plugin/index.html",
        "zigbee_support": true,
        "ble_support": false
      }
    ],
    "support_sub_devices": [
      {
        "zigbee_support": true,
        "ble_support": false,
        "model": "motion_sensor",
        "logo_url": "www.example.com/logo.png",
        "provisioning_url": "www.example.com/plugin/provisioning.html"
      }
    ]
  },
  "success": true
}
```