package brand

import (
	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/analytics"
	"github.com/zhiting-tech/smartassistant/pkg/event"
)

func DelPlugins(c *gin.Context) {
	var (
		req  handlePluginsReq
		resp handlePluginsResp
		err  error
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()
	if err = c.BindUri(&req); err != nil {
		return
	}
	if err = c.BindJSON(&req); err != nil {
		return
	}
	plgs, err := req.GetPluginsWithContext(c.Request.Context())
	if err != nil {
		return
	}

	user := session.Get(c)
	for _, plg := range plgs {
		if plg.Brand != req.BrandName {
			continue
		}
		if err = plg.Remove(c.Request.Context()); err != nil {
			return
		}
		go analytics.RecordStruct(analytics.EventTypePluginDelete, user.UserID, plg)
		resp.SuccessPlugins = append(resp.SuccessPlugins, plg.ID)
	}
	event.Notify(event.NewEventMessage(event.DeviceDecrease, session.Get(c).AreaID))
	return
}
