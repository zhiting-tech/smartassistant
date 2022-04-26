package plugin

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// PluginInfoResp 插件详情接口返回数据
type PluginInfoResp struct {
	Plugin PluginInfo `json:"plugin"`
}

type PluginInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Version     string `json:"version"`
	Brand       string `json:"brand"`
	Info        string `json:"info"`
	IsAdded     bool   `json:"is_added"`
	IsNewest    bool   `json:"is_newest"`
	DownloadURL string `json:"download_url"` // 前端插件压缩包？
}

// PluginInfoReq 插件详情接口请求参数
type PluginInfoReq struct {
	PluginID string `uri:"id"`
}

// GetPluginInfo 用于处理插件详情接口的请求
func GetPluginInfo(c *gin.Context) {
	var (
		err  error
		req  PluginInfoReq
		resp PluginInfoResp
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.BindUri(&req); err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	var plg plugin.Plugin
	isDevelopPlugin := entity.IsPluginDevelop(req.PluginID, session.Get(c).AreaID)

	if isDevelopPlugin {
		plg, err = getPlugin(req.PluginID, session.Get(c).AreaID)
		if err != nil {
			return
		}
	} else {
		// 系统插件从SC获取数据，失败则用本地数据
		var p *plugin.Plugin
		p, err = plugin.GetGlobalManager().GetPluginWithContext(c.Request.Context(), req.PluginID)
		if err != nil {
			plg, err = getPlugin(req.PluginID, session.Get(c).AreaID)
			if err != nil {
				return
			}
		} else {
			plg = *p
		}
	}
	resp.Plugin = PluginInfo{
		ID:      plg.ID,
		Info:    plg.Info,
		Name:    plg.Name,
		Version: plg.Version,
		Brand:   plg.Brand,
		IsAdded: plg.IsAdded(), IsNewest: plg.IsNewest()}
	resp.Plugin.DownloadURL = plugin.ArchiveURL(plg.ID, c.Request)
}

func getPlugin(pluginID string, areaID uint64) (plg plugin.Plugin, err error) {
	pi, err := entity.GetPlugin(pluginID, areaID)
	if err != nil {
		return
	}
	plg = plugin.NewFromEntity(pi)
	return
}
