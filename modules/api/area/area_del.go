package area

import (
	"github.com/zhiting-tech/smartassistant/modules/api/extension"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/clouddisk"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type DelAreaReq struct {
	IsDelCloudDisk *bool `json:"is_del_cloud_disk"`
}

// DelArea 用于处理删除家庭接口的请求
func DelArea(c *gin.Context) {
	var (
		id  uint64
		err error
		req DelAreaReq
		resp clouddisk.DelAreaStatus
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	id, err = strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if err = c.BindJSON(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	// 校验AreaID
	if _, err = entity.GetAreaByID(id); err != nil {
		return
	}

	if !entity.IsOwnerOfArea(session.Get(c).UserID, id) {
		err = errors.New(status.Deny)
		return
	}

	isDelCloudDiskFile := req.IsDelCloudDisk != nil && *req.IsDelCloudDisk
	resp, err = ProcessDelArea(c, id, isDelCloudDiskFile)
	return
}

func ProcessDelArea(c *gin.Context, areaID uint64, isDelCloudDiskFile bool) (resp clouddisk.DelAreaStatus, err error){
	if !extension.HasExtension(types.CloudDisk) {
		err = clouddisk.DelArea(areaID)
		resp.RemoveStatus = clouddisk.CloudDiskDelSuccess
		return
	}

	// FIXME 云端没有网盘
	result, err := clouddisk.DelAreaCloudDisk(c, isDelCloudDiskFile, areaID)
	if err != nil {
		return
	}
	resp.RemoveStatus = result.Data.RemoveStatus
	return
}