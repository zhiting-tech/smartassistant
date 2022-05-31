package task

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/entity"
)

func addDevice() *entity.Device {
	d := &entity.Device{
		Name:         "testing device",
		Manufacturer: "testing",
		PluginID:     "testing",
		CreatedAt:    time.Now(),
	}
	_ = entity.GetDB().Transaction(func(tx *gorm.DB) error {
		return entity.CreateDevice(d, tx)
	})
	return d
}

func addDeviceManualScene(d *entity.Device, name string) *entity.Scene {
	var manualScene = &entity.Scene{
		Name:      name,
		AutoRun:   false,
		CreatorID: 1,
		CreatedAt: time.Now(),
		SceneTasks: []entity.SceneTask{
			{
				ControlSceneID: 0,
				DelaySeconds:   2,
				Type:           entity.TaskTypeSmartDevice,
				DeviceID:       d.ID,
			},
		},
	}
	db := entity.GetDB().Session(&gorm.Session{FullSaveAssociations: true}).Model(entity.Scene{})
	db.Create(manualScene)
	return manualScene
}

func addAutoScene(d *entity.Device, s *entity.Scene, name string) *entity.Scene {
	var autoScene = &entity.Scene{
		Name:           name,
		AutoRun:        true,
		ConditionLogic: entity.MatchAllCondition,
		RepeatType:     entity.RepeatTypeAllDay,
		RepeatDate:     "1234567",
		SceneConditions: []entity.SceneCondition{
			{
				ConditionType: entity.ConditionTypeTiming,
				TimingAt:      time.Now().Add(3 * time.Second),
				DeviceID:      d.ID,
			},
		},
		IsOn:           true,
		TimePeriodType: entity.TimePeriodTypeAllDay,
		CreatorID:      1,
		CreatedAt:      time.Now(),

		SceneTasks: []entity.SceneTask{
			{
				ControlSceneID: 0,
				DelaySeconds:   2,
				Type:           entity.TaskTypeSmartDevice,
				DeviceID:       d.ID,
			},
			{
				ControlSceneID: s.ID,
				DelaySeconds:   3,
				Type:           entity.TaskTypeManualRun,
			},
		},
	}
	db := entity.GetDB().Session(&gorm.Session{FullSaveAssociations: true}).Model(entity.Scene{})
	db.Create(autoScene)
	return autoScene
}

func TestDeviceManualScene(t *testing.T) {
	name := "test_device_manual_scene"
	d := addDevice()
	s := addDeviceManualScene(d, name)
	GetManager().(*LocalManager).setSceneOn(s.ID)
	time.Sleep(3 * time.Second)
	var taskLogs []entity.TaskLog
	err := entity.GetDB().Where("name=?", name).Find(&taskLogs).Error
	assert.Nil(t, err)
	assert.NotEmpty(t, len(taskLogs), "task log not found")
}

func TestDeviceAutoScene(t *testing.T) {
	mName := "test_device_manual_scene"
	aName := "test_device_auto_scene"
	d := addDevice()
	s := addDeviceManualScene(d, mName)
	as := addAutoScene(d, s, aName)
	GetManager().AddSceneTask(*as)
	time.Sleep(15 * time.Second)
	var taskLogs []entity.TaskLog
	err := entity.GetDB().Where("name=?", mName).Find(&taskLogs).Error
	assert.Nil(t, err)
	assert.NotEmpty(t, len(taskLogs), "manual task log not found")
	err = entity.GetDB().Where("name=?", aName).Find(&taskLogs).Error
	assert.Nil(t, err)
	assert.NotEmpty(t, len(taskLogs), "auto task log not found")
}
