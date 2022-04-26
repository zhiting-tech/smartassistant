package extension

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/supervisor"
)

type extensionListResp struct {
	ExtensionNames []string `json:"extension_names"`
}

// ListExtension 处理扩展列表接口的请求
func ListExtension(c *gin.Context) {
	var (
		err  error
		resp extensionListResp
	)

	defer func() {
		if resp.ExtensionNames == nil {
			resp.ExtensionNames = make([]string, 0)
		}
		response.HandleResponse(c, err, resp)
	}()
	resp = GetExtensionsWithContext(c.Request.Context())
}

// GetExtensionsWithContext 获取扩展列表
func GetExtensionsWithContext(ctx context.Context) (exList extensionListResp) {
	resp, err := supervisor.GetClient().GetExtensionsWithContext(ctx)
	if err != nil {
		logrus.Warnf("getExtensions err is %s", err.Error())
		return
	}
	for _, e := range resp.Extensions {
		exList.ExtensionNames = append(exList.ExtensionNames, e.Name)
	}
	return
}

// HasExtensionWithContext 是否有该扩展
func HasExtensionWithContext(ctx context.Context, extensionName string) bool {
	resp, err := supervisor.GetClient().GetExtensionsWithContext(ctx)
	if err != nil {
		logrus.Warnf("HasExtension err is %s", err.Error())
		return false
	}
	for _, en := range resp.Extensions {
		if en.Name == extensionName {
			return true
		}
	}
	return false
}
