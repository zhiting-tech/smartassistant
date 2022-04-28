package device

import (
	errors2 "errors"
	"net/http"
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/device"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type infoType int

const (
	AllAttributes          infoType = iota // 所有属性
	WriteAttributes                        // 有写权限的属性 (场景执行任务使用)
	ReadOrNotifyAttributes                 // 有读权限或者通知的属性 (场景触发条件使用)
)

// infoDeviceReq 设备详情接口请求参数
type infoDeviceReq struct {
	Type infoType `form:"type"`
}

// infoDeviceResp 设备详情接口返回数据
type infoDeviceResp struct {
	Device infoDevice `json:"device_info"`
}

// infoDevice 设备详情
type infoDevice struct {
	ID         int          `json:"id"`
	IID        string       `json:"iid"`
	Name       string       `json:"name"`
	LogoURL    string       `json:"logo_url"`
	Model      string       `json:"model"`
	Location   infoLocation `json:"location,omitempty"`
	Department infoLocation `json:"department,omitempty"`
	Plugin     infoPlugin   `json:"plugin"`
	Logo       infoLogo     `json:"logo"`

	Attributes []entity.Attribute `json:"attributes"` // 有权限的action

	Permissions Permissions `json:"permissions"`
}

// Permissions 设备权限
type Permissions struct {
	UpdateDevice bool `json:"update_device"`
	DeleteDevice bool `json:"delete_device"`
}

// infoLocation 设备所属房间/部门详情
type infoLocation struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// infoPlugin 设备的插件详情
type infoPlugin struct {
	Name string `json:"name"`
	ID   string `json:"id"`

	URL     string `json:"url"`
	Control string `json:"control"` // 控制页相对路径
}

type infoLogo struct {
	Type int    `json:"type"`
	Name string `json:"name"`
}

// InfoDevice 用于处理设备详情接口的请求
func InfoDevice(c *gin.Context) {
	var (
		err        error
		id         int
		deviceInfo entity.Device
		req        infoDeviceReq
		resp       infoDeviceResp
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindQuery(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	id, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	if deviceInfo, err = entity.GetDeviceByID(id); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.NotFound)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}

	if resp.Device, err = BuildInfoDevice(deviceInfo, session.Get(c), c.Request, req.Type); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return

}

func BuildInfoDevice(d entity.Device, user *session.User, req *http.Request, infoType infoType) (iDevice infoDevice, err error) {
	var (
		iLocation infoLocation
		location  entity.Location

		iDepartment infoLocation
		department  entity.Department
		iLogo       infoLogo
	)
	if d.LocationID > 0 {
		if location, err = entity.GetLocationByID(d.LocationID); err != nil {
			if !errors2.Is(err, gorm.ErrRecordNotFound) {
				return
			} else {
				err = nil
			}
		} else {
			iLocation.ID = d.LocationID
			iLocation.Name = location.Name
		}
	}

	if d.DepartmentID > 0 {
		if department, err = entity.GetDepartmentByID(d.DepartmentID); err != nil {
			if !errors2.Is(err, gorm.ErrRecordNotFound) {
				return
			} else {
				err = nil
			}
		} else {
			iDepartment.ID = d.DepartmentID
			iDepartment.Name = department.Name
		}
	}

	currentLogoInfo := types.NormalLogoInfo
	if d.LogoType != nil && *d.LogoType != int(types.NormalLogo) {
		if logoInfo, ok := types.GetLogo(types.LogoType(*d.LogoType)); ok {
			currentLogoInfo = logoInfo
		}
	}

	iLogo.Type = int(currentLogoInfo.LogoType)
	iLogo.Name = currentLogoInfo.Name

	iDevice = infoDevice{
		ID:         d.ID,
		IID:        d.IID,
		Name:       d.Name,
		Model:      d.Model,
		Location:   iLocation,
		Department: iDepartment,
		LogoURL:    device.LogoURL(req, d),
		Logo:       iLogo,
	}

	userID := user.UserID
	if d.Model != types.SaModel {
		iDevice.Plugin = infoPlugin{
			Name: d.PluginID,
			ID:   d.PluginID,
		}
		var pluginURL *plugin.URL
		pluginURL, err = plugin.ControlURL(d, req, user.UserID)
		if err != nil {
			return
		}
		iDevice.Plugin.URL = pluginURL.String()
		iDevice.Plugin.Control = pluginURL.PluginPath()
		iDevice.Attributes, err = getDeviceAttributes(userID, d, infoType)
		if err != nil {
			return
		}
	}
	iDevice.Permissions.DeleteDevice = entity.JudgePermit(userID,
		types.NewDeviceDelete(d.ID))
	iDevice.Permissions.UpdateDevice = entity.JudgePermit(userID,
		types.NewDeviceUpdate(d.ID))

	return
}

// getDeviceAttributes 获取设备有权限的action
func getDeviceAttributes(userID int, d entity.Device, infoType infoType) (as []entity.Attribute, err error) {

	up, err := entity.GetUserPermissions(userID)
	if err != nil {
		return
	}

	attributes, err := d.UserControlAttributes(up)
	if err != nil {
		return
	}

	for _, attr := range attributes {

		switch infoType {
		case WriteAttributes:
			if !attr.PermissionWrite() {
				continue
			}
		case ReadOrNotifyAttributes:
			// 读和通知权限
			if !attr.PermissionRead() && !attr.PermissionNotify() {
				continue
			}
		default:
		}

		as = append(as, attr)
	}
	return
}
