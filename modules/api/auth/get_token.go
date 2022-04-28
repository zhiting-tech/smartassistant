package auth

import (
	errors2 "errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/oauth2.v3"
	errors3 "gopkg.in/oauth2.v3/errors"
	"gorm.io/gorm"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth/generate"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/file"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/hash"
	"github.com/zhiting-tech/smartassistant/modules/utils/jwt"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type GetTokenReq struct {
	GrantType    string `json:"grant_type"` // 授权类型
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`

	Code string `json:"code"` // grant type为authorization_code时使用

	AccountName string `json:"account_name"` // 密码授权模式
	Password    string `json:"password"`     // 密码授权模式

	RefreshToken string `json:"refresh_token"` // 刷新token

	AreaID uint64   `json:"area_id"`
	Scopes []string `json:"scopes"`
}

type GetTokenResp struct {
	TokenInfo TokenInfo       `json:"token_info"`
	UserInfo  entity.UserInfo `json:"user_info,omitempty"`
}

type TokenInfo struct {
	AccessToken     string `json:"access_token" `
	AccessTokenExp  int64  `json:"access_token_exp"`
	RefreshToken    string `json:"refresh_token"`
	RefreshTokenExp int64  `json:"refresh_token_exp"`
}

func GetToken(c *gin.Context) {
	var (
		err  error
		resp GetTokenResp
		req  GetTokenReq
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	gt, tgr, err := req.HandleTokenRequest(c)
	if err != nil {
		return
	}

	ti, err := oauth.GetOauthServer().GetAccessToken(gt, tgr)
	if err != nil {
		logger.Errorf("get token failed: (%v)\n", err)
		if err.Error() == jwt.ErrTokenIsExpired.Error() {
			err = errors.Wrap(err, status.ErrRefreshTokenExpired)
			return
		}

		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	resp.TokenInfo = TokenInfo{
		AccessToken:     ti.GetAccess(),
		AccessTokenExp:  int64(ti.GetAccessExpiresIn().Seconds()),
		RefreshToken:    ti.GetRefresh(),
		RefreshTokenExp: int64(ti.GetRefreshExpiresIn().Seconds()),
	}

	if req.GrantType == string(oauth2.PasswordCredentials) || req.GrantType == string(oauth2.Refreshing) {
		var u entity.User
		u, err = GetUserByToken(resp.TokenInfo.AccessToken)
		if err != nil {
			return
		}
		imgUrl, _ := file.GetFileUrl(c, u.AvatarID)
		resp.UserInfo = entity.UserInfo{
			UserId:        u.ID,
			AccountName:   u.AccountName,
			Nickname:      u.Nickname,
			IsSetPassword: u.Password != "",
			Token:         resp.TokenInfo.AccessToken,
			AvatarUrl:     imgUrl,
		}
	}
}

// HandleTokenRequest 处理token请求
func (req *GetTokenReq) HandleTokenRequest(c *gin.Context) (oauth2.GrantType, *oauth2.TokenGenerateRequest, error) {
	gt := oauth2.GrantType(req.GrantType)
	if gt == "" {
		return "", nil, errors3.ErrInvalidGrant
	}

	tgr := &oauth2.TokenGenerateRequest{
		Request:      c.Request,
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret,
		Scope:        strings.Join(req.Scopes, ","),
	}

	var areaID uint64
	switch gt {
	case oauth2.AuthorizationCode:
		if req.Code == "" {
			return "", nil, errors.New(errors.BadRequest)
		}
		tgr.Code = req.Code
		claims, err := generate.DecodeJwt(tgr.Code)
		if err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return "", nil, err
		}
		areaID = claims.AreaID
	case oauth2.PasswordCredentials:
		u, err := req.passwordAuthorizeHandler()
		if err != nil {
			return "", nil, err
		}
		areaID = u.AreaID
		tgr.UserID = strconv.Itoa(u.ID)

		client, err := entity.GetSAClient(areaID)
		if err != nil {
			return "", nil, err
		}
		tgr.ClientID = client.ClientID
		tgr.ClientSecret = client.ClientSecret
		tgr.Scope = client.AllowScope
	case oauth2.ClientCredentials:
		areaID = req.AreaID
	case oauth2.Refreshing:
		claims, err := generate.DecodeJwt(req.RefreshToken)
		if err != nil {
			err = errors.Wrap(err, errors.InternalServerErr)
			return "", nil, err
		}
		clientInfo, err := entity.GetClientByClientID(claims.ClientID)
		if err != nil {
			return "", nil, err
		}

		tgr.ClientSecret = clientInfo.ClientSecret
		tgr.ClientID = clientInfo.ClientID
		tgr.Refresh = req.RefreshToken

		areaID = claims.AreaID
	}

	if areaID == 0 {
		err := errors.New(errors.BadRequest)
		return "", nil, err
	}
	tgr.Request.Header.Set(types.GrantType, string(gt))
	return gt, tgr, nil
}

// passwordAuthorizeHandler 验证用户名密码
func (req GetTokenReq) passwordAuthorizeHandler() (u entity.User, err error) {
	if req.AccountName == "" || req.Password == "" {
		err = errors.New(errors.BadRequest)
		return
	}
	u, err = entity.GetUserByAccountName(req.AccountName)
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			err = errors.Wrap(err, status.AccountNotExistErr)
			return
		}
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	// 校验密码是否正确
	if !hash.CheckPassword(req.Password, u.Salt, u.Password) {
		err = errors.New(status.AccountPassWordErr)
		return
	}
	return
}

func GetUserByToken(accessToken string) (u entity.User, err error) {
	ti, err := oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
	if err != nil {
		var uerr = errors.Wrap(err, status.UserNotExist)
		if err.Error() == uerr.Error() {
			return u, uerr
		}

		if err.Error() == jwt.ErrTokenIsExpired.Error() {
			return u, errors.Wrap(err, status.ErrAccessTokenExpired)
		}

		return u, errors.Wrap(err, status.RequireLogin)
	}

	uid, _ := strconv.Atoi(ti.GetUserID())
	u, err = entity.GetUserByID(uid)
	if err != nil {
		return
	}
	return u, err
}
