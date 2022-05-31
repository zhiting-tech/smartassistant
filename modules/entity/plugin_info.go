package entity

import (
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
)

const (
	// StatusInstallFail 添加插件失败
	StatusInstallFail = -1

	StatusInstalling = 0

	// StatusInstallSuccess 添加插件成功
	StatusInstallSuccess = 1
)

const (
	SourceTypeDefault     = "default"     // 默认插件
	SourceTypeDevelopment = "development" // 开发者插件
)

// PluginInfo 开发者插件信息
type PluginInfo struct {
	ID        int
	Name      string
	AreaID    uint64 `gorm:"uniqueIndex:area_plugin"`
	Area      Area   `gorm:"constraint:OnDelete:CASCADE;"`
	PluginID  string `gorm:"uniqueIndex:area_plugin"`
	Image     string
	Info      string
	ConfigMsg datatypes.JSON
	Status    int
	Version   string
	Source    string
	Brand     string
	ErrorInfo string
}

func (p PluginInfo) TableName() string {
	return "plugin_infos"
}

func SavePluginInfo(pi PluginInfo) (err error) {
	return GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "area_id"}, {Name: "plugin_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"info", "version", "image"}),
	}).Create(&pi).Error
}

// UpdatePluginInfo 更新插件
func UpdatePluginInfo(pluginID string, pluginInfo PluginInfo) (err error) {
	return GetDB().Where(PluginInfo{PluginID: pluginID}).Updates(pluginInfo).Error
}

func IsPluginAdd(pluginID string, areaID uint64) bool {
	pluginInfo, err := GetPlugin(pluginID, areaID)
	if err != nil {
		return false
	}
	return pluginInfo.Status == StatusInstallSuccess
}

// GetInstalledPlugins 获取所有已安装插件
func GetInstalledPlugins() (pis []PluginInfo, err error) {
	err = GetDB().Where(PluginInfo{Status: StatusInstallSuccess}).Find(&pis).Error
	return
}

// GetDevelopPlugins 获取所有开发插件
func GetDevelopPlugins(areaID uint64) (pis []PluginInfo, err error) {
	err = GetDB().Where(PluginInfo{AreaID: areaID, Source: SourceTypeDevelopment}).Find(&pis).Error
	return
}

func GetPlugin(pluginID string, areaID uint64) (plugin PluginInfo, err error) {
	if err = GetDB().Where(PluginInfo{PluginID: pluginID, AreaID: areaID}).
		First(&plugin).Error; err != nil {
		return
	}
	return
}

// DelPlugin 删除插件
func DelPlugin(pluginID string, areaID uint64) (err error) {
	err = GetDB().Where(PluginInfo{PluginID: pluginID, AreaID: areaID}).
		Delete(&PluginInfo{}).Error
	return
}

// IsPluginDevelop 是否是开发插件
func IsPluginDevelop(pluginID string, areaID uint64) bool {
	p, _ := GetPlugin(pluginID, areaID)
	return p.Source == SourceTypeDevelopment
}
