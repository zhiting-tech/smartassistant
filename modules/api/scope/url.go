// Package scope 用户 Scope Token
package scope

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/cache"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
)

// RegisterScopeRouter scope token 路由注册
func RegisterScopeRouter(r gin.IRouter) {
	scopeGroup := r.Group("scopes")
	scopeGroup.GET("", scopeList)
	scopeGroup.POST("token", requireCode, scopeToken)
}

func requireCode(c *gin.Context) {
	// 校验请求头verification-code是否有值
	code := c.GetHeader(types.VerificationKey)
	val, err := cache.Get(code)
	if err == nil {
		c.Request.Header.Set(types.SATokenKey, val.(string))
		c.Next()
		// 验证成功后删除code
		cache.Delete(code)
		return
	}

	// TODO code的处理
	u := session.Get(c)
	if u == nil {
		response.HandleResponse(c, errors.New(status.RequireLogin), nil)
		c.Abort()
		return
	}
}
