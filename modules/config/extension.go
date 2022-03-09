package config

import "fmt"

type Extension struct {
	GRPCPort int `json:"grpc_port" yaml:"grpc_port"`
}

func (e Extension) GRPCAddress() string {
	return fmt.Sprintf("%s:%d", GetConf().SmartAssistant.Host, e.GRPCPort)
}
