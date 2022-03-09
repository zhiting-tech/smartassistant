package generate

import (
	"encoding/json"
	errors2 "errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"gopkg.in/oauth2.v3"
	"strconv"
	"strings"
	"time"
)

var (
	ErrTokenNotValid = errors2.New("jwt token is not valid")
	ErrExpiredToken  = errors2.New("jwt token is expired")
)

// JWTAccessClaims jwt claims
type JWTAccessClaims struct {
	UserID          int    `json:"user_id,omitempty"`
	ExpiresAt       int64  `json:"exp,omitempty"`
	AreaID          uint64 `json:"area_id,omitempty"`
	AccessCreateAt  int64  `json:"access_create_at,omitempty"`
	RefreshCreateAt int64  `json:"refresh_create_at,omitempty"`
	ClientID        string `json:"client_id,omitempty"`
	Scope           string `json:"scope,omitempty"`
	CodeCreateAt    int64  `json:"code_create_at,omitempty"`
	GrantType		string `json:"grant_type,omitempty"`
}

// Valid claims verification
func (a *JWTAccessClaims) Valid() error {
	createAt := a.AccessCreateAt
	// 获取token的创建时间
	if createAt == 0 {
		createAt = a.CodeCreateAt
		if createAt == 0 {
			createAt = a.RefreshCreateAt
		}
	}
	if time.Unix(createAt, 0).Add(time.Duration(a.ExpiresAt) * time.Second).Before(time.Now()) {
		return ErrExpiredToken
	}
	return nil
}

// NewJWTAccessGenerate create to generate the jwt access token instance
func NewJWTAccessGenerate(method jwt.SigningMethod) *JWTAccessGenerate {
	return &JWTAccessGenerate{
		SignedMethod: method,
	}
}

// JWTAccessGenerate generate the jwt access token
type JWTAccessGenerate struct {
	SignedMethod jwt.SigningMethod
}

// Token based on the UUID generated token
func (a *JWTAccessGenerate) Token(data *oauth2.GenerateBasic, isGenRefresh bool) (string, string, error) {
	areaID, _ := strconv.ParseUint(data.Request.Header.Get(types.AreaID), 10, 64)
	grantType := data.Request.Header.Get(types.GrantType)
	var userID int
	if data.UserID != "" {
		var uerr error
		userID, uerr = strconv.Atoi(data.UserID)
		if uerr != nil {
			return "", "", uerr
		}
	}

	claims := &JWTAccessClaims{
		UserID:   userID,
		AreaID:   areaID,
		ClientID: data.TokenInfo.GetClientID(),
		Scope:    data.TokenInfo.GetScope(),
		GrantType: grantType,
	}

	userKey := data.Request.Header.Get(types.UserKey)
	// 授权码模式获取code
	if data.TokenInfo.GetCodeExpiresIn() != 0 {
		claims.CodeCreateAt = data.TokenInfo.GetCodeCreateAt().Unix()
		claims.ExpiresAt = int64(data.TokenInfo.GetCodeExpiresIn().Seconds())
		code, err := a.GetToken(claims, userKey)
		if err != nil {
			return "", "", err
		}
		return code, "", nil
	}

	claims.ExpiresAt = int64(data.TokenInfo.GetAccessExpiresIn().Seconds())
	claims.AccessCreateAt = data.TokenInfo.GetAccessCreateAt().Unix()

	var key = userKey
	if userKey == "" {
		// 客户端模式授权
		key = data.Client.GetSecret()
	}

	if key == "" {
		return "", "", jwt.ErrInvalidKey
	}

	access, err := a.GetToken(claims, key)
	if err != nil {
		return "", "", err
	}

	refresh := ""
	if isGenRefresh {
		refreshClaims := &JWTAccessClaims{
			UserID:          userID,
			ExpiresAt:       int64(data.TokenInfo.GetRefreshExpiresIn().Seconds()),
			AreaID:          areaID,
			RefreshCreateAt: data.TokenInfo.GetRefreshCreateAt().Unix(),
			ClientID:        data.TokenInfo.GetClientID(),
			Scope:           data.TokenInfo.GetScope(),
		}

		refresh, err = a.GetToken(refreshClaims, userKey)
		if err != nil {
			return "", "", err
		}
	}

	return access, refresh, nil
}

func ParseJwt(access string) (*JWTAccessClaims, error) {
	token, err := jwt.ParseWithClaims(access, &JWTAccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("signing method is invalid,method: %v", token.Header["alg"])
		}

		claims, err := DecodeJwt(access)
		if err != nil {
			return nil, err
		}
		var key string
		userID := claims.UserID
		if userID != 0 {
			user, err := entity.GetUserByIDAndAreaID(claims.UserID, claims.AreaID)
			if err != nil {
				return nil, err
			}
			key = user.Key
			switch oauth2.GrantType(claims.GrantType) {
			case oauth2.PasswordCredentials:
				if claims.AccessCreateAt < user.PasswordUpdateTime.Unix() {
					return nil, errors.New(status.PasswordChanged)
				}
			case oauth2.Refreshing:
				if claims.RefreshCreateAt < user.PasswordUpdateTime.Unix() {
					return nil, errors.New(status.PasswordChanged)
				}
			}
		} else {
			// 是否客户端模式授权,client key 加密的
			clientID := claims.ClientID
			client, err := entity.GetClientByClientID(clientID)
			if err != nil {
				return nil, err
			}
			key = client.ClientSecret
		}
		return []byte(key), nil
	})

	if err != nil {
		return nil, err
	}

	var claims *JWTAccessClaims
	var ok bool
	if claims, ok = token.Claims.(*JWTAccessClaims); !ok {
		return nil, ErrTokenNotValid
	}

	if err = claims.Valid(); err != nil {
		return nil, err
	}
	return claims, nil
}

func DecodeJwt(access string) (*JWTAccessClaims, error) {
	strSlice := strings.Split(access, ".")
	if len(strSlice) < 2 {
		return nil, ErrTokenNotValid
	}
	bytes, err := jwt.DecodeSegment(strSlice[1])
	if err != nil {
		return nil, err
	}

	var claims JWTAccessClaims
	json.Unmarshal(bytes, &claims)
	return &claims, nil
}

func (a *JWTAccessGenerate) GetToken(claims *JWTAccessClaims, key string) (string, error) {
	token := jwt.NewWithClaims(a.SignedMethod, claims)
	str, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}
	return str, nil
}
