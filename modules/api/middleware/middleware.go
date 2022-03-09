// Package middleware GIN 框架中间件
package middleware

import (
	"context"
	errors2 "errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/utils/jwt"
	"gopkg.in/oauth2.v3"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/modules/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/reverseproxy"
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
func RequireAccountWithScope(scope string) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		// TODO 兼容代码，后续删除
		if c.GetHeader(types.ScopeTokenKey) != "" &&
			c.GetHeader(types.SATokenKey) == "" {
			c.Request.Header.Set(types.SATokenKey, c.GetHeader(types.ScopeTokenKey))
		}

		// 校验token
		ti, err := verifyAccessToken(c)
		if err != nil {
			response.HandleResponse(c, err, nil)
			c.Abort()
			return
		}

		// 没有设置scope，默认不限制
		if ti.GetScope() == "" {
			c.Next()
			return
		}
		// 校验scope
		if !strings.Contains(ti.GetScope(), scope) {
			err = errors2.New("permission deny: invalid scope")
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

		if err.Error() == errors.New(status.PasswordChanged).Error() {
			return ti, err
		}

		return ti, errors.Wrap(err, status.RequireLogin)
	}
	return
}

// RequireOwner 拥有者才能访问
func RequireOwner(c *gin.Context) {
	u := session.Get(c)
	if u == nil {
		response.HandleResponse(c, errors.New(status.RequireLogin), nil)
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
	response.HandleResponse(c, errors.New(status.RequireLogin), nil)
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
		// TODO 扩展不用SA反向代理后删掉
		if ctx.GetHeader(types.ScopeTokenKey) != "" &&
			ctx.GetHeader(types.SATokenKey) == "" {
			ctx.Request.Header.Set(types.SATokenKey, ctx.GetHeader(types.ScopeTokenKey))
		}

		req := ctx.Request.Clone(context.Background())
		user := session.Get(ctx)
		if user != nil {
			req.Header.Add("scope-user-id", strconv.Itoa(user.UserID))
		}

		// 替换插件静态文件地址（api/static/:sa_id/plugin/demo -> api/plugin/demo）
		oldPrefix := fmt.Sprintf("%s/plugin/%s", url.StaticPath(), path)
		newPrefix := fmt.Sprintf("api/plugin/%s", path)
		req.URL.Path = strings.Replace(req.URL.Path, oldPrefix, newPrefix, 1)
		logger.Debugf("serve request from %s to %s", ctx.Request.URL.Path, req.URL.Path)
		up.Proxy.ServeHTTP(ctx.Writer, req)
	}
}

// ValidateSCReq 校验来自sc的请求
func ValidateSCReq(c *gin.Context) {
	accessToken := c.GetHeader("Auth-Token")
	logrus.Debug("areaToken in request Header: ", accessToken)
	_, err := oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
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
}
