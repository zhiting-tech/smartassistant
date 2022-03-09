package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	orm2 "github.com/zhiting-tech/smartassistant/modules/entity"
)

var (
	ErrEmptySignedKey = errors.New("jwt signed key is empty")
	ErrTokenNotValid  = errors.New("jwt token is not valid")
	ErrTokenIsExpired = errors.New("jwt token is expired")
)

type AccessClaims struct {
	UID     int    `json:"uid"`
	AreaID  uint64 `json:"area_id,omitempty"`
	RoleIds []int  `json:"role_ids,omitempty"`
	SAID    string `json:"sa_id,omitempty"`
	Exp     int64  `json:"exp,omitempty"`
	Scope   string `json:"scope,omitempty"`
	AreaType orm2.AreaType `json:"area_type"`
	DepartmentIds []int `json:"department_ids"`
}

func (c AccessClaims) Valid() error {
	if time.Unix(c.Exp, 0).Before(time.Now()) {
		return ErrTokenIsExpired
	}
	return nil
}

// GenerateUserJwt 以用户 token 作为加密串生成 JWT，用户 ID 使用 uid 字段
func GenerateUserJwt(claims AccessClaims, userKey string, uID int) (jwtToken string, err error) {
	if len(userKey) == 0 {
		return "", ErrEmptySignedKey
	}

	if claims.UID == 0 {
		claims.UID = uID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(userKey))
}

// ValidateUserJwt 校验用户 JWT 的有效性
func ValidateUserJwt(jwtToken string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		claims, ok := token.Claims.(*AccessClaims)
		if !ok {
			return nil, ErrTokenNotValid
		}
		if claims.UID == 0 {
			return nil, ErrTokenNotValid
		}

		user, err := orm2.GetUserByID(claims.UID)
		if err != nil {
			return nil, err
		}
		return []byte(user.Key), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, ErrTokenNotValid
	}
}
