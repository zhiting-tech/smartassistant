package app

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

type UnbindAppReq struct {
	AppID  int    `uri:"id"`
	AreaID uint64 `uri:"area_id"`
}

// UnbindApp 用于处理解绑云端接口的请求
func UnbindApp(c *gin.Context) {
	var (
		req UnbindAppReq
		err error
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	if err = c.BindUri(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	u := session.Get(c)
	if err = req.validate(u); err != nil {
		return
	}

	if err = cloud.UnbindApp(c.Request.Context(), u.AreaID, req.AppID); err != nil {
		return
	}
}

func (req UnbindAppReq) validate(u *session.User) (err error) {
	if !u.IsOwner {
		err = errors.New(status.Deny)
		return
	}

	if req.AreaID != u.AreaID {
		err = errors.New(status.Deny)
		return
	}

	return
}
