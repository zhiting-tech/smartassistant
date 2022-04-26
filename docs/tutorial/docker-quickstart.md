# 使用Docker运行智汀家庭云
本文档描述如何在docker上运行智汀家庭云（以下简称SA），并使用浏览器体验SA。

## 环境准备
- linux主机
- docker
- docker-compose

## 运行SA

### docker-compose文件运行

选择一个目录存放SA运行所需文件，执行以下命令：

```shell
wget https://github.com/zhiting-tech/smartassistant/releases/download/v1.9.1/smartassistant.1.9.1.zip && unzip smartassistant.1.9.1.zip

cd zt-smartassistant

docker-compose up
```

### 使用脚本运行SA

#### 注意事项

**运行脚本所需环境**

- awk命令
- openssl
- root用户权限

#### 安装说明

- 执行./build/install/install.sh

![执行安装脚本](../images/exec_install.png)

- 填写配置并确认

![填写SA配置](../images/install_config.png)

- 安装成功

![安装成功](../images/install_success.png)

- 检查运行情况

![运行检查](../images/install_check.png)

### 测试运行情况

服务启动后可以通过以下命令检查SA的状态：

```shell
curl http://localhost:37965/api/check
```

如果返回以下内容则说明服务已运行起来并且未被绑定

```json
{"status":0,"reason":"成功","data":{"is_bind":false,"revision":""}}
```

## 使用智汀App进行体验

### SA体验演示

- 点击添加智能设备发现SA

![点击添加设备](../images/app_smartassistant_ready_add.png)

- 选中SA,点击**添加**

![点击添加设备](../images/app_smartassistant_add.png)

![正在添加设备](../images/app_smartassistant_add_2.png)

![设置是否允许找回用户凭证](../images/app_smartassistant_add_3.png)

![添加设备成功](../images/app_smartassistant_add_success.png)

- 点击**支持品牌**，界面中会列出SA内置的一些插件

![查询支持品牌](../images/app_search_plugin.png)

- 点击其中的一个插件，点击**添加**对插件进行安装

![添加插件](../images/app_plugin_add.png)

![添加插件成功](../images/app_plugin_add_success.png)

- 插件添加成功后，返回到首页**家居**，点击右上角**+**按钮就可以添加插件支持的设备

![准备添加设备](../images/app_ready_add_device_home.png)

- 选中其中的一个设备点击**添加**,添加完成后返回到首页的**家居**即可对设备进行操控

![添加设备](../images/app_device_add.png)

![添加设备成功](../images/app_device_add_success.png)

## 使用专业版进行体验

### 账号和密码设置

SA服务启动后，如果未绑定，需要发送以下请求对SA进行绑定：

```shell
curl -X POST -d '{"device": {"model": "smart_assistant"}}' http://localhost:37965/api/devices
```

绑定成功后返回以下内容：

```json
{"status":0,"reason":"成功","data":{"device_id":1,"plugin_url":"","user_info":{"user_id":1,"role_infos":null,"account_name":"","nickname":"wyqicai","token":"MTYzNTMxOTE1M3xOd3dBTkVoRE5WRXpTME5CV0VrMFVFOVFWa0pJUjA1UFNVRklRMEUxUkZaWVFsRkNWVWhUVWxWRFEwNUVTek5QU0VwVlJVVlJSMEU9fNXschLMFQtEogQo2AlJu4wfSJfLrsk994FwSGUhp-_3","phone":"","is_set_password":false},"area_info":{"id":"20763937284831135"}}}
```

SA绑定成功后设置用户名和密码：

```shell
curl -X PUT -d '{"account_name": "admin","password": "123456"}' --header 'smart-assistant-token:MTYzNTMxOTE1M3xOd3dBTkVoRE5WRXpTME5CV0VrMFVFOVFWa0pJUjA1UFNVRklRMEUxUkZaWVFsRkNWVWhUVWxWRFEwNUVTek5QU0VwVlJVVlJSMEU9fNXschLMFQtEogQo2AlJu4wfSJfLrsk994FwSGUhp-_3' http://localhost:37965/api/users/1

其中header参数smart-assistant-token值为绑定成功后的token, 
请求url path（/api/users/:id）中的路径参数id为绑定成功后返回的用户id。

```

设置成功后返回：

```json
{"status":0,"reason":"成功"}
```

当用户账号和密码设置成功后打开浏览器，访问 http://localhost:9020 即可体验SA了。

### SA体验演示

- 使用账号密码登录

![SA登录](../images/sa_login.png)

- 登录成功后，切换到**我的**

![我的](../images/mine.png)

- 点击**支持品牌**，界面中会列出SA内置的一些插件

![插件列表](../images/plugin_list.png)

- 点击其中的一个插件对插件进行安装

![添加插件](../images/add_plugin.png)

![成功添加插件](../images/add_plugin_success.png)

- 插件添加成功后，返回到首页**家居**，点击右上角**+**按钮就可以添加插件支持的设备

![添加设备](../images/device_list.png)

- 选中其中的一个设备点击**添加**,添加完成后返回到首页的**家居**即可对设备进行操控

![设备添加成功](../images/add_device_success.png)

![控制设备](../images/device_control.png)