package supervisor

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type updateReq struct {
	Version string `json:"version"`
}

// Update 更新SA到新版
func Update(c *gin.Context) {
	var (
		req updateReq
		err error
	)
	defer func() {
		response.HandleResponse(c, err, nil)
	}()
	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	err = supervisor.GetManager().StartUpdateJobWithContext(c.Request.Context(), req.Version)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
	}
}
