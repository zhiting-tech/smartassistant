package config

import (
	"github.com/zhiting-tech/smartassistant/config"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gopkg.in/yaml.v2"
)

func defaultOptions() (option Options) {
	err := yaml.Unmarshal(config.DefaultSmartassistantConfigData, &option)
	if err != nil {
		logger.Panic(err)
	}

	return
}

type Options struct {
	Debug          bool           `json:"debug" yaml:"debug"`
	SmartCloud     SmartCloud     `json:"smartcloud" yaml:"smartcloud"`
	SmartAssistant SmartAssistant `json:"smartassistant" yaml:"smartassistant"`
	Datatunnel     Datatunnel     `json:"datatunnel" yaml:"datatunnel"`
	Extension      Extension      `json:"extension" yaml:"extension"`
}
