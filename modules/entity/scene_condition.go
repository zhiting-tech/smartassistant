package entity

import (
	"encoding/json"
	"time"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"

	"gorm.io/datatypes"

	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type ConditionType int

const (
	ConditionTypeTiming ConditionType = iota + 1
	ConditionTypeDeviceStatus
)

type OperatorType string

const (
	OperatorGT      OperatorType = ">"
	OperatorGTE     OperatorType = ">="
	OperatorLT      OperatorType = "<"
	OperatorLTE     OperatorType = "<="
	OperatorEQ      OperatorType = "="
	OperatorBetween OperatorType = "between"
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
func (c ConditionInfo) CheckCondition(userId int, isRequireNotify bool) (err error) {
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
		if err = c.checkConditionDevice(userId, isRequireNotify); err != nil {
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
func (c ConditionInfo) checkConditionDevice(userId int, isRequireNotify bool) (err error) {
	if c.DeviceID <= 0 || c.Timing != 0 {
		err = errors.New(status.ConditionMisMatchTypeAndConfigErr)
		return
	}

	if err = c.CheckConditionItem(userId, c.DeviceID, isRequireNotify); err != nil {
		return
	}
	return
}

// CheckConditionItem 触发条件为设备状态变化时，校验对应参数
func (d SceneCondition) CheckConditionItem(userId, deviceId int, isRequireNotify bool) (err error) {
	if err = d.checkOperatorType(); err != nil {
		return
	}
	var item Attribute
	if err = json.Unmarshal(d.ConditionAttr, &item); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	item.Permission = GetPermission(deviceId, item.AID)
	// 通过属性通知触发的场景任务, 需要属性具有通知权限
	if isRequireNotify && !item.PermissionNotify() {
		err = errors.Newf(status.ConditionOfDeviceAttrWithoutNotifyPermission)
		return
	}
	// 通过定时查询属性状态触发的场景任务, 需要属性具有读权限
	if !isRequireNotify && !item.PermissionRead() {
		err = errors.Newf(status.ConditionOfDeviceAttrWithoutReadPermission)
		return
	}

	// 设备控制权限的判断
	if !IsDeviceControlPermit(userId, deviceId, item) {
		err = errors.New(status.DeviceOrSceneControlDeny)
		return
	}

	return
}

// GetPermission 获取属性的权限
func GetPermission(deviceId int, aid int) (permission uint) {
	device, err := GetDeviceByID(deviceId)
	if err != nil {
		return
	}

	das, err := device.GetThingModel()
	if err != nil {
		return
	}
	for _, instance := range das.Instances {
		if instance.IID != device.IID {
			continue
		}
		for _, srv := range instance.Services {
			for _, attr := range srv.Attributes {
				if attr.AID != aid {
					continue
				}
				permission = attr.Permission
				return
			}
		}
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
func GetScenesByCondition(deviceID int, attr definer.AttributeEvent) (scenes []Scene, err error) {
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
func GetConditions(deviceID int, ae definer.AttributeEvent) (conds []SceneCondition, err error) {

	var (
		deviceConds []SceneCondition
	)

	attrQuery := datatypes.JSONQuery("condition_attr").
		Equals(ae.AID, "aid")

	if err = GetDB().Where("device_id=?", deviceID).
		Find(&deviceConds, attrQuery).Error; err != nil {
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

		if item.Operate(cond.Operator, ae.Val) {
			conds = append(conds, cond)
		}
	}
	return
}

type Attribute struct {
	ServiceType thingmodel.ServiceType `json:"service_type"`
	thingmodel.Attribute
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
