package plugin

import (
	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/event"
)

type delPluginReq struct {
	PluginID string `uri:"id"`
}

func DelPlugin(c *gin.Context) {
	var (
		err  error
		req  delPluginReq
		resp interface{}
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindUri(&req); err != nil {
		return
	}

	sessionUser := session.Get(c)
	p, err := entity.GetPlugin(req.PluginID, sessionUser.AreaID)
	if err != nil {
		return
	}
	plg := plugin.NewFromEntity(p)
	err = plg.Remove(c.Request.Context())
	event.Notify(event.NewEventMessage(event.DeviceDecrease, sessionUser.AreaID))
}
