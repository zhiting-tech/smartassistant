package entity

import (
	"encoding/json"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/attribute"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
	"gorm.io/datatypes"
)

type ConditionType int

const (
	ConditionTypeTiming ConditionType = iota + 1
	ConditionTypeDeviceStatus
)

type OperatorType string

const (
	OperatorGT OperatorType = ">"
	OperatorLT OperatorType = "<"
	OperatorEQ OperatorType = "="
)

// SceneCondition 场景条件
type SceneCondition struct {
	ID            int           `json:"id"`
	SceneID       int           `json:"scene_id"`
	ConditionType ConditionType `json:"condition_type"`
	TimingAt      time.Time     `json:"-"` // 定时在某个时间

	// 设备有关配置
	DeviceID      int            `json:"device_id"`      // 或某个设备状态变化时
	Operator      OperatorType   `json:"operator"`       // 操作符，大于、小于、等于
	ConditionAttr datatypes.JSON `json:"condition_attr"` // refer to Attribute
}

func (d SceneCondition) TableName() string {
	return "scene_conditions"
}

func GetConditionsBySceneID(sceneID int) (conditions []SceneCondition, err error) {
	err = GetDB().Where("scene_id = ?", sceneID).Find(&conditions).Error
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)

	}
	return
}

type ConditionInfo struct {
	SceneCondition
	Timing int64 `json:"timing"`
}

// CheckCondition 触发条件校验
func (c ConditionInfo) CheckCondition(userId int) (err error) {
	if err = c.checkConditionType(); err != nil {
		return
	}

	// 定时类型
	if c.ConditionType == ConditionTypeTiming {
		if err = c.checkConditionTypeTiming(); err != nil {
			return
		}
	} else {
		// 设备状态变化时
		if err = c.checkConditionDevice(userId); err != nil {
			return
		}
	}
	return
}

// checkConditionType 校验触发条件类型
func (c ConditionInfo) checkConditionType() (err error) {
	if c.ConditionType < ConditionTypeTiming || c.ConditionType > ConditionTypeDeviceStatus {
		err = errors.Newf(status.SceneParamIncorrectErr, "触发条件类型")
		return
	}
	return
}

// checkConditionTypeTiming 校验定时类型
func (c ConditionInfo) checkConditionTypeTiming() (err error) {
	if c.Timing == 0 || c.DeviceID != 0 {
		err = errors.New(status.ConditionMisMatchTypeAndConfigErr)
		return
	}
	return
}

// checkConditionDevice 校验设备类型
func (c ConditionInfo) checkConditionDevice(userId int) (err error) {
	if c.DeviceID <= 0 || c.Timing != 0 {
		err = errors.New(status.ConditionMisMatchTypeAndConfigErr)
		return
	}

	if err = c.CheckConditionItem(userId, c.DeviceID); err != nil {
		return
	}
	return
}

// CheckConditionItem 触发条件为设备状态变化时，校验对应参数
func (d SceneCondition) CheckConditionItem(userId, deviceId int) (err error) {
	if err = d.checkOperatorType(); err != nil {
		return
	}
	var item Attribute
	if err = json.Unmarshal(d.ConditionAttr, &item); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 如果设备的属性是只读，例如传感器，则默认有权限（这种属性，不需要权限控制）
	// TODO 这里信任了前端给予的数据，后期需要做调整
	if item.Permission == attribute.AttrPermissionOnlyRead {
		return
	}

	// 设备控制权限的判断
	if !IsDeviceControlPermit(userId, deviceId, item) {
		err = errors.New(status.DeviceOrSceneControlDeny)
		return
	}

	return
}

// checkOperatorType() 校验操作类型
func (d SceneCondition) checkOperatorType() (err error) {
	var opMap = map[OperatorType]bool{
		OperatorGT: true,
		OperatorLT: true,
		OperatorEQ: true,
	}

	if d.Operator != "" {
		if _, ok := opMap[d.Operator]; !ok {
			err = errors.Newf(status.SceneParamIncorrectErr, "设备操作符")
			return
		}
	}
	return
}

// GetScenesByCondition 根据条件获取场景
func GetScenesByCondition(deviceID int, attr Attribute) (scenes []Scene, err error) {
	conds, err := GetConditions(deviceID, attr)
	var sceneIDs []int
	for _, cond := range conds {
		sceneIDs = append(sceneIDs, cond.SceneID)
	}
	if len(sceneIDs) == 0 {
		return
	}
	if err = GetDB().Where("auto_run = true and id in (?)", sceneIDs).Find(&scenes).Error; err != nil {
		return
	}

	return
}

// GetConditions 获取符合设备属性的条件
func GetConditions(deviceID int, attr Attribute) (conds []SceneCondition, err error) {

	var (
		deviceConds []SceneCondition
	)

	attrQuery := datatypes.JSONQuery("condition_attr").
		Equals(attr.Attribute.Attribute, "attribute")
	instanceIDQuery := datatypes.JSONQuery("condition_attr").
		Equals(attr.InstanceID, "instance_id")

	if err = GetDB().Where("device_id=?", deviceID).
		Find(&deviceConds, attrQuery, instanceIDQuery).Error; err != nil {
		return
	}
	for _, cond := range deviceConds {
		if cond.ConditionType == ConditionTypeTiming {
			continue
		}

		var item Attribute
		if err := json.Unmarshal(cond.ConditionAttr, &item); err != nil {
			continue
		}

		if item.Operate(cond.Operator, attr.Val) {
			conds = append(conds, cond)
		}
	}
	return
}

type Attribute struct {
	server.Attribute
	InstanceID int `json:"instance_id"`
}

func (attr *Attribute) Operate(operatorType OperatorType, val interface{}) bool {
	switch operatorType {
	case OperatorEQ:
		return val == attr.Val
	case OperatorGT:
		switch val.(type) {
		case int:
			return val.(int) > attr.Val.(int)
		case float64:
			return val.(float64) > attr.Val.(float64)
		default:
			return false
		}
	case OperatorLT:
		switch val.(type) {
		case int:
			return val.(int) < attr.Val.(int)
		case float64:
			return val.(float64) < attr.Val.(float64)
		default:
			return false
		}
	}

	return false
}
