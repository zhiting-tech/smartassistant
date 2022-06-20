package brand

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	version2 "github.com/zhiting-tech/smartassistant/modules/utils/version"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// Brand 品牌信息
type Brand struct {
	cloud.Brand
	Plugins  []Plugin `json:"plugins"`
	IsAdded  bool     `json:"is_added"`  // 是否已添加
	IsNewest bool     `json:"is_newest"` // 是否是最新
}

type Plugin struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Brand    string `json:"brand"`
	Info     string `json:"info"`
	IsAdded  bool   `json:"is_added"`
	IsNewest bool   `json:"is_newest"`
}

// brandInfoReq 品牌详情接口请求参数
type brandInfoReq struct {
	Name string `uri:"name"`
}

// GetBrandInfoWithContext 获取品牌详情
func GetBrandInfoWithContext(ctx context.Context, name string) (brand Brand, err error) {
	brand.Plugins = make([]Plugin, 0)

	brand = Brand{
		Brand: cloud.Brand{Name: name},
	}

	var installedPlgs []entity.PluginInfo
	installedPlgs, err = entity.GetInstalledPlugins()
	if err != nil {
		return
	}

	installedPlgMap := make(map[string]entity.PluginInfo)
	for _, p := range installedPlgs {
		installedPlgMap[p.PluginID] = p
	}
	brandInfo, err := cloud.GetBrandInfoWithContext(ctx, name)
	// 请求sc失败则读取本地信息
	if err != nil {
		logger.Error(err)

		for _, p := range installedPlgs {
			pp := Plugin{
				ID:       p.PluginID,
				Version:  p.Version,
				Brand:    p.Brand,
				IsAdded:  false,
				IsNewest: false,
			}
			brand.IsNewest = true
			brand.IsAdded = true
			brand.Plugins = append(brand.Plugins, pp)
		}
		return
	}
	brand.LogoURL = brandInfo.LogoURL
	for _, p := range brandInfo.Plugins {
		pp := Plugin{
			ID:      p.UID,
			Name:    p.Name,
			Version: p.Version,
			Brand:   p.Brand,
			Info:    p.Intro,
		}
		_, pp.IsAdded = installedPlgMap[p.UID]
		if pp.IsAdded {
			// 判断已安装版本是否大于等于最新版本
			pp.IsNewest, _ = version2.GreaterOrEqual(installedPlgMap[p.UID].Version, p.Version)
			brand.IsAdded = true
		}
		brand.Plugins = append(brand.Plugins, pp)
	}

	brand.PluginAmount = len(brand.Plugins)
	return
}

// InfoResp 品牌详情接口返回数据
type InfoResp struct {
	Brand Brand `json:"brand"`
}

// Info 用于处理品牌详情接口的请求
func Info(c *gin.Context) {
	var (
		req  brandInfoReq
		resp InfoResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindUri(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	var brand Brand
	if brand, err = GetBrandInfoWithContext(c.Request.Context(), req.Name); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	} else {
		resp.Brand = brand
	}
}
