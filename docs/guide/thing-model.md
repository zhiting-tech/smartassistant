# 物模型

## 什么是物模型

智汀为设备定义了一套物模型，用于描述设备的功能。

物模型是物理世界的实体东西的一个抽象，是进行数字化描述后，用于数字世界的数字模型。

以智能灯为例，不同的灯，尽管规格不同，但它们的属性是相似的，比如都有开关状态的属性，功能逻辑也相仿。我们可以将这些特征标准化，形成智能灯的物模型。

第三方开发者通过使用智汀定义的物模型自由组合就可以描述产品/硬件，并将其能力接入到智汀家庭云。

## 物模型功能

- 服务插件开发，使得插件开发者可以方便地将第三方协议转换成物模型
- 使得SA上的设备相关接口统一API，方便第三方云对接，不需要关心设备实际使用的通讯协议
- 一定程度上规范硬件定义以及接口开发

大致等于在”第三方云“和”硬件厂商“之间充当一个翻译者的角色，将不同厂商的协议翻译为物模型，再将物模型翻译成第三方云厂商的物模型

## 设备结构

### 核心概念

- device 具体的、物理意义上的设备
- instance 设备中的实例，任何一个device本身就是一个instance；表现在网关上时，接口中返回多个instance，包括网关、以及子设备
- service 设备的服务（比如灯，开关就是服务，一个设备可以拥有多个相同或不相同的服务）
- attribute 设备的服务上的具体属性 （比如灯的亮度，开关的开关）

### 字段描述

#### instance

| 字段       | 类型     | 描述                     |
|----------|--------|------------------------|
| **iid**  | string | instance id， 等于设备的唯一标识 |
| services | array  | service 数组             | 

#### service

| 字段         | 类型     | 描述          |
|------------|--------|-------------|
| type       | string | service类型   |
| attributes | array  | attribute数组 | 

#### attribute

| 字段         | 类型          | 描述                                |
|------------|-------------|-----------------------------------|
| **aid**    | int         | attribute id(根据instance中属性的数量，自增) |
| type       | string      | 属性类型                              | 
| val        | 取决于val_type | 属性值                               | 
| val_type   | string      | 属性值类型                             | 
| permission | uint        | 权限                                | 
| min        | 取决于val_type | 最小值                               | 
| max        | 取决于val_type | 最大值                               | 

例如，有一个包含开关和灯的网关，网关作为与SA通讯的设备，需要上报所有设备信息（包括自己和子设备）， 以下为该网关需要上报的数据的结构，也等同于网关的物模型：

```json
{
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
  ]
}
```

#### permission

- 类型：uint
- 该字段表示所有属性权限的和，比如：同时拥有读写和通知权限则等于1+2+4=7，可以通过位运算快速计算是否拥有某个权限

| 权限  | 值   | 描述      |
|-----|-----|---------|
| 读   | 1   | 属性的读权限  |
| 写   | 2   | 属性的写权限  |
| 通知  | 4   | 属性的通知权限 |
