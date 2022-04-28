package oauth

import (
	"net/http"
	"strconv"
	"time"

	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/server"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// GetUserPluginToken 获取插件token，限制有效期和权限范围
func GetUserPluginToken(userID int, req *http.Request, areaID uint64) (token string, err error) {

	saClient, _ := entity.GetSAClient(areaID) // TODO 考虑将扩展和插件作为client管理起来

	authReq := &server.AuthorizeRequest{
		ResponseType:   oauth2.Token,
		ClientID:       saClient.ClientID,
		Scope:          types.ScopePlugin.Scope,
		UserID:         strconv.Itoa(userID),
		AccessTokenExp: 2 * time.Hour,
		Request:        req,
	}

	ti, err := GetOauthServer().GetAuthorizeToken(authReq)
	if err != nil {
		logger.Errorf("GetAuthorizeToken err: %s", err)
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return ti.GetAccess(), nil
}

// GetSAUserToken 获取SA用户Token,提供给添加sa设备，扫码加入使用
func GetSAUserToken(user entity.User, req *http.Request) (token string, err error) {
	saClient, _ := entity.GetSAClient(user.AreaID)

	authReq := &server.AuthorizeRequest{
		ResponseType: oauth2.Token,
		ClientID:     saClient.ClientID,
		Scope:        saClient.AllowScope,
		UserID:       strconv.Itoa(user.ID),
		Request:      req,
	}

	ti, err := GetOauthServer().GetAuthorizeToken(authReq)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	return ti.GetAccess(), nil
}
