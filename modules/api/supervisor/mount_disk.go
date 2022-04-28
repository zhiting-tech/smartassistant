package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/disk"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type mountReq struct {
	PVName string `json:"pv_name"`
}

func MountDisk(c *gin.Context) {
	var (
		req mountReq
		err error
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	client, err := disk.NewDiskManagerClient()
	if err != nil {
		logger.Warnf("handlerFirstPath new DiskManager Client err is %s", err)
		return
	}
	err = client.MountPhysicalWithContext(c.Request.Context(), req.PVName)
	if err != nil {
		err = errors.New(errors.InternalServerErr)
		return
	}

}
