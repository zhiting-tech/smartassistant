package generate

import (
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/oauth2.v3"

	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
)

// NewAuthorizeGenerate create to generate the authorize code instance
func NewAuthorizeGenerate() *AuthorizeGenerate {
	return &AuthorizeGenerate{}
}

// AuthorizeGenerate generate the authorize code
type AuthorizeGenerate struct{}

// Token based on the UUID generated token
func (ag *AuthorizeGenerate) Token(data *oauth2.GenerateBasic) (string, error) {
	grantType := data.Request.Header.Get(types.GrantType)
	userID, err := strconv.Atoi(data.UserID)
	if err != nil {
		return "", err
	}
	user, err := entity.GetUserByID(userID)
	if err != nil {
		return "", err
	}

	claims := &JWTAccessClaims{
		UserID:       userID,
		AreaID:       user.AreaID,
		ClientID:     data.TokenInfo.GetClientID(),
		Scope:        data.TokenInfo.GetScope(),
		GrantType:    grantType,
		CodeCreateAt: data.TokenInfo.GetCodeCreateAt().Unix(),
		ExpiresAt:    int64(data.TokenInfo.GetCodeExpiresIn().Seconds()),
	}
	code, err := getToken(claims, user.Key, jwt.SigningMethodHS256)
	if err != nil {
		return "", err
	}
	return code, nil
}
