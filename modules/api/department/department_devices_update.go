package department

import (
	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/device"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
)

type updateDepartmentDevicesReq struct {
	DepartmentID int   `uri:"id" binding:"required"`
	Devices      []int `json:"devices"`
}

type updateDepartmentDevicesResp struct {
	Devices []device.Device
}

// updateDepartmentDevices 更新部门设备，增加，删除，排序房间设备
func updateDepartmentDevices(c *gin.Context) {

	var (
		req  updateDepartmentDevicesReq
		resp updateDepartmentDevicesResp
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
	if err = entity.ReorderDepartmentDevices(req.DepartmentID, req.Devices); err != nil {
		return
	}
	devices, err := entity.GetOrderDepartmentDevices(req.DepartmentID)
	if err != nil {
		return
	}

	if resp.Devices, err = device.WrapDevices(c, devices, device.AllDevice); err != nil {
		return
	}
}
