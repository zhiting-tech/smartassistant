package device

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-unidecode"

	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// UpdateDeviceReq 修改设备接口请求参数
type UpdateDeviceReq struct {
	Name            *string `json:"name"`
	LogoType        *int    `json:"logo_type"`
	LocationID      int     `json:"location_id"`
	DepartmentID    int     `json:"department_id"`
	Common          *bool   `json:"common"`
	CascadeLocation bool    `json:"cascade_location"`

	SyncData string `json:"sync_data"`
}

func (req *UpdateDeviceReq) Validate() (updateDevice entity.Device, err error) {
	if req.LocationID != 0 {
		if _, err = entity.GetLocationByID(req.LocationID); err != nil {
			return
		}
	}
	updateDevice.LocationID = req.LocationID
	if req.DepartmentID != 0 {
		if _, err = entity.GetDepartmentByID(req.DepartmentID); err != nil {
			return
		}
	}
	updateDevice.DepartmentID = req.DepartmentID

	if req.Name != nil {
		if err = checkDeviceName(*req.Name); err != nil {
			return
		} else {
			updateDevice.Name = *req.Name
			updateDevice.Pinyin = unidecode.Unidecode(*req.Name)
		}
	}

	if req.LogoType != nil {
		_, ok := types.GetLogo(types.LogoType(*req.LogoType))
		if types.LogoType(*req.LogoType) != types.NormalLogo && !ok {
			err = errors.New(status.DeviceLogoNotExist)
			return
		}
		updateDevice.LogoType = req.LogoType
	}
	return
}

func checkDeviceName(name string) (err error) {

	if name == "" || strings.TrimSpace(name) == "" {
		err = errors.Wrap(err, status.DeviceNameInputNilErr)
		return
	}

	if utf8.RuneCountInString(name) > 20 {
		err = errors.Wrap(err, status.DeviceNameLengthLimit)
		return
	}
	return
}

// UpdateDevice 用于处理修改设备接口的请求
func UpdateDevice(c *gin.Context) {
	var (
		err          error
		req          UpdateDeviceReq
		id           int
		updateDevice entity.Device
		curDevice    entity.Device
		curArea      entity.Area
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()
	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	id, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if curDevice, err = entity.GetDeviceByID(id); err != nil {
		err = errors.New(status.DeviceNotExist)
		return
	}

	p := types.NewDeviceUpdate(id)
	if !device.IsPermit(c, p) {
		err = errors.Wrap(err, status.Deny)
		return
	}

	if curArea, err = entity.GetAreaByID(curDevice.AreaID); err != nil {
		return
	}

	if updateDevice, err = req.Validate(); err != nil {
		return
	}

	if req.LocationID == 0 && entity.IsHome(curArea.AreaType) {
		// 未勾选房间, 设备与房间解绑
		if err = entity.UnBindLocationDevice(id); err != nil {
			return
		}
	}

	if req.DepartmentID == 0 && entity.IsCompany(curArea.AreaType) {
		// 未勾选房间, 设备与部门解绑
		if err = entity.UnBindDepartmentDevice(id); err != nil {
			return
		}
	}
	if len(req.SyncData) != 0 {
		updateDevice.SyncData = req.SyncData
	}

	if err = entity.UpdateDevice(id, updateDevice); err != nil {
		return
	}

	if req.CascadeLocation == true && req.LocationID != 0 {
		if err = entity.UpdateSubDevicesLocation(curDevice.IID, updateDevice.LocationID); err != nil {
			return
		}
	}

	if req.Common != nil {
		u := session.Get(c)
		if *req.Common {
			if err = entity.SetDeviceCommon(u.UserID, u.AreaID, id); err != nil {
				return
			}
		} else {
			if err = entity.RemoveUserCommonDevice(u.UserID, u.AreaID, id); err != nil {
				return
			}
		}
	}

	return
}
