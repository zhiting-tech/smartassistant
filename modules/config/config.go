// Package config 配置模块，由程序入口加载，全局可用
package config

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	mrand "math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/rand"
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

func InitSAIDAndSAKeyIfEmpty(cfg string) {
	if options.SmartAssistant.ID != "" && options.SmartAssistant.Key != "" {
		return
	}

	var id, key string
	var err error

	// 允许重试，目前不重试获取
	for i := 0; i < 1; i++ {
		if id, key, err = cloudRandomSAIDAndSAKey(); err == nil {
			break
		}

		logger.Infof("request said and key error:%v", err)
	}

	if id == "" || key == "" {
		id, key = generateRandomSAIDAndSAKey()
		logger.Infof("generate local said and key, said:%s, sakey:%s", id, key)
	}

	options.SmartAssistant.ID = id
	options.SmartAssistant.Key = key

	var data []byte
	if strings.HasSuffix(cfg, ".json") {
		if data, err = jsoniter.Marshal(&options); err != nil {
			logger.Panic(fmt.Sprintf("marshal conf file:%s error:%v", cfg, err.Error()))
		}
	} else if strings.HasSuffix(cfg, ".yaml") {
		if data, err = yaml.Marshal(&options); err != nil {
			logger.Panic(fmt.Sprintf("marshal conf file:%s error:%v", cfg, err.Error()))
		}
	}

	// 写入新的配置文件
	var f *os.File
	defer f.Close()
	if f, err = ioutil.TempFile(filepath.Dir(cfg), fmt.Sprintf("*.%s", filepath.Base(cfg))); err != nil {
		logger.Panicf("new temp file error:%v", err)
	}

	if _, err = f.Write(data); err != nil {
		logger.Panicf("write file error:%v", err)
	}

	if err = f.Sync(); err != nil {
		logger.Panicf("sync file error:%v", err)
	}

	if err = os.Rename(f.Name(), cfg); err != nil {
		logger.Panicf("rename file error:%v", err)
	}

}

func generateRandomSAIDAndSAKey() (string, string) {
	// 本地生成的SAID必须以22开头
	raw := make([]byte, 6)
	raw[0] = 0x22
	raw[1] = 0x01
	binary.BigEndian.PutUint32(raw[2:], mrand.Uint32())
	id := hex.EncodeToString(raw)
	key := rand.StringK(14, rand.KindAll)

	return id, key
}

func cloudRandomSAIDAndSAKey() (string, string, error) {
	url := fmt.Sprintf("%s/sa/new", options.SmartCloud.URL())
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	cloudRandomSAIDAndKeyResp := struct {
		Status int
		Reason string
		Data   struct {
			SAID  string
			SAKey string
		}
	}{}
	if err = jsoniter.Unmarshal(data, &cloudRandomSAIDAndKeyResp); err != nil {
		return "", "", err
	}

	if cloudRandomSAIDAndKeyResp.Status != 0 {
		return "", "", fmt.Errorf(cloudRandomSAIDAndKeyResp.Reason)
	}

	return cloudRandomSAIDAndKeyResp.Data.SAID, cloudRandomSAIDAndKeyResp.Data.SAKey, nil
}
