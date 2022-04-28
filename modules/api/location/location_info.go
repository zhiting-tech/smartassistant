package location

import (
	"sort"
	"strconv"

	device2 "github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/plugin"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/logger"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// infoResp 房间详情接口返回数据
type infoResp struct {
	Name    string   `json:"name"`
	Devices []Device `json:"devices"`
}

// Device 设备信息
type Device struct {
	ID        int    `json:"id"`
	Logo      string `json:"logo"`
	LogoURL   string `json:"logo_url"`
	Name      string `json:"name"`
	IsSA      bool   `json:"is_sa"`
	Control   string `json:"control"`
	PluginURL string `json:"plugin_url"`
	PluginID  string `json:"plugin_id"`
	Type      string `json:"type"`
}

// InfoLocation 用于处理房间详情接口的请求
func InfoLocation(c *gin.Context) {
	var (
		err         error
		locationId  int
		infoDevices []Device
		resp        infoResp
		location    entity.Location
	)
	defer func() {
		if resp.Devices == nil {
			resp.Devices = make([]Device, 0)
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

func GetDeviceByLocationID(LocationId int, c *gin.Context) (infoDevices []Device, err error) {
	var (
		devices []entity.Device
	)

	devices, err = entity.GetDevicesByLocationID(LocationId)

	if err != nil {
		return
	}
	infoDevices, err = WrapDevices(c, devices)
	if err != nil {
		return
	}

	return
}

func WrapDevices(c *gin.Context, devices []entity.Device) (result []Device, err error) {

	u := session.Get(c)

	for _, d := range devices {
		logoUrl, logo := device2.LogoInfo(c, d)
		dd := Device{
			ID:      d.ID,
			Name:    d.Name,
			Logo:    logo,
			LogoURL: logoUrl,
			Type:    d.Type,
		}
		if d.Model == types.SaModel {
			dd.IsSA = true
			dd.PluginID = ""
		} else {
			dd.PluginID = d.PluginID
			var pluginURL *plugin.URL
			pluginURL, err = plugin.ControlURL(d, c.Request, u.UserID)
			if err != nil {
				logger.Errorf("Get plugin url err: %v\n", err)
				err = nil
				continue
			}
			dd.PluginURL = pluginURL.String()
			dd.Control = pluginURL.PluginPath()
		}
		result = append(result, dd)

	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].IsSA {
			return true
		}
		return false
	})
	return
}
