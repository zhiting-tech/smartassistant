package device

import (
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
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

// deviceListReq 设备列表接口请求参数
type deviceListReq struct {
	Type listType `form:"type"`
}

// deviceListResp 设备列表接口返回数据
type deviceListResp struct {
	Devices []Device `json:"devices"`
}

// Device 设备信息
type Device struct {
	ID              int                   `json:"id"`
	IID             string                `json:"iid"`
	Name            string                `json:"name"`
	Logo            string                `json:"logo"` // logo相对路径
	LogoURL         string                `json:"logo_url"`
	PluginID        string                `json:"plugin_id"`
	LocationID      int                   `json:"location_id,omitempty"`
	LocationName    string                `json:"location_name,omitempty"`
	IsSA            bool                  `json:"is_sa"`
	Control         string                `json:"control"` // 控制页相对路径
	PluginURL       string                `json:"plugin_url"`
	Type            string                `json:"type"`
	DepartmentID    int                   `json:"department_id,omitempty"`
	DepartmentName  string                `json:"department_name,omitempty"`
	DeviceInstances thingmodel.ThingModel `json:"device_instances"`
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
		err = errors.New(status.RequireLogin)
		return
	}

	devices, err = entity.GetDevices(sessionUser.AreaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.Devices, err = WrapDevices(c, devices, req.Type)
	return
}

// ListLocationDevices 用于处理房间设备列表接口的请求
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
		devices, err = entity.GetDevicesByLocationID(id)
	} else {
		devices, err = entity.GetDevicesByDepartmentID(id)
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

	for _, d := range devices {
		// 有可以控制权限的设备（场景的执行任务页面使用），只显示有控制权限的
		if listType == WriteDevice && !isDeviceControllable(up, d) {
			continue
		}
		// 有读权限或者通知权限的设备（场景的触发条件页面使用）
		if listType == ReadOrNotifyDevice && !isDeviceTriggerable(d, up) {
			continue
		}
		logoUrl, logo := device.LogoInfo(c, d)
		dd := Device{
			ID:           d.ID,
			IID:          d.IID,
			Name:         d.Name,
			Logo:         logo,
			LocationID:   d.LocationID,
			DepartmentID: d.DepartmentID,
			LogoURL:      logoUrl,
			Type:         d.Type,
		}
		if d.Model == types.SaModel {
			dd.IsSA = true
			dd.PluginID = ""
		} else {
			location, _ := entity.GetLocationByID(d.LocationID)
			department, _ := entity.GetDepartmentByID(d.DepartmentID)
			dd.LocationName = location.Name
			dd.DepartmentName = department.Name
			dd.PluginID = d.PluginID
			var pluginURL *plugin.URL
			pluginURL, err = plugin.ControlURL(d, c.Request, u.UserID)
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
		}
		result = append(result, dd)

	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].IsSA {
			return true
		}
		return false
	})
	return
}

// isDeviceTriggerable 判断设备可以作为触发条件被选择
func isDeviceTriggerable(d entity.Device, up entity.UserPermissions) bool {

	attributes, err := d.UserControlAttributes(up)
	if err != nil {
		return false
	}

	for _, attr := range attributes {
		// 读和通知权限
		if attr.PermissionRead() || attr.PermissionNotify() {
			return true
		}
	}
	return false
}

// isDeviceControllable 判断设备可以作为执行任务被选择
func isDeviceControllable(up entity.UserPermissions, d entity.Device) bool {
	attributes, err := d.UserControlAttributes(up)
	if err != nil {
		return false
	}

	for _, attr := range attributes {
		if attr.PermissionWrite() {
			return true
		}

	}
	return false
}
