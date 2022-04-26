package app

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
)

// ListAppResp 第三方平台列表接口返回列表
type ListAppResp struct {
	Apps []cloud.App `json:"apps"`
}

// ListApp 第三方平台列表
func ListApp(c *gin.Context) {
	var (
		err  error
		resp ListAppResp
	)
	resp.Apps = make([]cloud.App, 0)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	u := session.Get(c)
	apps, err := cloud.GetAppList(c.Request.Context(), u.AreaID)
	if err != nil {
		return
	}
	resp.Apps = apps

	return
}
