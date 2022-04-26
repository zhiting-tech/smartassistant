package brand

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/analytics"
)

type handlePluginsReq struct {
	BrandName string   `uri:"brand_name" binding:"required"`
	Plugins   []string `json:"plugins"`
}

type handlePluginsResp struct {
	SuccessPlugins []string `json:"success_plugins"`
}

func (req handlePluginsReq) GetPluginsWithContext(ctx context.Context) (plugins []*plugin.Plugin, err error) {

	if len(req.Plugins) == 0 { // 没有指定插件时，更新品牌所有插件
		var plgs map[string]*plugin.Plugin
		plgs, err = plugin.GetGlobalManager().LoadPluginsWithContext(ctx)
		if err != nil {
			return
		}
		for _, plg := range plgs {
			if plg.Brand != req.BrandName {
				continue
			}
			plugins = append(plugins, plg)
		}
	} else {
		for _, pluginID := range req.Plugins {
			var plg *plugin.Plugin
			plg, err = plugin.GetGlobalManager().GetPluginWithContext(ctx, pluginID)
			if err != nil {
				return
			}
			plugins = append(plugins, plg)
		}
	}
	return
}

func UpdatePlugin(c *gin.Context) {

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
		if err = plg.UpdateOrInstall(); err != nil {
			return
		}
		go analytics.RecordStruct(analytics.EventTypePluginAdd, user.UserID, plg)
		resp.SuccessPlugins = append(resp.SuccessPlugins, plg.ID)
	}
	return
}
