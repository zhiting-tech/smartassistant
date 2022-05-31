package brand

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/cloud"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// 0,获取所有品牌;1,获取已添加的品牌列表
const (
	typeTotal = 0
	typeAdded = 1
)

// BrandInfo 品牌信息
type BrandInfo struct {
	cloud.Brand
	IsAdded  bool `json:"is_added"`  // 是否已添加
	IsNewest bool `json:"is_newest"` // 是否是最新
}

// Resp 品牌列表接口返回数据
type Resp struct {
	Brands []BrandInfo `json:"brands"`
}

// listBrandsReq 品牌列表接口请求参数
type listBrandsReq struct {
	Type int `form:"type"` // 0全部1已安装
}

// ListAddedBrands 获取已添加插件的品牌，不请求SC
func ListAddedBrands() (brandInfos []BrandInfo, err error) {

	var installedPlgs []entity.PluginInfo
	installedPlgs, err = entity.GetInstalledPlugins()
	if err != nil {
		return
	}

	brandInfoMap := make(map[string]BrandInfo)
	for _, plg := range installedPlgs {
		brand, ok := brandInfoMap[plg.Brand]
		if ok {
			continue
		}
		brand = BrandInfo{
			Brand: cloud.Brand{
				LogoURL:      "",
				Name:         plg.Brand,
				PluginAmount: 0,
			},
			IsAdded:  true,
			IsNewest: false, // TODO 品牌是否最新判断逻辑调整
		}
		brandInfoMap[plg.Brand] = brand
	}

	for _, brandInfo := range brandInfoMap {
		brandInfos = append(brandInfos, brandInfo)
	}
	return
}

// ListBrandsWithContext 获取所有品牌
func ListBrandsWithContext(ctx context.Context) (brandInfos []BrandInfo, err error) {

	var brands []cloud.Brand
	brands, err = cloud.GetBrandsWithContext(ctx)
	if err != nil {
		return
	}

	// 获取所有已安装插件
	installedPlugins, err := entity.GetInstalledPlugins()
	if err != nil {
		return
	}
	// 获取所有已安装插件的品牌
	brandInstallMap := make(map[string]bool)
	for _, plg := range installedPlugins {
		brandInstallMap[plg.Brand] = true
	}

	for _, b := range brands {
		bi := BrandInfo{
			Brand: b,
		}
		if _, ok := brandInstallMap[b.Name]; ok {
			bi.IsAdded = true
		}
		brandInfos = append(brandInfos, bi)
	}
	return
}

// List 用于处理品牌列表接口的请求
func List(c *gin.Context) {
	var (
		resp Resp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()
	var req listBrandsReq
	if err = c.ShouldBind(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if req.Type == typeAdded {
		resp.Brands, err = ListAddedBrands()
	} else {
		resp.Brands, err = ListBrandsWithContext(c.Request.Context())
	}
}
