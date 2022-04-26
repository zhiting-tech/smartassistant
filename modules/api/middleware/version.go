package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-version"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"net/http"
	"sync/atomic"
)

type apiVersionKey struct{}

var globalDefaultVersion atomic.Value

func VersionMiddleware(paramKey, defaultVersion, minVersion string) gin.HandlerFunc {
	dVersion, err := version.NewSemver(defaultVersion)
	if err != nil {
		logger.Errorf("wrong defaultVersion format")
		return nil
	}
	globalDefaultVersion.Store(dVersion)
	mVersion, err := version.NewSemver(minVersion)
	if err != nil {
		logger.Errorf("wrong minVersion format")
		return nil
	}
	return func(c *gin.Context) {
		savedCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()
		var apiVersion *version.Version
		versionStr := c.Param(paramKey)
		if versionStr == "" {
			apiVersion = dVersion
		} else {
			apiVersion, err = version.NewSemver(versionStr)
			if err != nil {
				response.HandleResponseWithStatus(c, http.StatusOK, errors.New(errors.APIVersion), nil)
				c.Abort()
				return
			}
		}
		if apiVersion.LessThan(mVersion) {
			response.HandleResponseWithStatus(c, http.StatusOK, errors.New(errors.APIVersion), nil)
			c.Abort()
			return
		}
		if apiVersion.GreaterThan(dVersion) {
			response.HandleResponseWithStatus(c, http.StatusOK, errors.New(errors.APIVersion), nil)
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(context.WithValue(savedCtx, apiVersionKey{}, apiVersion))
		c.Next()
	}
}

func VersionFromContext(ctx context.Context) *version.Version {
	if ctx == nil {
		return globalDefaultVersion.Load().(*version.Version)
	}
	if val := ctx.Value(apiVersionKey{}); val != nil {
		return val.(*version.Version)
	}
	return globalDefaultVersion.Load().(*version.Version)
}

func init() {
	d, _ := version.NewSemver("1.0.0")
	globalDefaultVersion.Store(d)
}
