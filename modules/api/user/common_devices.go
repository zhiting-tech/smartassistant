package user

import (
	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/device"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
)

type updateUserCommonDevicesReq struct {
	Devices []int `json:"devices"`
}
type userCommonDevicesResp struct {
	Devices []device.Device `json:"devices"`
}

func updateUserCommonDevices(c *gin.Context) {
	var (
		req  updateUserCommonDevicesReq
		resp userCommonDevicesResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()
	if err = c.BindJSON(&req); err != nil {
		return
	}
	u := session.Get(c)

	if err = entity.UpdateUserCommonDevices(u.UserID, u.AreaID, req.Devices); err != nil {
		return
	}
	devices, err := entity.GetUserCommonDevices(u.UserID)
	if err != nil {
		return
	}
	resp.Devices, err = device.WrapDevices(c, devices, device.AllDevice)

}

func getUserCommonDevices(c *gin.Context) {
	var (
		resp userCommonDevicesResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()
	u := session.Get(c)

	devices, err := entity.GetUserCommonDevices(u.UserID)
	if err != nil {
		return
	}
	resp.Devices, err = device.WrapDevices(c, devices, device.AllDevice)
}
