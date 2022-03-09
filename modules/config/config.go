// Package config 配置模块，由程序入口加载，全局可用
package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gopkg.in/yaml.v2"
)

var (
	options           = defaultOptions()
	alreadyInitConfig bool
)

func GetConf() *Options {
	if !alreadyInitConfig {
		panic("please call InitConfig first")
	}
	return &options
}

func InitConfig(fn string) *Options {
	fd, err := os.Open(fn)
	if err != nil {
		logger.Panic(fmt.Sprintf("open conf file:%s error:%v", fn, err.Error()))
	}
	defer fd.Close()

	content, err := ioutil.ReadAll(fd)
	if err != nil {
		logger.Panic(fmt.Sprintf("read conf file:%s error:%v", fn, err.Error()))
	}

	if strings.HasSuffix(fn, ".json") {
		if err = jsoniter.Unmarshal(content, &options); err != nil {
			logger.Panic(fmt.Sprintf("unmarshal conf file:%s error:%v", fn, err.Error()))
		}
	} else if strings.HasSuffix(fn, ".yaml") {
		if err = yaml.Unmarshal(content, &options); err != nil {
			logger.Panic(fmt.Sprintf("unmarshal conf file:%s error:%v", fn, err.Error()))
		}
	}
	alreadyInitConfig = true
	return &options
}
