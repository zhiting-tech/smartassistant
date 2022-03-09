package plugin

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/zhiting-tech/smartassistant/modules/cloud"

	"io/fs"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin/docker"
)

type manager struct {
	areaID uint64
	docker *docker.Client
}

// GetPlugin 获取单个插件信息
func (m *manager) GetPlugin(id string) (p *Plugin, err error) {
	// TODO 从云端获取，失败则本地获取
	plg, err := cloud.GetPlugin(id)
	if err != nil {
		return
	}

	area, err := getCurrentArea()
	if err != nil {
		return
	}
	p = &Plugin{
		Config: Config{
			Name:    plg.Name,
			Version: plg.Version,
			Info:    plg.Intro,
		},
		ID:     plg.UID,
		Image:  plg.Image,
		Brand:  plg.Brand,
		Source: entity.SourceTypeDefault,
		AreaID: area.ID,
	}
	return
}

func NewManager() *manager {
	area, _ := getCurrentArea()
	return &manager{area.ID, docker.GetClient()}
}

// LoadPlugins 加载插件列表
func (m *manager) LoadPlugins() (plugins map[string]*Plugin, err error) {
	defaultPlugins, err := m.loadDefaultPlugins()
	if err != nil {
		return
	}
	plugins = make(map[string]*Plugin)
	for i, plg := range defaultPlugins {
		plugins[plg.ID] = &defaultPlugins[i]
	}
	return plugins, nil
}

// loadDefaultPlugins 加载插件列表
func (m *manager) loadDefaultPlugins() (plugins []Plugin, err error) {

	plgs, err := cloud.GetPlugins()
	if err != nil {
		return
	}

	for _, plg := range plgs {
		p := Plugin{
			Config: Config{
				Name:    plg.Name,
				Version: plg.Version,
				Info:    plg.Intro,
			},
			ID:     plg.UID,
			Image:  plg.Image,
			Brand:  plg.Brand,
			AreaID: m.areaID,
			Source: entity.SourceTypeDefault,
		}
		plugins = append(plugins, p)
	}
	return
}

// loadCustomPlugins 加载开发者插件列表
func (m *manager) loadCustomPlugins() (plugins []Plugin, err error) {
	customDir := "./plugins/"
	var localPluginFiles []fs.FileInfo
	localPluginFiles, err = ioutil.ReadDir(customDir)
	if err != nil {
		return
	}
	for _, fileInfo := range localPluginFiles {
		if !fileInfo.IsDir() {
			continue
		}
		var plg Plugin

		plg, err = m.loadCustomPlugin(customDir + fileInfo.Name())
		if err != nil {
			return
		}
		plugins = append(plugins, plg)
	}
	return
}

// loadCustomPlugin 加载开发者插件
func (m *manager) loadCustomPlugin(path string) (plg Plugin, err error) {
	configPath := filepath.Join(path, "config.json")
	plgFile, err := os.Open(configPath)
	if err != nil {
		return
	}
	defer plgFile.Close()

	data, err := ioutil.ReadAll(plgFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &plg)
	return
}
