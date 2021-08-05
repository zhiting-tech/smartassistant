package user

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/internal/api/utils/response"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/types/status"
	"github.com/zhiting-tech/smartassistant/internal/utils/cache"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/rand"
	"time"
)

const (
	codeLength   = 6                // 验证码的长度
	codeExpireIn = time.Minute * 10 // 验证码有效期
)

type GetVerificationCodeResp struct {
	Code     string `json:"code"`
	ExpireIn int    `json:"expire_in"`
}

func GetVerificationCode(c *gin.Context) {

	var (
		err  error
		resp GetVerificationCodeResp
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	user := session.Get(c)
	if user == nil {
		err = errors.New(status.RequireLogin)
		return
	}

	// 校验用户是否是拥有者
	if !entity.IsSAOwner(user.UserID) {
		err = errors.New(status.Deny)
		return
	}

	code := rand.StringK(codeLength, rand.KindAll)

	cache.GetCache().Set(code, user.Token, codeExpireIn)

	resp.Code = code
	resp.ExpireIn = int(codeExpireIn / time.Second)
}