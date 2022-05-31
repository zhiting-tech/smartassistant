package resource

import (
	"fmt"
	"math"
	"strconv"
)

type SystemResource struct {
	Id          string      `json:"id"`
	CpuStats    CpuStats    `json:"cpu_stats"`
	PrecpuStats CpuStats    `json:"precpu_stats"`
	MemoryStats MemoryStats `json:"memory_stats"`
}

type CpuStats struct {
	CpuUsage       CpuUsage `json:"cpu_usage"`
	SystemCpuUsage uint64   `json:"system_cpu_usage"`
	OnlineCpus     int      `json:"online_cpus"`
}

type CpuUsage struct {
	TotalUsage  uint64   `json:"total_usage"`
	PercpuUsage []uint64 `json:"percpu_usage"`
}

type MemoryStats struct {
	Usage    uint64 `json:"usage"`
	MaxUsage uint64 `json:"max_usage"`
	Stats    Stats  `json:"stats"`
	Limit    uint64 `json:"limit"`
}

type Stats struct {
	Cache uint64 `json:"cache"`
}

// GetUsedMem 获取已用内存
func (sr SystemResource) GetUsedMem() uint64 {
	return sr.MemoryStats.Usage - sr.MemoryStats.Stats.Cache
}

// GetPerUsedMem 获取百分比已用内存，保留两位小数
func (sr SystemResource) GetPerUsedMem() float64 {
	if sr.MemoryStats.Limit > 0 {
		return Decimal((float64(sr.GetUsedMem()) / float64(sr.MemoryStats.Limit)) * 100)
	} else {
		return 0
	}
}

// GetPerUsedCpu 获取cpu使用率,保留两位小数
func (sr SystemResource) GetPerUsedCpu() float64 {
	cpuDelta := float64(sr.CpuStats.CpuUsage.TotalUsage - sr.PrecpuStats.CpuUsage.TotalUsage)
	systemCpuDelta := float64(sr.CpuStats.SystemCpuUsage - sr.PrecpuStats.SystemCpuUsage)
	if systemCpuDelta > 0 {
		return Decimal((cpuDelta / systemCpuDelta) * 100)
	} else {
		return 0
	}

}

// Decimal 保留两位小数
func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// FormatFileSize 文件大小格式换算
func FormatFileSize(size uint64) string {
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%vMB", math.Ceil(float64(size)/float64(1024*1024)))
	} else if size < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%vGB", math.Ceil(float64(size)/float64(1024*1024*1024)))
	} else {
		return fmt.Sprintf("%vTB", math.Ceil(float64(size)/float64(1024*1024*1024*1024)))
	}
}

// FormatTimeSize 时间间隔大小格式换算
func FormatTimeSize(time int64) string {
	if time < 60*60 {
		return fmt.Sprintf("%v分钟", math.Floor(float64(time)/float64(60)))
	} else if time < 60*60*24 {
		return fmt.Sprintf("%v小时", math.Floor(float64(time)/float64(60*60)))
	} else {
		return fmt.Sprintf("%v天", math.Floor(float64(time)/float64(60*60*24)))
	}
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

type Services []Service

// Len 实现sort.Interface接口取元素数量方法
func (s Services) Len() int {
	return len(s)
}

// Less 实现sort.Interface接口比较元素方法,不能重启的排前面,按name排序
func (s Services) Less(i, j int) bool {
	_, b1 := unRestartList[s[i].Name]
	_, b2 := unRestartList[s[j].Name]
	if b1 && b2 {
		return s[i].Name < s[j].Name
	} else if b1 && !b2 {
		return true
	} else if !b1 && b2 {
		return false
	}
	// 默认升序排列
	return s[i].ServiceName < s[j].ServiceName
}

// Swap 实现sort.Interface接口交换元素方法
func (s Services) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
