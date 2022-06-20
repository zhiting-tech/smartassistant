package event

import (
	"encoding/json"
	"sync"

	"github.com/sirupsen/logrus"

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
	event.RegisterEvent(event.AttributeChange, ws.MulticastMsg,
		UpdateDeviceShadowBeforeExecuteTask, RecordDeviceState)
	event.RegisterEvent(event.DeviceDecrease, ws.MulticastMsg)
	event.RegisterEvent(event.DeviceIncrease, ws.MulticastMsg)
	event.RegisterEvent(event.OnlineStatus, ws.MulticastMsg)
	event.RegisterEvent(event.ThingModelChange, UpdateThingModel, ws.MulticastMsg)
}

func UpdateThingModel(em event.EventMessage) (err error) {
	tm := em.Param["thing_model"].(thingmodel.ThingModel)
	areaID := em.AreaID
	pluginID := em.Param["plugin_id"].(string)
	iid := em.Param["iid"].(string)

	// 网关未添加则设备不更新
	isGatewayExist, err := entity.IsDeviceExist(areaID, pluginID, iid)
	if err != nil {
		return
	}
	if !isGatewayExist {
		return
	}
	// 更新设备物模型
	var primaryDevice entity.Device
	primaryDevice, err = plugin.ThingModelToEntity(iid, tm, pluginID, areaID)
	if err != nil {
		return err
	}
	if err = entity.UpdateThingModel(&primaryDevice); err != nil {
		return
	}

	// 更新子设备物模型
	for _, ins := range tm.Instances {
		if ins.IsBridge() {
			continue
		}
		var e entity.Device

		e, err = plugin.InstanceToEntity(ins, pluginID, iid, areaID)
		if err != nil {
			logrus.Error(err)
			err = nil
			continue
		}
		configDeviceName := e.Name

		// 更新前判断子设备是否已存在
		isChildExist, _ := entity.IsDeviceExist(areaID, pluginID, e.IID)
		if err = entity.UpdateThingModel(&e); err != nil {
			logrus.Error(err)
			err = nil
			continue
		}

		// 子设备不存在则为所有角色增加改设备的权限 && 更新子设备房间为网关默认房间
		if !isChildExist {
			if err = device.AddDevicePermissionForRoles(e, entity.GetDB()); err != nil {
				logrus.Error(err)
				err = nil
				continue
			}

			// 更新设备房间为默认房间
			updates := map[string]interface{}{
				"name":           configDeviceName,
				"location_id":    primaryDevice.LocationID,
				"location_order": 0,
			}
			if err = entity.UpdateDeviceWithMap(e.ID, updates); err != nil {
				logrus.Error(err)
				err = nil
				continue
			}
		}

		// 发送通知有设备增加 FIXME 更新也发通知（子设备重复添加时需要通知来完成添加流程）
		m := event.NewEventMessage(event.DeviceIncrease, areaID)
		m.Param = map[string]interface{}{
			"device": e,
		}
		event.Notify(m)

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

var updateShadowMap sync.Map

func UpdateDeviceShadow(em event.EventMessage) error {

	attr := em.GetAttr()
	if attr == nil {
		logger.Warn(" attr is nil")
		return nil
	}
	deviceID := em.GetDeviceID()

	val, _ := updateShadowMap.LoadOrStore(deviceID, &sync.Mutex{})
	mu := val.(*sync.Mutex)
	mu.Lock()
	defer func() {
		mu.Unlock()
		updateShadowMap.Delete(deviceID)
	}()

	d, err := entity.GetDeviceByID(deviceID)
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
	if err = entity.GetDB().Model(&entity.Device{ID: d.ID}).Update("shadow", d.Shadow).Error; err != nil {
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
	deviceID := em.GetDeviceID()
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
