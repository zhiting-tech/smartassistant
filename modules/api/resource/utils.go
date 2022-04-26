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
		return Decimal((cpuDelta / systemCpuDelta) * float64(sr.CpuStats.OnlineCpus) * 100)
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
