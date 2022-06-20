package websocket

import (
	"context"
	"encoding/json"
	errors2 "errors"
	"sort"

	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	version2 "github.com/zhiting-tech/smartassistant/modules/utils/version"
	"github.com/zhiting-tech/smartassistant/pkg/analytics"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/event"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/v2"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

type DeviceHandleParams struct {
	IID string `json:"iid"`
}

type getInstancesResp struct {
	thingmodel.ThingModel
	IsOnline     bool          `json:"is_online"`
	ParentDevice *ParentDevice `json:"parent_device,omitempty"`
	SyncData     string        `json:"sync_data"`
}

type ParentDevice struct {
	IID       string `json:"iid"`
	DeviceID  int    `json:"device_id"`
	Name      string `json:"name"`
	PluginUrl string `json:"plugin_url"`
}

func GetInstances(req Request) (result interface{}, err error) {

	var p DeviceHandleParams
	json.Unmarshal(req.Data, &p)
	user := req.User
	var resp getInstancesResp

	d, err := entity.GetPluginDevice(user.AreaID, req.Domain, p.IID)
	if err != nil {
		return
	}
	up, err := entity.GetUserPermissions(req.User.UserID)
	if err != nil {
		return
	}
	resp.ThingModel, err = d.GetThingModelWithState(up)
	if err != nil {
		return
	}
	identify := plugin.Identify{
		PluginID: req.Domain,
		IID:      p.IID,
		AreaID:   req.User.AreaID,
	}
	resp.IsOnline = plugin.GetGlobalClient().IsOnline(identify)
	resp.SyncData = d.SyncData
	if d.ParentIID != "" {
		var pDevice entity.Device
		pDevice, err = entity.GetPluginDevice(user.AreaID, req.Domain, d.ParentIID)
		if err != nil {
			return
		}
		var pluginToken string
		pluginToken, err = oauth.GetUserPluginToken(user.UserID, req.ginCtx.Request, user.AreaID)
		if err != nil {
			return
		}
		var pluginUrl *plugin.URL
		pluginUrl, err = plugin.ControlURLWithToken(pDevice, req.ginCtx.Request, pluginToken)
		if err != nil {
			return
		}
		resp.ParentDevice = &ParentDevice{
			IID:       pDevice.IID,
			DeviceID:  pDevice.ID,
			Name:      pDevice.Name,
			PluginUrl: pluginUrl.String(),
		}
	}
	result = resp
	return
}

func SetAttrs(req Request) (result interface{}, err error) {
	var sr sdk.SetRequest
	json.Unmarshal(req.Data, &sr)
	user := req.User
	up, err := entity.GetUserPermissions(user.UserID)
	if err != nil {
		return
	}
	// 判断控制权限
	for _, attr := range sr.Attributes {
		var d entity.Device
		d, err = entity.GetPluginDevice(req.User.AreaID, req.Domain, attr.IID)
		tm, _ := d.GetThingModel()
		_, err = tm.GetAttribute(attr.IID, attr.AID)
		if err != nil {
			err = errors.New(status.AttrNotFound)
			return
		}
		if err == nil {
			if !up.IsDeviceAttrControlPermit(d.ID, attr.AID) {
				err = errors.New(status.Deny)
				return
			}
		} else {
			if err != gorm.ErrRecordNotFound {
				return
			}
			err = nil
		}
	}
	// 发送控制命令
	err = plugin.SetAttributes(context.Background(), req.Domain, req.User.AreaID, sr)
	if err != nil {
		return
	}
	return
}

type connectDeviceResp struct {
	thingmodel.ThingModel
	Device Device `json:"device"`
}

type Device struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Model     string `json:"model"`
	PluginID  string `json:"plugin_id"`
	PluginURL string `json:"plugin_url"`
	Control   string `json:"control"` // 控制页相对路径
}

type ConnectParams struct {
	DeviceHandleParams
	AuthParams map[string]interface{} `json:"auth_params"`
}

// ConnectDevice 连接设备
func ConnectDevice(req Request) (result interface{}, err error) {
	if !entity.JudgePermit(req.User.UserID, types.DeviceAdd) {
		err = errors.New(status.Deny)
		return
	}

	var p ConnectParams
	json.Unmarshal(req.Data, &p)
	identify := plugin.Identify{
		PluginID: req.Domain,
		IID:      p.IID,
		AreaID:   req.User.AreaID,
	}
	tm, err := plugin.ConnectDevice(context.Background(), identify, p.AuthParams)
	if err != nil {
		return
	}
	resp := connectDeviceResp{
		ThingModel: tm,
	}

	isGateway := tm.IsBridge()
	for _, ins := range tm.Instances {
		var e entity.Device
		isChildIns := !ins.IsBridge() && isGateway
		if isChildIns {
			e, err = plugin.InstanceToEntity(ins, req.Domain, p.IID, req.User.AreaID)
			if err != nil {
				logger.Error(err)
				err = nil
				continue
			}
		} else {
			e, err = plugin.ThingModelToEntity(p.IID, tm, req.Domain, req.User.AreaID)
			if err != nil {
				return
			}
		}

		if err = device.Create(req.User.AreaID, &e); err != nil {
			logger.Error(err)
			err = nil
			continue
		}

		// 记录添加设备信息
		go analytics.RecordStruct(analytics.EventTypeDeviceAdd, req.User.UserID, e)
		// 通知SC
		em := event.NewEventMessage(event.DeviceIncrease, req.User.AreaID)
		em.Param = map[string]interface{}{
			"device": e,
		}
		event.Notify(em)
		if isChildIns {
			continue
		}

		resp.Device = Device{
			ID:       e.ID,
			Name:     e.Name,
			Model:    e.Model,
			PluginID: e.PluginID,
		}
		var pluginURL *plugin.URL
		pluginURL, err = plugin.ControlURL(e, req.ginCtx.Request, req.User.UserID)
		if err != nil {
			logger.Error(err)
			err = nil
			continue
		}
		resp.Device.PluginURL = pluginURL.String()
		resp.Device.Control = pluginURL.PluginPath()

	}
	if resp.Device.ID == 0 {
		err = errors.New(status.AddDeviceFail)
	}
	return resp, err
}

// DisconnectDevice 设备断开连接（取消配对等）
func DisconnectDevice(req Request) (result interface{}, err error) {
	if !entity.JudgePermit(req.User.UserID, types.DeviceDelete) {
		err = errors.New(status.Deny)
		return
	}
	var p ConnectParams
	if err = json.Unmarshal(req.Data, &p); err != nil {
		return
	}
	identify := plugin.Identify{
		PluginID: req.Domain,
		IID:      p.IID,
		AreaID:   req.User.AreaID,
	}
	err = plugin.DisconnectDevice(context.Background(), identify, p.AuthParams)
	if err != nil {
		logger.Errorf("disconnect device err: %s", err)
	}
	d, err := entity.GetPluginDevice(req.User.AreaID, req.Domain, p.IID)
	if err != nil {
		return
	}
	tm, err := d.GetThingModel()
	if err != nil {
		return
	}
	if err = entity.DelDeviceByID(d.ID); err != nil {
		return
	}

	if err = entity.RemoveCommonDevice(d.ID); err != nil {
		return
	}

	// 如果是网关，遍历子设备并删除
	for _, ins := range tm.Instances {
		if err = entity.DelDeviceByIID(req.User.AreaID, req.Domain, ins.IID); err != nil {
			logger.Errorf("del device by iid %s err %s", ins.IID, err)
			err = nil
			continue
		}
	}

	// 通知SC
	event.Notify(event.NewEventMessage(event.DeviceDecrease, req.User.AreaID))
	// 记录删除设备信息
	go analytics.RecordStruct(analytics.EventTypeDeviceDelete, req.User.UserID, d)

	return
}

type CheckUpdateResp struct {
	UpdateAvailable bool           `json:"update_available"`
	CurrentVersion  string         `json:"current_version"`
	LatestVersion   cloud.Firmware `json:"latest_version"`
}

// CheckUpdate 检查设备是否有更新
func CheckUpdate(req Request) (result interface{}, err error) {
	var p DeviceHandleParams
	json.Unmarshal(req.Data, &p)
	user := req.User
	d, err := device.GetThingModel(user.AreaID, req.Domain, p.IID)
	if err != nil {
		return
	}

	info, err := d.GetInfo(p.IID)
	if err != nil {
		return
	}

	latestFirmware, err := cloud.GetLatestFirmwareWithContext(context.TODO(), req.Domain, info.Model)
	if err != nil {
		if errors2.Is(err, cloud.ErrNoFirmware) {
			result = CheckUpdateResp{
				UpdateAvailable: false,
				CurrentVersion:  info.Version,
			}
			err = nil
		}
		return
	}
	ok, err := version2.Greater(latestFirmware.Version, info.Version)
	if err != nil {
		return
	}

	result = CheckUpdateResp{
		UpdateAvailable: ok,
		CurrentVersion:  info.Version,
		LatestVersion:   latestFirmware,
	}
	return
}

// OTA 更新固件
func OTA(req Request) (result interface{}, err error) {

	var p DeviceHandleParams
	json.Unmarshal(req.Data, &p)
	user := req.User

	d, err := entity.GetPluginDevice(user.AreaID, req.Domain, p.IID)
	if err != nil {
		return
	}

	permission := types.NewDeviceFwUpgrade(d.ID)
	if !entity.JudgePermit(req.User.UserID, permission) {
		err = errors.New(status.Deny)
		return
	}
	instances, err := d.GetThingModel()
	if err != nil {
		return
	}

	info, err := instances.GetInfo(p.IID)
	if err != nil {
		return
	}
	firmware, err := cloud.GetLatestFirmwareWithContext(context.TODO(), req.Domain, info.Model)
	if err != nil {
		return
	}
	err = plugin.OTA(context.Background(), user.AreaID, req.Domain, p.IID, firmware.URL)
	if err != nil {
		return
	}

	// 硬件响应ota成功仅表示固件flash成功，真正是否成功需要等待硬件重启后才能确定固件版本号
	return
}

type gatewayInfo struct {
	gateway
	Name     string `json:"name"`
	IID      string `json:"iid"`
	IsOnline bool   `json:"is_online"`
}

type listDevicesReq struct {
	Model string
}
type listDevicesResp struct {
	Gateways        []gatewayInfo `json:"gateways"`
	SupportGateways []gateway     `json:"support_gateways"`
}

type gateway struct {
	Model    string `json:"model"`
	LogoURL  string `json:"logo_url"`
	PluginID string `json:"plugin_id"`
}

// ListGateways 网关列表，返回所有设备支持的网关，子设备需要通过控制网关实现入网
func ListGateways(req Request) (result interface{}, err error) {
	var resp listDevicesResp
	resp.SupportGateways = make([]gateway, 0)
	resp.Gateways = make([]gatewayInfo, 0)
	var listReq listDevicesReq
	json.Unmarshal(req.Data, &listReq)

	// 获取设备的配置信息中支持的网关列表
	conf := plugin.GetGlobalClient().Config(req.Domain).SupportDeviceConfig(listReq.Model)
	// 记录支持的网关列表，方便后面查询
	gwMap := make(map[string]gateway)
	for _, sg := range conf.SupportGateways {
		name := sg.Name
		if name == "" {
			name = sg.Model
		}
		gw := gateway{
			Model:    name,
			LogoURL:  plugin.DeviceLogoURL(req.ginCtx.Request, req.Domain, sg.Model, ""),
			PluginID: req.Domain,
		}
		resp.SupportGateways = append(resp.SupportGateways, gw)
		gwMap[sg.Model] = gw
	}

	// 遍历已添加设备获取返回支持的网关
	ds, err := entity.GetPluginDevices(req.User.AreaID, req.Domain)
	if err != nil {
		return
	}
	for _, d := range ds {
		if gw, ok := gwMap[d.Model]; ok {
			identify := plugin.Identify{
				PluginID: d.PluginID,
				IID:      d.IID,
				AreaID:   d.AreaID,
			}
			isOnline := plugin.GetGlobalClient().IsOnline(identify)

			e := gatewayInfo{
				gateway:  gw,
				Name:     d.Name,
				IID:      d.IID,
				IsOnline: isOnline,
			}
			resp.Gateways = append(resp.Gateways, e)
		}
	}
	return resp, nil
}

type DeviceStatesReq struct {
	DeviceHandleParams
	AttrType *string `json:"attr_type"` // 属性类型
	StartAt  *int64  `json:"start_at"`
	EndAt    *int64  `json:"end_at"`
	Size     int     `json:"size"`  // 分页大小
	Index    *int    `json:"index"` // 日志id，用于滚动加载
}

type DeviceStatesResp struct {
	States []State `json:"states"`
}

type State struct {
	thingmodel.Attribute `json:",inline"`
	ID                   int   `json:"id"`
	Timestamp            int64 `json:"timestamp"`
}

// DeviceStates 设备状态（日志）
func DeviceStates(req Request) (result interface{}, err error) {
	var deviceStatesReq DeviceStatesReq
	json.Unmarshal(req.Data, &deviceStatesReq)
	d, err := entity.GetPluginDevice(req.User.AreaID, req.Domain, deviceStatesReq.IID)
	if err != nil {
		return
	}

	if deviceStatesReq.Size <= 0 {
		// deviceStatesReq.Size = 20
	}

	var resp DeviceStatesResp
	resp.States = make([]State, 0)
	states, err := entity.GetDeviceStates(d.ID, deviceStatesReq.AttrType, deviceStatesReq.Size,
		deviceStatesReq.Index, deviceStatesReq.StartAt, deviceStatesReq.EndAt)
	for _, state := range states {
		var s State
		json.Unmarshal(state.State, &s)
		s.Timestamp = state.CreatedAt.Unix()
		s.ID = state.ID
		resp.States = append(resp.States, s)
	}

	return resp, err
}

type SubDevicesResp struct {
	SubDevices        []SubDevice        `json:"devices"`             // 已连接的子设备
	SupportSubDevices []SupportSubDevice `json:"support_sub_devices"` // 支持的子设备
}
type SubDevice struct {
	ID              int                   `json:"id"`
	Name            string                `json:"name"`
	LogoURL         string                `json:"logo_url"`
	PluginURL       string                `json:"plugin_url"`
	Control         string                `json:"control"`
	LocationName    string                `json:"location_name"`
	DepartmentName  string                `json:"department_name"`
	ZigbeeSupport   bool                  `json:"zigbee_support"`
	BleSupport      bool                  `json:"ble_support"`
	DeviceInstances thingmodel.ThingModel `json:"device_instances"`
}
type SupportSubDevice struct {
	Model         string `json:"model"`
	LogoURL       string `json:"logo_url"`
	ZigbeeSupport bool   `json:"zigbee_support"`
	BleSupport    bool   `json:"ble_support"`
}

func SubDevices(req Request) (result interface{}, err error) {
	var deviceLogReq DeviceHandleParams
	json.Unmarshal(req.Data, &deviceLogReq)
	user := req.User
	var tm thingmodel.ThingModel
	pluginDevice, err := entity.GetPluginDevice(user.AreaID, req.Domain, deviceLogReq.IID)
	if err != nil {
		return
	}
	up, err := entity.GetUserPermissions(req.User.UserID)
	if err != nil {
		return
	}
	tm, err = pluginDevice.GetThingModelWithState(up)
	if err != nil {
		return
	}

	logger.Debugf("found %d instance from %s", len(tm.Instances), deviceLogReq.IID)
	// 遍历物模型，获取子设备IID,从数据库中获取子设备列表
	var resp SubDevicesResp
	resp.SubDevices = make([]SubDevice, 0)
	resp.SupportSubDevices = make([]SupportSubDevice, 0)

	pluginToken, err := oauth.GetUserPluginToken(req.User.UserID, req.ginCtx.Request, req.User.AreaID)
	if err != nil {
		return nil, err
	}
	// 获取网关支持的子设备
	config := plugin.GetGlobalClient().Config(req.Domain)
	var d entity.Device
	d, err = entity.GetPluginDevice(user.AreaID, req.Domain, deviceLogReq.IID)
	if err != nil {
		return
	}
	configMap := make(map[string]plugin.DeviceConfig)
	for _, deviceConf := range config.SupportDevices {
		if deviceConf.IsGatewaySupport(d.Model) {
			sd := SupportSubDevice{
				ZigbeeSupport: deviceConf.ZigbeeSupport,
				BleSupport:    deviceConf.BleSupport,
				Model:         deviceConf.Model,
				LogoURL:       plugin.PluginTargetURL(req.ginCtx.Request, req.Domain, deviceConf.Model, deviceConf.Logo),
			}
			resp.SupportSubDevices = append(resp.SupportSubDevices, sd)
			configMap[deviceConf.Model] = deviceConf
		}
	}
	for _, instance := range tm.Instances {
		if instance.IsBridge() {
			continue
		}
		var d entity.Device
		d, err = entity.GetPluginDevice(user.AreaID, req.Domain, instance.IID)
		if err != nil {
			logger.Error(err)
			continue
		}
		info, err := instance.GetInfo()
		if err != nil {
			logger.Error(err)
			continue
		}
		deviceConf := configMap[info.Model]
		location, _ := entity.GetLocationByID(d.LocationID)
		department, _ := entity.GetDepartmentByID(d.DepartmentID)
		subDevice := SubDevice{
			ID:             d.ID,
			Name:           d.Name,
			LogoURL:        device.LogoURL(req.ginCtx.Request, d),
			LocationName:   location.Name,
			DepartmentName: department.Name,
			ZigbeeSupport:  deviceConf.ZigbeeSupport,
			BleSupport:     deviceConf.BleSupport,
			DeviceInstances: thingmodel.ThingModel{
				Instances: []thingmodel.Instance{instance},
			},
		}
		var pluginURL *plugin.URL
		pluginURL, err = plugin.ControlURLWithToken(d, req.ginCtx.Request, pluginToken)
		if err != nil {
			logger.Error(err)
			continue
		}
		subDevice.PluginURL = pluginURL.String()
		subDevice.Control = pluginURL.PluginPath()
		resp.SubDevices = append(resp.SubDevices, subDevice)
	}
	subDevices := resp.SubDevices
	sort.SliceStable(subDevices, func(i, j int) bool {
		if subDevices[i].ID < subDevices[j].ID {
			return true
		}
		return false
	})

	return resp, nil
}

func RegisterCmd() {
	RegisterCallFunc(ServiceOTA, OTA)                     // OTA
	RegisterCallFunc(ServiceConnect, ConnectDevice)       // 添加/连接设备
	RegisterCallFunc(ServiceSetAttributes, SetAttrs)      // 设置属性
	RegisterCallFunc(ServiceCheckUpdate, CheckUpdate)     // 检查固件更新
	RegisterCallFunc(ServiceGetInstances, GetInstances)   // 获取物模型
	RegisterCallFunc(ServiceDisconnect, DisconnectDevice) // 删除设备/断开连接

	RegisterCallFunc(ServiceSubDevices, SubDevices)     // 子设备列表
	RegisterCallFunc(ServiceListGateways, ListGateways) // 列出网关列表
	RegisterCallFunc(ServiceDeviceStates, DeviceStates) // 设备状态（日志）
}
