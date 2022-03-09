package entity

import (
	"encoding/json"
	"errors"
	errors2 "github.com/zhiting-tech/smartassistant/pkg/errors"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 默认配置表
var (
	defaultSettingMap = map[string]interface{}{
		UserCredentialFoundType: defaultUserCredentialFoundSetting,
		LogSwitch:               defaultLogSetting,
		GetCloudDiskCredential:  defaultGetCloudDiskCredentialSetting,
	}
)

// 配置类型
const (
	UserCredentialFoundType = "user_credential_found"
	LogSwitch               = "log_switch"
	GetCloudDiskCredential  = "get_cloud_disk_credential"
)

// 默认配置项
var (
	defaultUserCredentialFoundSetting    = UserCredentialFoundSetting{}
	defaultLogSetting                    = LogSwitchSetting{}
	defaultGetCloudDiskCredentialSetting = GetCloudDiskCredentialSetting{}
)

// 用户凭证配置
type UserCredentialFoundSetting struct {
	UserCredentialFound bool `json:"user_credential_found"`
}

// 日志开关
type LogSwitchSetting struct {
	LogSwitch bool `json:"log_switch"`
}

type GetCloudDiskCredentialSetting struct {
	GetCloudDiskCredentialSetting bool `json:"get_cloud_disk_credential_setting"`
}

// GlobalSetting SA全局设置
type GlobalSetting struct {
	ID      int
	Type    string `gorm:"uniqueIndex:area_id_type"`
	Value   datatypes.JSON
	AreaID  uint64 `gorm:"type:bigint;uniqueIndex:area_id_type"`
	Area    Area   `gorm:"constraint:OnDelete:CASCADE;"`
	Deleted gorm.DeletedAt
}

func (g GlobalSetting) TableName() string {
	return "global_setting"
}

// GetSetting 获取全局设置
func GetSetting(settingType string, setting interface{}, areaID uint64) (err error) {
	var gs GlobalSetting
	err = GetDB().Where(GlobalSetting{Type: settingType, AreaID: areaID}).First(&gs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		return
	}

	return json.Unmarshal(gs.Value, setting)
}

func GetAllSetting(areaID uint64) (gsList []GlobalSetting, err error) {
	if err = GetDB().Where(GlobalSetting{AreaID: areaID}).Find(&gsList).Error; err != nil {
		err = errors2.Wrap(err, errors2.InternalServerErr)
		return
	}
	return
}

// UpdateSetting 添加全局设置
func UpdateSetting(settingType string, setting interface{}, areaID uint64) (err error) {

	v, err := json.Marshal(setting)
	if err != nil {
		return
	}

	s := GlobalSetting{
		Type:   settingType,
		Value:  v,
		AreaID: areaID,
	}
	return GetDB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}, {Name: "area_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&s).Error
}

// GetDefaultUserCredentialFoundSetting 获取找回用户凭证默认配置
func GetDefaultUserCredentialFoundSetting() UserCredentialFoundSetting {
	return defaultSettingMap[UserCredentialFoundType].(UserCredentialFoundSetting)
}

// GetDefaultLogSetting 获取日志开关默认配置
func GetDefaultLogSetting() LogSwitchSetting {
	return defaultSettingMap[LogSwitch].(LogSwitchSetting)
}

func GetDefaultCloudDiskCredentialSetting() GetCloudDiskCredentialSetting {
	return defaultSettingMap[GetCloudDiskCredential].(GetCloudDiskCredentialSetting)
}
