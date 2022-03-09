# WebSocket API

通常，一个 WebSocket 消息格式如下：

```json
{
  "id": 1,
  "domain": "",
  "service": "",
  "service_data": {
    "device_id": 1
  }
}
```

* id: 消息ID，必填，服务端会返回对应 ID 的结果
* domain: `plugin`或者插件id

## 设备相关命令

## 插件设备状态变更

```json
{
  "type": "attribute_change",
  "identity": "2762071932",
  "instance_id": 2,
  "attr": {
    "attribute": "power",
    "val": "on",
    "val_type": "string"
  }
}
```

### 发现设备

### req

```json
{
  "id": 1,
  "service": "discover"
}
```

### resp

```json
{
  "id": 1,
  "type": "",
  "result": {
    "device": {
      "name": "zhiting_M1",
      "identity": "hijklmn",
      "model": "M1",
      "manufacturer": "zhiting",
      "plugin_id": "demo"
    }
  },
  "success": true
}
```

### 获取设备属性

#### req

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "get_attributes",
  "identity": "2762071932"
}
```

#### resp

```json
{
  "id": 1,
  "result": {
    "identity": "2762071932",
    "device": {
      "name": "",
      "identity": "2762071932",
      "instances": [
        {
          "type": "light_bulb",
          "instance_id": 0,
          "attrs": [
            {
              "attribute": "power",
              "val": "on",
              "val_type": "string"
            },
            {
              "attribute": "brightness",
              "val": 55,
              "val_type": "int"
            },
            {
              "attribute": "color_temp",
              "val": 3500,
              "val_type": "int"
            }
          ]
        }
      ],
      "ota_support": true
    }
  },
  "success": true
}
```

### 设置设备属性

#### req

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "set_attributes",
  "identity": "2762071932",
  "service_data": {
    "attributes": [
      {
        "instance_id": 1,
        "attribute": "power",
        "val": "on"
      }
    ]
  }
}

```

#### resp

```json
{
  "id": 1,
  "type": "response",
  "success": true,
  "error": "error"
}
```

### 检查设备是否有固件更新

#### req

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "check_update",
  "identity": "2095030692"
}
```

#### resp

```json
{
  "id": 1,
  "type": "response",
  "result": {
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

### 通过插件更新设备固件

#### req

```json
{
  "id": 1,
  "domain": "zhiting",
  "service": "ota",
  "identity": "2095030692"
}
```

#### resp

```json
{
  "id": 1,
  "type": "response",
  "result": null,
  "success": true
}
```