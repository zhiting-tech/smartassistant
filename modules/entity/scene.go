package entity

import (
	errors2 "errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zhiting-tech/smartassistant/modules/types/status"

	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type TimePeriodType int

const (
	TimePeriodTypeAllDay TimePeriodType = iota + 1
	TimePeriodTypeCustom
)

type RepeatType int

const (
	RepeatTypeAllDay RepeatType = iota + 1
	RepeatTypeWorkDay
	RepeatTypeCustom
)

const (
	MatchAllCondition = 1 // 全部满足
	MatchAnyCondition = 2 // 任一满足
)

// Scene 场景
type Scene struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	ConditionLogic int    `json:"condition_logic"`       // 1 为 全部满足，2为满足任一
	Sort           int    `json:"sort" gorm:"default:0"` // 排序

	// 生效时间的配置
	TimePeriodType TimePeriodType `json:"time_period"` // 全天1、时间段2
	EffectStart    time.Time      `json:"-"`
	EffectEnd      time.Time      `json:"-"`

	// 重复执行的配置
	RepeatType RepeatType `json:"repeat_type"` // 每天1，工作日2，自定义3
	RepeatDate string     `json:"repeat_date"` // 自定义的情况下：1234567

	// 设置为手动：false，则不能再设置其他两种
	AutoRun bool `json:"auto_run"` // true 就需要设置scene_condition，false 只需表示手动
	// 场景会自动执行: true
	IsOn bool `json:"is_on"`

	CreatorID       int              `json:"creator_id"`
	CreatedAt       time.Time        `json:"-"`
	SceneConditions []SceneCondition `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	SceneTasks      []SceneTask      `json:"scene_tasks" gorm:"constraint:OnDelete:CASCADE;"`
	Deleted         gorm.DeletedAt   `json:"-"`

	AreaID uint64 `gorm:"type:bigint;index"`
	Area   Area   `gorm:"constraint:OnDelete:CASCADE;"`

	Version int `json:"-" gorm:"column:version;default:1;NOT NULL"`
}

func (s Scene) TableName() string {
	return "scenes"
}

const (
	sceneNameMinLength = 1
	sceneNameMaxLength = 40
)

func (s *Scene) BeforeSave(tx *gorm.DB) (err error) {
	if err = s.CheckRepeatConfig(); err != nil {
		return
	}

	s.RepeatDate = strings.TrimSpace(s.RepeatDate)

	if err = s.CheckSceneNameLength(); err != nil {
		return
	}

	return nil
}

// HaveTimeCondition 场景是否有定时条件
func (s Scene) HaveTimeCondition() bool {
	for _, c := range s.SceneConditions {
		if c.ConditionType == ConditionTypeTiming {
			return true
		}
	}
	return false
}

func (s Scene) IsMatchAllCondition() bool {
	return s.ConditionLogic == MatchAllCondition
}

func GetScenes(areaID uint64) (scenes []Scene, err error) {
	err = GetDBWithAreaScope(areaID).Order("sort asc,id desc").Find(&scenes).Error
	return
}

func CreateScene(scene *Scene) (err error) {
	if err = GetDB().Create(scene).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

func GetSceneById(id int) (scene Scene, err error) {
	if err = GetDB().Where("id=?", id).First(&scene).Error; err != nil {
		return
	}
	return
}

// GetSceneInfoById 获取场景所有信息
func GetSceneInfoById(id int) (scene Scene, err error) {
	if err = GetDB().
		Preload("SceneConditions").Preload("SceneTasks").
		First(&scene, id).Error; err != nil {
		return
	}
	return
}

// GetSceneByIDWithUnscoped 获取场景，包括已删除
func GetSceneByIDWithUnscoped(id int) (scene Scene, err error) {
	err = GetDB().Unscoped().
		Preload("SceneConditions").
		First(&scene, id).Error
	return
}

func IsSceneNameExist(name string, sceneId int, areaId uint64) (err error) {
	var db *gorm.DB
	if sceneId != 0 {
		db = GetDB().Where("id != ? and name = ? and area_id = ?", sceneId, name, areaId)
	} else {
		db = GetDB().Where("name = ? and area_id = ?", name, areaId)
	}

	err = db.First(&Scene{}).Error
	if err != nil && errors2.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	err = errors.New(status.SceneNameExist)
	return err
}

// CheckRepeatConfig 校验重复执行的配置
func (s Scene) CheckRepeatConfig() (err error) {
	if s.AutoRun {
		if s.RepeatType < RepeatTypeAllDay || s.RepeatType > RepeatTypeCustom {
			err = errors.Newf(status.SceneParamIncorrectErr, "重复执行配置")
			return
		}

		if !CheckIllegalRepeatDate(s.RepeatDate) {
			err = errors.Newf(status.SceneParamIncorrectErr, "重复生效时间")
			return
		}

	}
	return
}

func (s Scene) CheckSceneNameLength() (err error) {
	if s.Name == "" || utf8.RuneCountInString(s.Name) < sceneNameMinLength || utf8.RuneCountInString(s.Name) > sceneNameMaxLength {
		err = errors.New(errors.BadRequest)
		return
	}
	return
}

func CheckSceneExitById(sceneId int) (err error) {
	_, err = GetSceneById(sceneId)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New(status.SceneNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}
	return
}

// CheckConditionLogic 校验满足条件
func (s Scene) CheckConditionLogic() bool {
	return !s.IsMatchAllCondition() && s.ConditionLogic != MatchAnyCondition

}

// CheckPeriodType 生效时间类型校验
func (s Scene) CheckPeriodType() (err error) {
	if s.TimePeriodType < TimePeriodTypeAllDay || s.TimePeriodType > TimePeriodTypeCustom {
		err = errors.Newf(status.SceneParamIncorrectErr, "生效时间类型")
		return
	}
	return
}

func DeleteScene(sceneId int) (err error) {
	s := Scene{ID: sceneId}
	err = GetDB().First(&s).Delete(&s).Error
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.SceneNotExist)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
	}
	return
}

// SwitchAutoScene 切换自动场景开关
func SwitchAutoScene(scene *Scene, isExecute bool) error {

	updateMap := map[string]interface{}{
		"is_on": isExecute,
	}
	if err := GetDB().Model(&scene).Updates(&updateMap).Error; err != nil {
		return errors.Wrap(err, errors.InternalServerErr)
	}
	return nil
}

// SwitchAutoSceneByID 切换自动场景开关
func SwitchAutoSceneByID(sceneID int, isExecute bool) error {

	s, err := GetSceneInfoById(sceneID)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(status.SceneNotExist)
		}
		return errors.Wrap(err, errors.InternalServerErr)
	}
	return SwitchAutoScene(&s, isExecute)
}

// GetPendingScenesByTime 根据时间获取待执行的场景
func GetPendingScenesByTime(t time.Time) (scenes []Scene, err error) {
	weekDay := strconv.Itoa(int(t.Weekday()))
	if err = GetDB().Where("auto_run=? and is_on=? and repeat_date like ?", true, true, "%"+weekDay+"%").
		Preload("SceneConditions").
		Find(&scenes).Error; err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return
}

// UpdateSceneByIDWithTx 根据SceneID更新场景
func UpdateSceneByIDWithTx(sceneID int, update *Scene, tx *gorm.DB) (err error) {
	// 用户编辑场景时更新场景版本, 旧版本场景触发的场景任务不执行
	if err = tx.Omit("version", "area_id").Where("id=?", sceneID).Updates(update).Error; err != nil {
		return
	}
	err = tx.Model(&Scene{}).Where("id=?", sceneID).UpdateColumn("version", gorm.Expr("version+1")).Error
	return
}

// UpdateSceneSort 修改场景的排序
func UpdateSceneSort(tx *gorm.DB, id int, sort int, areaID uint64) (err error) {
	err = tx.First(&Scene{}, "id=? and area_id=?", id, areaID).
		Update("sort", sort).Error
	return
}
