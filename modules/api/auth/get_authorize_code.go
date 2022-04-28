package auth

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/server"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type GetAuthorizeCodeReq struct {
	ClientID     string   `json:"client_id"`
	ResponseType string   `json:"response_type"` // 此处应固定填写为code
	State        string   `json:"state"`         // 第三方指定任意值
	Scopes       []string `json:"scope"`         // 获取的权限，可选
}

type GetAuthorizeCodeResp struct {
	Code string `json:"code"`
}

func GetAuthorizeCode(c *gin.Context) {
	var (
		req  GetAuthorizeCodeReq
		resp GetAuthorizeCodeResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, &resp)
	}()

	if err = c.BindJSON(&req); err != nil {
		err = errors.New(errors.BadRequest)
		return
	}

	userInfo := session.Get(c)
	authReq := &server.AuthorizeRequest{
		ResponseType: oauth2.ResponseType(req.ResponseType),
		ClientID:     req.ClientID,
		Scope:        strings.Join(req.Scopes, ","),
		State:        req.State,
		UserID:       strconv.Itoa(userInfo.UserID),
		Request:      c.Request,
	}
	ti, err := oauth.GetOauthServer().GetAuthorizeToken(authReq)
	if err != nil {
		logger.Error("get authorize code err: %v", err)
		return
	}
	resp.Code = ti.GetCode()
}
