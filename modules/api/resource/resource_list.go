package resource

import (
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/extension"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"io"
	"math"
	"sync"
	"time"
)

type Resp struct {
	PerCpuUsage          float64   `json:"percpu_usage"`           // cpu使用率
	MemUsage             string    `json:"mem_usage"`              // 已使用内存
	MemTotal             string    `json:"mem_total"`              // 总内存
	Service              []Service `json:"service"`                // 容器服务
	BasisPerCpuUsage     float64   `json:"basis_percpu_usage"`     // 基础服务cpu使用率
	BasisMemUsage        string    `json:"basis_mem_usage"`        // 基础服务已用内存
	PluginPerCpuUsage    float64   `json:"plugin_percpu_usage"`    // 插件服务cpu使用率
	PluginMemUsage       string    `json:"plugin_mem_usage"`       // 插件服务已用内存
	ExtensionPerCpuUsage float64   `json:"extension_percpu_usage"` // 拓展服务cpu使用率
	ExtensionMemUsage    string    `json:"extension_mem_usage"`    // 拓展服务已用内存
}

type Service struct {
	Id          string  `json:"id"`           // 容器id
	Name        string  `json:"name"`         // 容器名称
	ServiceName string  `json:"service_name"` //
	State       string  `json:"state"`        // 运行状态,running 运行中，exited 已暂停
	RunTime     string  `json:"run_time"`     // 运行时间
	PerCpuUsage float64 `json:"percpu_usage"` // cpu使用率
	MemUsage    string  `json:"mem_usage"`    // 已使用内存
	Type        uint8   `json:"type"`         // 容器服务类型,基础服务 1，插件类型 2，拓展服务 3
}

// 智慧中心基础服务
var basisList = map[string]struct{}{
	"smartassistant": {},
	"etcd":           {},
	"zt-nginx":       {},
	"disk-manager":   {},
}

var serviceList = map[string]string{ // 服务名称映射
	"crm":            "客户管理服务",
	"wangpan":        "智汀云盘服务",
	"scm":            "供应链管理服务",
	"smartassistant": "智慧中心系统服务",
	"etcd":           "插件扫描服务",
	"disk-manager":   "磁盘管理服务",
	"zt-nginx":       "反向代理服务",
}

// ListResource 资源列表（cpu、内存）
func ListResource(c *gin.Context) {
	var (
		err                  error
		resp                 Resp
		basisPerCpuUsage     float64
		basisMemUsage        uint64
		pluginPerCpuUsage    float64
		pluginMemUsage       uint64
		extensionPerCpuUsage float64
		extensionMemUsage    uint64
		flag                 bool
		mutex                sync.Mutex

		memTotal uint64
		memUsage uint64
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	dClient := docker.GetClient().DockerClient
	containerList, err := dClient.ContainerList(c, types.ContainerListOptions{All: true})
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	var wg sync.WaitGroup
	for _, container := range containerList {

		wg.Add(1)

		go func(container types.Container) {

			defer wg.Done()
			var service Service
			service.Id = container.ID

			switch container.State {
			case "running":
				service.State = "运行中"
			case "exited":
				service.State = "已暂停"
			}

			conJson, err := dClient.ContainerInspect(c, container.ID)
			if err != nil {
				err = errors.Wrap(err, errors.InternalServerErr)
				return
			}

			startTime := conJson.State.StartedAt
			timeStr := startTime[:10] + " " + startTime[11:19]
			theTime, err := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)

			service.RunTime = FormatTimeSize(time.Now().Unix() - theTime.Unix())

			stats, err := dClient.ContainerStats(c, container.ID, false)
			if err != nil {
				err = errors.Wrap(err, errors.InternalServerErr)
				return
			}
			defer stats.Body.Close()

			b := make([]byte, 10240)
			n, err := stats.Body.Read(b)
			if err != nil && err != io.EOF {
				err = errors.Wrap(err, errors.InternalServerErr)
				return
			}

			var res SystemResource
			if err = json.Unmarshal(b[:n], &res); err != nil {
				err = errors.Wrap(err, errors.InternalServerErr)
				return
			}

			service.MemUsage = FormatFileSize(res.GetUsedMem())
			service.PerCpuUsage = math.Ceil(res.GetPerUsedCpu())
			memTotal = res.MemoryStats.Limit

			mutex.Lock()
			defer mutex.Unlock()

			_, flag = basisList[container.Labels["com.docker.compose.service"]]
			if container.Labels["com.zhiting.smartassistant.resource.service_type"] == "plugin" {
				service.Type = 2
				service.Name = container.Labels["com.zhiting.smartassistant.resource.service_name"]
				service.ServiceName = container.Labels["com.zhiting.smartassistant.resource.service_name"]
				pluginPerCpuUsage += res.GetPerUsedCpu()
				pluginMemUsage += res.GetUsedMem()
				memUsage += res.GetUsedMem()
				resp.PerCpuUsage += res.GetPerUsedCpu()
			} else if extension.HasExtensionWithContext(c, container.Labels["com.docker.compose.service"]) {
				service.Type = 3
				service.Name = container.Labels["com.docker.compose.service"]
				service.ServiceName = serviceList[container.Labels["com.docker.compose.service"]]
				extensionPerCpuUsage += res.GetPerUsedCpu()
				extensionMemUsage += res.GetUsedMem()
				memUsage += res.GetUsedMem()
				resp.PerCpuUsage += res.GetPerUsedCpu()
			} else if flag {
				service.Type = 1
				service.Name = container.Labels["com.docker.compose.service"]
				service.ServiceName = serviceList[container.Labels["com.docker.compose.service"]]
				basisPerCpuUsage += res.GetPerUsedCpu()
				basisMemUsage += res.GetUsedMem()
				memUsage += res.GetUsedMem()
				resp.PerCpuUsage += res.GetPerUsedCpu()
			}

			if service.Type != 0 {
				resp.Service = append(resp.Service, service)
			}

		}(container)

	}

	wg.Wait()

	resp.MemUsage = FormatFileSize(memUsage)
	resp.MemTotal = FormatFileSize(memTotal)
	resp.PerCpuUsage = math.Ceil(resp.PerCpuUsage)
	resp.BasisPerCpuUsage = math.Ceil(basisPerCpuUsage)
	resp.ExtensionPerCpuUsage = math.Ceil(extensionPerCpuUsage)
	resp.PluginPerCpuUsage = math.Ceil(pluginPerCpuUsage)
	resp.BasisMemUsage = FormatFileSize(basisMemUsage)
	resp.ExtensionMemUsage = FormatFileSize(extensionMemUsage)
	resp.PluginMemUsage = FormatFileSize(pluginMemUsage)
}
