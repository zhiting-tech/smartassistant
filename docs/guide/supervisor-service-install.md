# smartassistant supervisor 安装

## 依赖

* jq (linux命令行json解析工具)
* docker

* jq安装

Ubuntu:

```shell
sudo apt-get install jq
```

Centos:

```shell
yum install jq
```


## 安装步骤

以下以安装1.9.4版本为例：

* 拉取supervisor镜像, 
```shell
sudo docker pull zt.registry.zhitingtech.com/zhitingtech/supervisor:1.9.4
```

* 复制以下内容到/etc/smartassistant-supervisor/supervisor.conf, runtime_path为smartassistant的安装路径，docker_socket为docker的socket路径，请根据实际情况自行修改
```
{
    "runtime_path": "/mnt/data/zt-smartassistant",
    "docker_socket": "/var/run/docker.sock"
}
```

* 复制以下内容到/etc/smartassistant-supervisor/image.conf
```
{
    "supervisor_image":"zt.registry.zhitingtech.com/zhitingtech/supervisor:1.9.4"
}
```

* 复制以下内容到/usr/bin/smartassistant-supervisor.sh并给予文件可执行权限, OS变量请根据实际情况修改，目前只支持Centos和Ubuntu
```
#!/bin/bash

# 根据实际情况修改
OS="Centos"

loop() {
  for ((;;)) 
  do

    sleep 30

    docker inspect supervisor 2>&1 > /dev/null
    if [[ $? -ne 0 ]]; then
      echo "supervisor container not run!!"
      exit 1
    fi
  done
}

start() {
  IMAGE=$(jq -M -r '.supervisor_image' < /etc/smartassistant-supervisor/image.conf)
  if [[ $? -ne 0 ]]; then
    echo "read supervisor image.conf error!!!"
    exit 1
  fi

  if [[ -z "$IMAGE" ]]; then
    echo "image is empty!!!"
    exit 1
  fi

  ret=$(docker images $IMAGE)
  if [[ $? -ne 0 ]]; then
    echo "docker not start?"
    exit 1
  else 
    if [[ -z "$ret" ]] && [[ -e "/etc/smartassistant-supervisor/images/supervisor.tar" ]]; then
      docker load -i /etc/smartassistant-supervisor/images/supervisor.tar
      if [[ $? -eq 0 ]]; then
        rm -f /etc/smartassistant-supervisor/images/supervisor.tar
      else
        echo "Load supervisor image error!!!"
        exit 1
      fi
    elif [[ -z "$ret" ]] && [[ ! -e "/etc/smartassistant-supervisor/images/supervisor.tar" ]]; then
      echo "supervisor image not found, please reinstall supervisor"
      exit 1
    fi
  fi

  RUNTIMEPATH=$(jq -M -r '.runtime_path' < /etc/smartassistant-supervisor/supervisor.conf)
  if [[ $? -ne 0 ]]; then
    echo "read supervisor supervisor.conf error!!!"
    exit 1
  fi

  if [[ -z "$RUNTIMEPATH" ]]; then
    echo "runtime path is empty!!!"
    exit 1
  fi

  DOCKERSOCK=$(jq -M -r '.docker_socket' < /etc/smartassistant-supervisor/supervisor.conf)
  if [[ $? -ne 0 ]]; then
    echo "read supervisor supervisor.conf error!!!"
    exit 1
  fi

  if [[ -z "$DOCKERSOCK" ]]; then
    DOCKERSOCK=/var/run/docker.sock
  fi

  (docker run -itd --name supervisor --restart=always -v $DOCKERSOCK:/var/run/docker.sock \
            --mount type=bind,source="$RUNTIMEPATH",target=/mnt/data/zt-smartassistant \
            -v /etc/smartassistant-supervisor:/mnt/data/zt-smartassistant/run/supervisor/system/current/rootfs/etc/smartassistant-supervisor \
            --pid=host --net=host --privileged=true "$IMAGE" -os $OS) > /dev/null
  
  loop
}

stop() {
  docker rm -f supervisor
}

case $1 in
  "start")
    start
    ;;

  "stop")
    stop
    ;;

  *)
    echo "Invalid Argment!!!"
    exit 1
    ;;

esac
```

* 复制以下内容到/usr/lib/systemd/system/smartassistant-supervisor.service, 然后执行systemctl deamon-reload
```
[Unit]
Description=Supervisor Service
After=docker.service
Requires=docker.service

[Service]
ExecStart=/usr/bin/smartassistant-supervisor.sh start
ExecStop=/usr/bin/smartassistant-supervisor.sh stop

[Install]
WantedBy=multi-user.target
```

* 启动supervisor服务,查看状态
```shell
systemctl start smartassistant-supervisor
systemctl status smartassistant-supervisor
```