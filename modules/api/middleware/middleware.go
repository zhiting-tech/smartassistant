// Package middleware GIN 框架中间件
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"gopkg.in/oauth2.v3"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/utils/jwt"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/reverseproxy"

	"github.com/gin-gonic/gin"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// RequireAccount 用户需要登录才可访问对应的接口
func RequireAccount(c *gin.Context) {
	if _, err := verifyAccessToken(c); err != nil {
		response.HandleResponse(c, err, nil)
		c.Abort()
		return
	}
}

// RequireAccountWithScope 检查是否登录，并且是否有权限
func RequireAccountWithScope(scope types.Scope) func(ctx *gin.Context) {
	return func(c *gin.Context) {

		// 校验token
		ti, err := verifyAccessToken(c)
		if err != nil {
			response.HandleResponse(c, err, nil)
			c.Abort()
			return
		}
		if ti.GetScope() == types.ScopeAll.Scope {
			c.Next()
			return
		}
		// 校验scope
		if !strings.Contains(ti.GetScope(), scope.Scope) {
			err = fmt.Errorf("permission deny: invalid scope: %s, require: %s",
				ti.GetScope(), scope.Scope)
			response.HandleResponse(c, err, nil)
			c.Abort()
			return
		}
		c.Next()
		return
	}
}

func verifyAccessToken(c *gin.Context) (ti oauth2.TokenInfo, err error) {
	accessToken := c.GetHeader(types.SATokenKey)
	ti, err = oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
	if err != nil {
		var uerr = errors.New(status.UserNotExist)
		if err.Error() == uerr.Error() {
			return ti, uerr
		}

		if err.Error() == jwt.ErrTokenIsExpired.Error() {
			return ti, errors.New(status.ErrAccessTokenExpired)
		}

		var perr = errors.New(status.PasswordChanged)
		if err.Error() == perr.Error() {
			return ti, perr
		}

		return ti, errors.Wrap(err, status.InvalidUserCredentials)
	}
	return
}

// RequireOwner 拥有者才能访问
func RequireOwner(c *gin.Context) {
	u := session.Get(c)
	if u == nil {
		response.HandleResponse(c, errors.New(status.InvalidUserCredentials), nil)
		c.Abort()
		return
	}
	if u.IsOwner {
		c.Next()
		return
	}
	response.HandleResponse(c, errors.New(status.Deny), nil)
	c.Abort()
	return
}

// RequireToken 使用token验证身份，不依赖cookies.
func RequireToken(c *gin.Context) {

	uToken := c.Request.Header.Get(types.SATokenKey)
	queryToken := c.Query("token")
	if queryToken != "" && uToken == "" {
		c.Request.Header.Add(types.SATokenKey, queryToken)
	}

	if session.GetUserByToken(c) != nil {
		c.Next()
		return
	}
	response.HandleResponse(c, errors.New(status.InvalidUserCredentials), nil)
	c.Abort()
	return
}

// RequirePermission 判断是否有权限, 如果有多个权限判断表示的是满足其中一个权限就行
func RequirePermission(permissions ...types.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := session.Get(c)
		if u != nil {
			for _, p := range permissions {
				if entity.JudgePermit(u.UserID, p) {
					c.Next()
					return
				}
			}
		}
		err := errors.New(status.Deny)
		response.HandleResponse(c, err, nil)
		c.Abort()

	}
}

// ProxyToPlugin 根据路径转发到后端插件
func ProxyToPlugin(ctx *gin.Context) {

	path := ctx.Param("plugin")
	if up, err := reverseproxy.GetManager().GetUpstream(path); err != nil {
		response.HandleResponseWithStatus(ctx, http.StatusBadGateway, err, nil)
	} else {

		req := ctx.Request.Clone(context.Background())

		// 替换插件请求路径
		oldPrefix := fmt.Sprintf("plugin/%s", path)
		newPrefix := fmt.Sprintf("api/plugin/%s", path)
		req.URL.Path = strings.Replace(req.URL.Path, oldPrefix, newPrefix, 1)
		logger.Debugf("serve request from %s to %s", ctx.Request.URL.Path, req.URL.Path)
		up.Proxy.ServeHTTP(ctx.Writer, req)
	}
}

// ValidateSCReq 校验来自sc的请求
func ValidateSCReq(c *gin.Context) {
	accessToken := c.GetHeader("Auth-Token")
	logger.Debug("areaToken in request Header: ", accessToken)
	ti, err := oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
	if err != nil {
		// 忽略掉areaToken 过期问题
		if err.Error() == jwt.ErrTokenIsExpired.Error() {
			c.Next()
			return
		}
		err = errors.New(status.Deny)
		response.HandleResponse(c, err, nil)
		c.Abort()
		return
	}

	// token 不过期时校验scope
	if ti.GetScope() == types.ScopeGetTokenBySC.Scope {
		c.Next()
		return
	}

	err = errors.New(status.Deny)
	response.HandleResponse(c, err, nil)
	c.Abort()
	return
}
