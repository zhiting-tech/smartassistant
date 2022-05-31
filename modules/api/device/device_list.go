package device

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

// 设备列表过滤条件
type listType int

// 0:所有设备;1:有控制权限的设备;2有读或通知权限的设备
const (
	AllDevice          listType = iota // 所有设备
	WriteDevice                        // 有写权限的设备（场景的执行任务页面使用）
	ReadOrNotifyDevice                 // 有读或通知权限的设备（定时任务触发条件使用）
)

type orderType int

const (
	naturalOrder orderType = 0
	nameOrder    orderType = 1
	customOrder  orderType = 2
)

// deviceListReq 设备列表接口请求参数
type deviceListReq struct {
	Type      listType  `form:"type"`
	OrderType orderType `form:"order_type"`
}

// deviceListResp 设备列表接口返回数据
type deviceListResp struct {
	Devices []Device `json:"devices"`
}

// Device 设备信息
type Device struct {
	ID           int    `json:"id"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	IID          string `json:"iid"`
	Name         string `json:"name"`
	Logo         string `json:"logo"` // logo相对路径
	LogoURL      string `json:"logo_url"`
	PluginID     string `json:"plugin_id"` // 为空时表示不使用插件，仅在客户端之间同步
	IsSA         bool   `json:"is_sa"`
	Control      string `json:"control"` // 控制页相对路径
	PluginURL    string `json:"plugin_url"`
	Type         string `json:"type"`
	IsOnline     bool   `json:"is_online"`

	LocationID      int    `json:"location_id,omitempty"`
	LocationOrder   int    `json:"location_order,omitempty"`
	LocationName    string `json:"location_name,omitempty"`
	DepartmentID    int    `json:"department_id,omitempty"`
	DepartmentOrder int    `json:"department_order,omitempty"`
	DepartmentName  string `json:"department_name,omitempty"`

	DeviceInstances thingmodel.ThingModel `json:"device_instances"`

	SyncData string `json:"sync_data,omitempty"`
}

// ListAllDevice 用于处理设备列表接口的请求
func ListAllDevice(c *gin.Context) {

	var (
		err     error
		req     deviceListReq
		resp    deviceListResp
		devices []entity.Device
	)
	defer func() {
		if resp.Devices == nil {
			resp.Devices = make([]Device, 0)
		}
		response.HandleResponse(c, err, resp)
	}()
	if err = c.BindQuery(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	sessionUser := session.Get(c)
	if sessionUser == nil {
		err = errors.New(status.InvalidUserCredentials)
		return
	}

	switch req.OrderType {
	case nameOrder:
		devices, err = entity.GetDevicesOrderByPinyin(sessionUser.AreaID)
	case customOrder:
		devices, err = entity.GetOrderDevices(sessionUser.AreaID)
	case naturalOrder:
		devices, err = entity.GetDevices(sessionUser.AreaID)
	default:
		devices, err = entity.GetDevices(sessionUser.AreaID)
	}
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.Devices, err = WrapDevices(c, devices, req.Type)
	return
}

// ListLocationDevices 用于处理房间/部门设备列表接口的请求
func ListLocationDevices(c *gin.Context) {
	var (
		err     error
		req     deviceListReq
		resp    deviceListResp
		devices []entity.Device
		curArea entity.Area
	)
	defer func() {
		if resp.Devices == nil {
			resp.Devices = make([]Device, 0)
		}
		response.HandleResponse(c, err, resp)
	}()
	if err = c.BindQuery(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if curArea, err = entity.GetAreaByID(session.Get(c).AreaID); err != nil {
		return
	}
	if entity.IsHome(curArea.AreaType) {
		switch req.OrderType {
		case customOrder:
			devices, err = entity.GetOrderLocationDevices(id)
		default:
			devices, err = entity.GetDevicesByLocationID(id)
		}
	} else {
		switch req.OrderType {
		case customOrder:
			devices, err = entity.GetOrderDepartmentDevices(id)
		default:
			devices, err = entity.GetDevicesByDepartmentID(id)
		}
	}
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.Devices, err = WrapDevices(c, devices, req.Type)
	return

}

func WrapDevices(c *gin.Context, devices []entity.Device, listType listType) (result []Device, err error) {

	u := session.Get(c)
	up, err := entity.GetUserPermissions(u.UserID)
	if err != nil {
		return
	}

	pluginToken, err := oauth.GetUserPluginToken(u.UserID, c.Request, u.AreaID)
	if err != nil {
		return nil, err
	}
	for _, d := range devices {
		// 有可以控制权限的设备（场景的执行任务页面使用），只显示有控制权限的
		if listType == WriteDevice && !d.IsControllable(up) {
			continue
		}
		// 有读权限或者通知权限的设备（场景的触发条件页面使用）
		if listType == ReadOrNotifyDevice && !d.IsTriggerable() {
			continue
		}
		logoUrl, logo := device.LogoInfo(c, d)
		dd := Device{
			ID:           d.ID,
			Model:        d.Model,
			Manufacturer: d.Manufacturer,
			IID:          d.IID,
			Name:         d.Name,
			Logo:         logo,
			LocationID:   d.LocationID,
			DepartmentID: d.DepartmentID,
			LogoURL:      logoUrl,
			Type:         d.Type,
		}
		if d.IsSa() {
			dd.IsSA = true
			dd.PluginID = ""
			dd.IsOnline = true // sa设备默认为true
		} else {
			location, _ := entity.GetLocationByID(d.LocationID)
			department, _ := entity.GetDepartmentByID(d.DepartmentID)
			dd.LocationName = location.Name
			dd.LocationOrder = d.LocationOrder
			dd.DepartmentName = department.Name
			dd.DepartmentOrder = d.DepartmentOrder
			if d.PluginID != "" {
				dd.PluginID = d.PluginID
				var pluginURL *plugin.URL
				pluginURL, err = plugin.ControlURLWithToken(d, c.Request, pluginToken)
				if err != nil {
					logger.Errorf("Get plugin url err: %v\n", err)
					err = nil
					continue
				}
				dd.PluginURL = pluginURL.String()
				dd.Control = pluginURL.PluginPath()
				dd.DeviceInstances, err = d.GetThingModelWithState(up)
				if err != nil {
					logger.Errorf("Get Device instances err: %v\n", err)
					err = nil
				}
				dd.IsOnline = plugin.GetGlobalClient().IsOnline(plugin.Identify{
					PluginID: d.PluginID,
					IID:      d.IID,
					AreaID:   d.AreaID,
				})
			}
			dd.SyncData = d.SyncData
		}
		result = append(result, dd)

	}
	return
}
