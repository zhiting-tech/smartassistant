package session

import (
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"strconv"
	"time"
)

const sessionName = "user"

type User struct {
	UserID        int                    `json:"uid"`
	IsOwner       bool                   `json:"is_owner"`
	UserName      string                 `json:"user_name"`
	RoleID        int                    `json:"role_id"`
	Token         string                 `json:"token"`
	LoginAt       time.Time              `json:"login_at"`
	LoginDuration time.Duration          `json:"login_duration"`
	ExpiresAt     time.Time              `json:"expires_at"`
	AreaID        uint64                 `json:"area_id"`
	Option        map[string]interface{} `json:"option"`
	Key           string
}

func (u User) BelongsToArea(areaID uint64) bool {
	return u.AreaID == areaID
}

// Get 根据token或cookie获取用户数据
func Get(c *gin.Context) *User {
	if u, exists := c.Get("userInfo"); exists {
		return u.(*User)
	}
	var u *User
	token := c.GetHeader(types.SATokenKey)
	if token == "" {
		return nil
	}
	u = GetUserByToken(c)
	c.Set("userInfo", u)
	return u
}

func GetUserByToken(c *gin.Context) *User {
	accessToken := c.GetHeader(types.SATokenKey)
	ti, err := oauth.GetOauthServer().Manager.LoadAccessToken(accessToken)
	if err != nil {
		return nil
	}

	uid, _ := strconv.Atoi(ti.GetUserID())
	user, _ := entity.GetUserByID(uid)

	area, err := entity.GetAreaByID(user.AreaID)
	if err != nil {
		return nil
	}

	u := &User{
		UserID:   uid,
		UserName: user.AccountName,
		Token:    accessToken,
		AreaID:   user.AreaID,
		IsOwner:  area.OwnerID == user.ID,
		Key:      user.Key,
	}
	return u
}
