package device

import (
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/event"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/analytics"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// DelDevice 用于处理删除设备接口的请求
func DelDevice(c *gin.Context) {
	var (
		err      error
		deviceId int
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	deviceId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	p := types.NewDeviceDelete(deviceId)
	if !device.IsPermit(c, p) {
		err = errors.Wrap(err, status.Deny)
		return
	}
	// TODO：考虑优化，plugin.RemoveDevice也会查询一次
	device, err := entity.GetDeviceByID(deviceId)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if err = plugin.RemoveDevice(deviceId); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	event.GetServer().Notify(event.NewEventMessage(event.DeviceDecrease, session.Get(c).AreaID))
	uid := session.Get(c).UserID
	go analytics.RecordStruct(analytics.EventTypeDeviceDelete, uid, device)

	return
}
