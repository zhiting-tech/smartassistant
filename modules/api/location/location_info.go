package location

import (
	"github.com/zhiting-tech/smartassistant/modules/api/device"
	"strconv"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// infoResp 房间详情接口返回数据
type infoResp struct {
	Name    string       `json:"name"`
	Devices []InfoDevice `json:"devices"`
}

// InfoDevice 设备信息
type InfoDevice struct {
	ID        int    `json:"id"`
	LogoURL   string `json:"logo_url"`
	Name      string `json:"name"`
	IsSa      bool   `json:"is_sa"`
	PluginURL string `json:"plugin_url"`
	PluginID  string `json:"plugin_id"`
}

// InfoLocation 用于处理房间详情接口的请求
func InfoLocation(c *gin.Context) {
	var (
		err         error
		locationId  int
		infoDevices []InfoDevice
		resp        infoResp
		location    entity.Location
	)
	defer func() {
		if resp.Devices == nil {
			resp.Devices = make([]InfoDevice, 0)
		}
		response.HandleResponse(c, err, resp)
	}()

	locationId, err = strconv.Atoi(c.Param("id"))
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	if location, err = entity.GetLocationByID(locationId); err != nil {
		return
	}

	if infoDevices, err = GetDeviceByLocationID(locationId, c); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	resp.Devices = infoDevices
	resp.Name = location.Name
	return
}

func GetDeviceByLocationID(LocationId int, c *gin.Context) (infoDevices []InfoDevice, err error) {
	var (
		devices []entity.Device
	)

	devices, err = entity.GetDevicesByLocationID(LocationId)

	if err != nil {
		return
	}
	deviceInfos, err := device.WrapDevices(c, devices, device.AllDevice)
	if err != nil {
		return
	}
	for _, di := range deviceInfos {
		infoDevices = append(infoDevices, InfoDevice{
			ID:        di.ID,
			LogoURL:   di.LogoURL,
			Name:      di.Name,
			IsSa:      di.IsSA,
			PluginURL: di.PluginURL,
			PluginID:  di.PluginID,
		})
	}

	return
}
