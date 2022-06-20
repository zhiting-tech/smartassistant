package device

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-unidecode"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/thingmodel"
)

// deviceAddReq 添加设备接口请求参数
type deviceAddReq struct {
	Device dummyDevice `json:"device"`
}

// device
type dummyDevice struct {
	Name         string            `json:"name"`
	PluginID     string            `json:"plugin_id"`
	IID          string            `json:"iid"`
	Model        string            `json:"model"`        // 型号
	Manufacturer string            `json:"manufacturer"` // 制造商
	Type         plugin.DeviceType `json:"type"`         // 设备类型，如：light,switch...
	SyncData     string            `json:"sync_data"`    // 自定义的客户端同步信息

	AreaID uint64 `json:"area_id"`
}

// deviceAddResp 添加设备返回数据
type deviceAddResp struct {
	ID      int  `json:"device_id"`
	IsAdded bool `json:"is_added"`
}

func addDummyDevice(c *gin.Context) {
	var (
		err  error
		req  deviceAddReq
		resp deviceAddResp
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindJSON(&req); err != nil {
		return
	}
	if req.Device.Name == "" ||
		req.Device.IID == "" ||
		req.Device.PluginID == "" ||
		req.Device.Model == "" ||
		req.Device.Manufacturer == "" ||
		req.Device.Type == "" {
		err = errors.New(status.ParamRequireErr)
		return
	}

	u := session.Get(c)
	d := entity.Device{
		Name:         req.Device.Name,
		Pinyin:       unidecode.Unidecode(req.Device.Name),
		PluginID:     req.Device.PluginID,
		IID:          req.Device.IID,
		Model:        req.Device.Model,
		Manufacturer: req.Device.Manufacturer,
		Type:         req.Device.Type.String(),
		SyncData:     req.Device.SyncData,
		AreaID:       u.AreaID,
	}
	logoType := int(device.TypeToLogoType(req.Device.Type))
	d.LogoType = &logoType
	d.ThingModel, _ = json.Marshal(thingmodel.ThingModel{})
	isExist, err := entity.IsDeviceExist(u.AreaID, req.Device.PluginID, req.Device.IID)
	if err != nil {
		return
	}
	if err = entity.CreateDevice(&d, entity.GetDB()); err != nil {
		return
	}

	// 通过触发health check来更新物模型
	if req.Device.PluginID != "" {
		identify := plugin.Identify{
			PluginID: req.Device.PluginID,
			IID:      req.Device.IID,
			AreaID:   u.AreaID,
		}
		plugin.GetGlobalClient().IsOnline(identify)
	}
	resp.ID = d.ID
	if isExist {
		resp.IsAdded = true
		return
	} else {
		if err = device.AddDevicePermissionForRoles(d, entity.GetDB()); err != nil {
			return
		}
	}
}
