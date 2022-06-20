package device

import (
	errors2 "errors"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gorm.io/gorm"
)

//
type deviceLogoReq struct {
	DeviceID int `uri:"id" binding:"required"`
}

type deviceLogosResp struct {
	DeviceLogos []deviceLogoInfo `json:"device_logos"`
}

type deviceLogoInfo struct {
	Type types.LogoType `json:"type"`
	Name string         `json:"name"`
	Url  string         `json:"url"`
}

// InfoDeviceLogo 处理获取设备图标列表接口
func InfoDeviceLogo(c *gin.Context) {
	var (
		err        error
		deviceInfo entity.Device
		resp       deviceLogosResp
		req        deviceLogoReq
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindUri(&req); err != nil {
		return
	}
	if deviceInfo, err = entity.GetDeviceByID(req.DeviceID); err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.NotFound)
		} else {
			err = errors.Wrap(err, errors.InternalServerErr)
		}
		return
	}

	resp.DeviceLogos = append(resp.DeviceLogos, deviceLogoInfo{
		Name: "设备图片",
		Url:  plugin.DeviceLogoURL(c.Request, deviceInfo.PluginID, deviceInfo.Model, deviceInfo.Type),
	})

	for _, l := range types.DeviceLogos {
		resp.DeviceLogos = append(resp.DeviceLogos, deviceLogoInfo{
			Type: l.LogoType,
			Name: l.Name,
			Url:  url.ImageUrl(c.Request, l.FileName),
		})
	}
	return
}
