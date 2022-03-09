package plugin

import (
	"encoding/json"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
	"github.com/zhiting-tech/smartassistant/pkg/archive"
)

// LoadPluginFromZip 从压缩包中加载插件
func LoadPluginFromZip(path string, areaID uint64) (plg Plugin, err error) {

	dstDir := filepath.Dir(path)
	// unzip file
	if err = archive.UnZip(dstDir, path); err != nil {
		return
	}

	logrus.Debug(dstDir)
	dstDir, _ = filepath.Abs(dstDir)
	logrus.Debug(dstDir)
	pluginPath := pluginBasePath(dstDir)
	plgConf, err := LoadPluginConfig(pluginPath)
	if err != nil {
		return
	}

	// save plugin info
	data, _ := json.Marshal(plgConf)
	pi := entity.PluginInfo{
		AreaID:    areaID,
		Image:     plgConf.ID(),
		Info:      plgConf.Info,
		PluginID:  plgConf.ID(),
		ConfigMsg: data,
		Version:   plgConf.Version,
		Source:    entity.SourceTypeDevelopment,
	}
	if err = entity.SavePluginInfo(pi); err != nil {
		return
	}

	// docker build
	go func() {
		var err error
		defer func() {
			os.RemoveAll(dstDir)
			var status = entity.StatusInstallSuccess
			var errInfo string
			if err != nil {
				status = entity.StatusInstallFail
				errInfo = err.Error()
			}
			if uerr := entity.UpdatePluginInfo(plgConf.ID(), entity.PluginInfo{Status: status, ErrorInfo: errInfo}); uerr != nil {
				logrus.Errorf("UpdatePluginStatus err: %s", uerr.Error())
			}
		}()

		_, err = BuildFromDir(pluginPath, plgConf.ID())
		if err != nil {
			logrus.Errorf("build image err: %v\n", err)
			return
		}

		plg = NewFromEntity(pi)
		if err = plg.Up(); err != nil {
			logrus.Errorf("up image err: %v\n", err)
			return
		}
		logrus.Println("image build success")
	}()

	return
}

// LoadPluginConfig 加载插件配置
func LoadPluginConfig(path string) (plg Config, err error) {

	configFile, err := os.Open(path + "/config.json")
	if err != nil {
		return
	}
	defer configFile.Close()

	data, err := ioutil.ReadAll(configFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &plg)

	if err = plg.Validate(); err != nil {
		return
	}
	return
}

// pluginBasePath 根据配置文件config.json确定插件包准确目录
func pluginBasePath(path string) (plgPath string) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if filepath.Base(path) == "config.json" {
			plgPath = filepath.Dir(path)
		}
		return nil
	})
	return
}

// BuildFromDir 从源码编译镜像
func BuildFromDir(path, tag string) (imageID string, err error) {
	c := docker.GetClient()
	return c.BuildFromPath(path, tag)
}

// BuildFromTar 从源码tar压缩包中build镜像
func BuildFromTar(tarPath string) (imageID string, err error) {

	tar, err := os.Open(tarPath)
	if err != nil {
		return
	}
	defer tar.Close()

	c := docker.GetClient()
	return c.BuildFromTar(tar, "")
}
