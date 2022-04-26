package config

import (
	"fmt"
	"path/filepath"
)

type SmartAssistant struct {
	ID       string   `json:"id" yaml:"id"`
	Key      string   `json:"key" yaml:"key"`
	Db       string   `json:"db" yaml:"db"`
	Host     string   `json:"host" yaml:"host"`
	Port     int      `json:"port" yaml:"port"`
	GRPCPort int      `json:"grpc_port" yaml:"grpc_port"`
	Database Database `json:"database" yaml:"database"`
	// HostDataPath 宿主机runtime目录
	HostRuntimePath string `json:"host_runtime_path" yaml:"host_runtime_path"`
	RuntimePath     string `json:"runtime_path" yaml:"runtime_path"`

	DockerRegistry string `json:"docker_registry" yaml:"docker_registry"`

	// Deprecated: HostIP 插件取消host模式后删除
	HostIP string `json:"host_ip" yaml:"host_ip"`
}

type Database struct {
	Driver   string `json:"driver" yaml:"driver"`
	Name     string `json:"name" yaml:"name"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
}

func (sa SmartAssistant) HttpAddress() string {
	return fmt.Sprintf("%s:%d", sa.Host, sa.Port)
}

func (sa SmartAssistant) GRPCAddress() string {
	return fmt.Sprintf("%s:%d", sa.Host, sa.GRPCPort)
}

func (sa SmartAssistant) BackupPath() string {
	return filepath.Join(sa.RuntimePath, "backup")
}

func (sa SmartAssistant) DataPath() string {
	return filepath.Join(sa.RuntimePath, "data")
}

func (sa SmartAssistant) VolumePath() string {
	return filepath.Join(sa.RuntimePath, "volume")
}