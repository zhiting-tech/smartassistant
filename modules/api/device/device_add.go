package device

import (
	errors2 "errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/api/area"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/device"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/analytics"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// deviceAddReq 添加设备接口请求参数
type deviceAddReq struct {
	Device deviceInfo `json:"device"` // TODO 校验
}

type deviceInfo struct {
	entity.Device
	AreaType entity.AreaType `json:"area_type"`
}

// deviceAddResp 添加设备接口返回数据
type deviceAddResp struct {
	ID        int    `json:"device_id"`
	PluginURL string `json:"plugin_url"`

	// 添加SA成功时需要的响应
	UserInfo *entity.UserInfo `json:"user_info"` // 创建人的用户信息
	AreaInfo *area.Area       `json:"area_info"` // 家庭信息
}

// AddDevice 用于处理添加设备接口的请求
func AddDevice(c *gin.Context) {
	var (
		req  deviceAddReq
		resp deviceAddResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	err = c.BindJSON(&req)
	if err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	var uid int
	if req.Device.Model == types.SaModel {
		var (
			userInfo entity.UserInfo
			areaInfo area.Area
		)
		if userInfo, areaInfo, err = addSADevice(&req.Device.Device, c, req.Device.AreaType); err != nil {
			return
		} else {
			resp.UserInfo = &userInfo
			resp.AreaInfo = &areaInfo
			uid = userInfo.UserId
		}
	} else {
		err = errors.Wrap(errors2.New("invalid sa"), errors.BadRequest)
		return
	}
	// 记录添加设备信息
	go analytics.RecordStruct(analytics.EventTypeDeviceAdd, uid, req.Device.Device)
	resp.ID = req.Device.ID
	return
}

func addSADevice(sa *entity.Device, c *gin.Context, areaType entity.AreaType) (userInfo entity.UserInfo, areaInfo area.Area, err error) {

	// 判断SA是否存在
	_, err = entity.GetSaDevice()
	if err == nil {
		err = errors.Wrap(err, status.SaDeviceAlreadyBind)
		return
	} else {
		if !errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, errors.InternalServerErr)
			return
		}
	}
	if !entity.IsAreaType(areaType) {
		err = errors.New(errors.BadRequest)
		return
	}

	var areaObj entity.Area
	areaObj, err = entity.CreateArea("", areaType)
	if err != nil {
		return
	}
	areaID := areaObj.ID
	sa.CreatedAt = time.Now()
	if err = device.Create(areaID, sa); err != nil {
		return
	}
	areaObj, err = entity.GetAreaByID(areaID)
	if err != nil {
		return
	}
	var user entity.User
	if user, err = entity.GetUserByID(areaObj.OwnerID); err != nil {
		return
	}

	// 初始化client
	if err = entity.InitClient(areaID); err != nil {
		return
	}

	token, err := oauth.GetSAUserToken(user, c.Request)
	if err != nil {
		return
	}

	// 设备添加成功后需要获取Creator信息
	userInfo = entity.UserInfo{
		UserId:        user.ID,
		Nickname:      user.Nickname,
		IsSetPassword: user.Password != "",
		Token:         token,
	}
	areaInfo = area.Area{
		ID: strconv.FormatUint(sa.AreaID, 10),
	}
	return
}
