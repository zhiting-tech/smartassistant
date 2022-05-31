package plugin

import (
	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/brand"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
)

type listType int

const listTypeAll listType = 0
const listTypeDevelop listType = 1

type Req struct {
	ListType int `form:"list_type"`
}

type Resp struct {
	Plugins []Plugin `json:"plugins"`
}

type Plugin struct {
	brand.Plugin
	BuildStatus int `json:"build_status"` // build状态，-1 build失败,0正在build,1 build成功
}

func ListPlugin(c *gin.Context) {
	var (
		err  error
		req  Req
		resp Resp
	)
	resp.Plugins = make([]Plugin, 0)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindQuery(&req); err != nil {
		return
	}
	u := session.Get(c)
	var ps []entity.PluginInfo
	switch listType(req.ListType) {
	case listTypeAll:
		ps, err = entity.GetInstalledPlugins()
		if err != nil {
			return
		}
	case listTypeDevelop:
		ps, err = entity.GetDevelopPlugins(u.AreaID)
		if err != nil {
			return
		}
	default:

	}

	if len(ps) == 0 {
		return
	}
	for _, plg := range ps {
		p := Plugin{
			Plugin: brand.Plugin{
				Name:    plg.PluginID,
				Brand:   plg.Brand,
				Version: plg.Version,
				ID:      plg.PluginID,
				Info:    plg.Info,
				IsAdded: true,
			},
			BuildStatus: plg.Status,
		}
		resp.Plugins = append(resp.Plugins, p)
	}
}
