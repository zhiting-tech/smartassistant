package location

import (
	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/device"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
)

type updateLocationDevicesReq struct {
	LocationID int   `uri:"id" binding:"required"`
	Devices    []int `json:"devices"`
}

type updateLocationDevicesResp struct {
	Devices []device.Device
}

// updateLocationDevices 更新房间设备，增加，删除，排序房间设备
func updateLocationDevices(c *gin.Context) {

	var (
		req  updateLocationDevicesReq
		resp updateLocationDevicesResp
		err  error
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()
	if err = c.BindUri(&req); err != nil {
		return
	}

	if err = c.BindJSON(&req); err != nil {
		return
	}
	if err = entity.ReorderLocationDevices(req.LocationID, req.Devices); err != nil {
		return
	}
	devices, err := entity.GetOrderLocationDevices(req.LocationID)
	if err != nil {
		return
	}

	if resp.Devices, err = device.WrapDevices(c, devices, device.AllDevice); err != nil {
		return
	}
}
