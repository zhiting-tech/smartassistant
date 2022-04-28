package entity

import (
	"encoding/json"

	"gorm.io/datatypes"

	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// 一个任务仅允许关联一个设备，对应的多个功能点配置；
// 或者
// 一个任务仅允许控制同一场景类型下的多个场景

type TaskType int

const (
	TaskTypeSmartDevice TaskType = iota + 1
	TaskTypeManualRun
	TaskTypeEnableAutoRun
	TaskTypeDisableAutoRun
)

// SceneTask 场景任务
type SceneTask struct {
	ID             int      `json:"id"`
	SceneID        int      `json:"scene_id"`
	ControlSceneID int      `json:"control_scene_id"` // ControlSceneID 控制场景id
	DelaySeconds   int      `json:"delay_seconds"`    // 延迟的秒数
	Type           TaskType `json:"type"`             // 任务目标：智能设备device或者是场景scene

	DeviceID   int            `json:"device_id"`
	Attributes datatypes.JSON `json:"attributes"` // refer to Attribute
}

func (t SceneTask) TableName() string {
	return "scene_tasks"
}

func GetSceneTasksBySceneID(sceneID int) (sceneTasks []SceneTask, err error) {
	err = GetDB().Order("type asc").Where("scene_id = ?", sceneID).Find(&sceneTasks).Error
	return
}

func CreateSceneTask(sceneTask []SceneTask) (err error) {
	err = GetDB().Create(&sceneTask).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
	return
}

// CheckTaskDevice 校验设备任务类型
func (t SceneTask) CheckTaskDevice(userId int) (err error) {
	if len(t.Attributes) == 0 || t.DeviceID == 0 {
		err = errors.Newf(status.SceneParamIncorrectErr, "scene_task_devices")
		return
	}

	var ds []Attribute
	if err = json.Unmarshal(t.Attributes, &ds); err != nil {
		logger.Error(err)
		return
	}
	up, err := GetUserPermissions(userId)
	if err != nil {
		return
	}
	for _, taskDevice := range ds {
		if !up.IsDeviceAttrControlPermit(t.DeviceID, taskDevice.AID) {
			err = errors.New(status.DeviceOrSceneControlDeny)
			return
		}
	}
	return
}

// CheckTaskType 执行任务类型校验
func (t SceneTask) CheckTaskType() (err error) {
	if t.Type < TaskTypeSmartDevice || t.Type > TaskTypeDisableAutoRun {
		err = errors.New(status.TaskTypeErr)
	}
	return
}
