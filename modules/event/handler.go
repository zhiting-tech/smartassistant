package event

import (
	"encoding/json"
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/task"
	"github.com/zhiting-tech/smartassistant/modules/websocket"
	_ "github.com/zhiting-tech/smartassistant/modules/websocket"
	"github.com/zhiting-tech/smartassistant/pkg/event"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2/definer"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

func RegisterEventFunc(ws *websocket.Server) {
	event.RegisterEvent(event.AttributeChange, ws.BroadcastMsg,
		UpdateDeviceShadowBeforeExecuteTask, RecordDeviceState)
	event.RegisterEvent(event.DeviceDecrease, ws.BroadcastMsg)
	event.RegisterEvent(event.DeviceIncrease, ws.BroadcastMsg)
	event.RegisterEvent(event.ThingModelChange, UpdateThingModel)
}

func UpdateThingModel(em event.EventMessage) (err error) {
	tm := em.Param["thing_model"].(thingmodel.ThingModel)
	areaID := em.Param["area_id"].(uint64)
	pluginID := em.Param["plugin_id"].(string)
	iid := em.Param["iid"].(string)

	// 网关未添加则设备不更新
	if _, err = entity.GetPluginDevice(areaID, pluginID, iid); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error(err)
			return err
		}
		return nil
	}
	// 更新物模型
	isGateway := tm.IsGateway()
	for _, ins := range tm.Instances {
		var e entity.Device

		isChildIns := !ins.IsGateway() && isGateway
		if !isChildIns {
			e, err = plugin.ThingModelToEntity(iid, tm, pluginID, areaID)
			if err != nil {
				return err
			}
			if err = entity.UpdateThingModel(&e); err != nil {
				return
			}
		} else {
			e, err = plugin.InstanceToEntity(ins, pluginID, areaID)
			if err != nil {
				logrus.Error(err)
				err = nil
				continue
			}
			if err = entity.UpdateThingModel(&e); err != nil {
				logrus.Error(err)
				err = nil
				continue
			}
			// 为所有角色增加改设备的权限
			if err = device.AddDevicePermissionForRoles(e, entity.GetDB()); err != nil {
				logrus.Error(err)
				err = nil
				continue
			}
			// 发送通知有设备增加
			m := event.NewEventMessage(event.DeviceIncrease, areaID)
			m.Param = map[string]interface{}{
				"device": e,
			}
			event.Notify(m)
		}
	}
	return nil
}

func UpdateDeviceShadowBeforeExecuteTask(em event.EventMessage) (err error) {
	if err = UpdateDeviceShadow(em); err != nil {
		return
	}

	err = ExecuteTask(em)
	return
}

func UpdateDeviceShadow(em event.EventMessage) error {

	attr := em.GetAttr()
	if attr == nil {
		logger.Warn(" attr is nil")
		return nil
	}
	dID := em.GetDeviceID()
	d, err := entity.GetDeviceByID(dID)
	if err != nil {
		return err
	}
	// 从设备影子中获取属性
	shadow, err := d.GetShadow()
	if err != nil {
		return err
	}
	shadow.UpdateReported(attr.IID, attr.AID, attr.Val)
	d.Shadow, err = json.Marshal(shadow)
	if err != nil {
		return err
	}
	if err = entity.GetDB().Save(d).Error; err != nil {
		return err
	}

	return nil
}

func ExecuteTask(em event.EventMessage) error {
	deviceID := em.GetDeviceID()
	d, err := entity.GetDeviceByID(deviceID)
	if err != nil {
		return err
	}
	attr := em.GetAttr()
	if attr == nil {
		logger.Warn("device or attr is nil")
		return nil
	}
	return task.GetManager().DeviceStateChange(d, *attr)
}

type State struct {
	thingmodel.Attribute
}

func RecordDeviceState(em event.EventMessage) (err error) {
	deviceID := em.Param["device_id"].(int)
	d, err := entity.GetDeviceByID(deviceID)
	if err != nil {
		return
	}

	// 将变更的属性保存记录设备状态

	tm, err := d.GetThingModel()
	if err != nil {
		return
	}

	ae := em.Param["attr"].(definer.AttributeEvent)
	attribute, err := tm.GetAttribute(ae.IID, ae.AID)
	if err != nil {
		return
	}
	state := State{Attribute: attribute}
	state.Val = ae.Val
	stateBytes, _ := json.Marshal(state)
	return entity.InsertDeviceState(d, stateBytes)
}
