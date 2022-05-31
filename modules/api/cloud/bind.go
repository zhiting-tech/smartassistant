package cloud

import (
	"fmt"
	"net/http"
	"strconv"

	setting2 "github.com/zhiting-tech/smartassistant/modules/api/setting"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// bindCloudReq 绑定云端接口请求参数
type bindCloudReq struct {
	CloudUserID int    `json:"cloud_user_id"`
	AccessToken string `json:"access_token"`
}

type bindCloudResp struct {
	AreaID string `json:"area_id"` // 该AreaID用于客户端更新自己SC的家庭ID数据
}

// bindCloud 用于处理绑定云端接口的请求
func bindCloud(c *gin.Context) {

	var (
		req  bindCloudReq
		resp bindCloudResp
		area entity.Area
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	// 更新用户和家庭关系
	path := fmt.Sprintf("users/%d", req.CloudUserID)
	u := session.Get(c)
	body := map[string]interface{}{
		"access_token": req.AccessToken,
		"sa_user_id":   u.UserID,
		"sa_area_id":   u.AreaID,
	}

	area, err = entity.GetAreaByID(u.AreaID)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	body["area_name"] = area.Name
	body["area_type"] = area.AreaType

	setting := entity.GetDefaultUserCredentialFoundSetting()
	if err = entity.GetSetting(entity.UserCredentialFoundType, &setting, u.AreaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 判断是否允许找回找回凭证
	if setting.UserCredentialFound {
		token, err := setting2.GetAreaAuthToken(u.AreaID)
		if err != nil {
			return
		}
		body["area_token"] = token
	}

	_, err = cloud.DoWithContext(c.Request.Context(), path, http.MethodPost, body)
	if err != nil {
		err = errors.New(status.SABindError)
		return
	}

	err = SetAreaSynced(u.AreaID)
	if err != nil {
		return
	}
	resp.AreaID = strconv.FormatUint(u.AreaID, 10)
}
