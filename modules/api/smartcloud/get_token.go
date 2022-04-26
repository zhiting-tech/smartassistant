package smartcloud

import (
	"strconv"
	"strings"

	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/server"

	"github.com/zhiting-tech/smartassistant/modules/api/scope"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/pkg/logger"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/hash"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

const (
	sAUserTokenType        = 1 // sa用户的token类型标识
	cloudDiskUserTokenType = 2 // SA用户的网盘token类型标识
)

type GetTokenResp struct {
	Token string `json:"token"`
}

func updateToken(userID int, areaID uint64, c *gin.Context) (resp GetTokenResp, err error) {
	key := hash.GetSaUserKey()
	var u = entity.User{Key: key, ID: userID, AreaID: areaID}
	if err = entity.EditUser(userID, u); err != nil {
		return
	}
	token, err := oauth.GetSAUserToken(u, c.Request)
	if err != nil {
		return
	}
	resp = GetTokenResp{
		Token: token,
	}
	return
}

type GetTokenReq struct {
	Type   int `uri:"type"`
	UserID int `uri:"id"`
}

// 获取找回用户凭证
func GetToken(c *gin.Context) {
	var (
		err  error
		resp GetTokenResp
		req  GetTokenReq
	)

	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	if err = c.BindUri(&req); err != nil {
		err = errors.Wrap(err, errors.BadRequest)
		return
	}

	userInfo, err := entity.GetUserByID(req.UserID)
	if err != nil {
		err = errors.Wrap(err, status.AccountNotExistErr)
		return
	}

	switch req.Type {
	case sAUserTokenType:
		resp, err = UpdateSAUserToken(userInfo, c)
	case cloudDiskUserTokenType:
		resp, err = GetCloudDiskToken(userInfo, c)
	default:
		err = errors.New(errors.BadRequest)
	}
}

func UpdateSAUserToken(userInfo entity.User, c *gin.Context) (resp GetTokenResp, err error) {
	// 获取配置
	setting := entity.GetDefaultUserCredentialFoundSetting()
	if err = entity.GetSetting(entity.UserCredentialFoundType, &setting, userInfo.AreaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 判断是否允许找回找回凭证
	if !setting.UserCredentialFound {
		err = errors.New(status.GetUserTokenDeny)
		return
	}
	resp, err = updateToken(userInfo.ID, userInfo.AreaID, c)
	return
}

func GetCloudDiskToken(userInfo entity.User, c *gin.Context) (resp GetTokenResp, err error) {
	// 获取配置
	setting := entity.GetDefaultCloudDiskCredentialSetting()
	if err = entity.GetSetting(entity.GetCloudDiskCredential, &setting, userInfo.AreaID); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	// 判断是否允许找回找回凭证
	if !setting.GetCloudDiskCredentialSetting {
		err = errors.New(status.GetCloudDiskTokenDeny)
		return
	}
	resp, err = getCloudDiskToken(userInfo, c)
	return

}

func getCloudDiskToken(userInfo entity.User, c *gin.Context) (resp GetTokenResp, err error) {
	saClient, err := entity.GetSAClient(userInfo.AreaID)
	if err != nil {
		return
	}

	tgr := &server.AuthorizeRequest{
		ResponseType:   oauth2.Token,
		ClientID:       saClient.ClientID,
		UserID:         strconv.Itoa(userInfo.ID),
		Scope:          strings.Join([]string{"user", "area"}, ","),
		AccessTokenExp: scope.ExpiresIn,
		Request:        c.Request,
	}

	tokenInfo, err := oauth.GetOauthServer().GetAuthorizeToken(tgr)
	if err != nil {
		logger.Errorf("get oauth2 token error %s", err.Error())
		err = errors.Wrap(err, errors.BadRequest)
		return
	}
	resp.Token = tokenInfo.GetAccess()
	return
}
