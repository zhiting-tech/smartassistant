package device

import (
	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
)

type deviceReorderReq struct {
	Devices []int `json:"devices"`
}

type deviceReorderResp struct {
	Devices []Device
}

func deviceReorder(c *gin.Context) {

	var (
		req  deviceReorderReq
		resp deviceReorderResp
		err  error
	)

	defer func() {
		response.HandleResponse(c, err, &resp)
	}()
	if err = c.BindJSON(&req); err != nil {
		return
	}

	u := session.Get(c)
	if err = entity.ReorderDevices(u.AreaID, req.Devices); err != nil {
		return
	}
	devices, err := entity.GetOrderDevices(u.AreaID)
	if err != nil {
		return
	}

	if resp.Devices, err = WrapDevices(c, devices, AllDevice); err != nil {
		return
	}
}
