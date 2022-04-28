package oauth

import (
	"strconv"
	"time"

	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth/generate"
)

type TokenStore struct {
}

func NewTokenStore() *TokenStore {
	return &TokenStore{}
}

func (t TokenStore) Create(info oauth2.TokenInfo) error {
	return nil
}

func (t TokenStore) RemoveByCode(code string) error {
	return nil
}

func (t TokenStore) RemoveByAccess(access string) error {
	return nil
}

func (t TokenStore) RemoveByRefresh(refresh string) error {
	return nil
}

func (t TokenStore) GetByCode(code string) (oauth2.TokenInfo, error) {
	claims, err := generate.ParseCode(code)
	if err != nil {
		return nil, err
	}

	var ti models.Token
	ti.Code = code
	ti.UserID = strconv.Itoa(claims.UserID)
	ti.CodeExpiresIn = time.Duration(claims.ExpiresAt) * time.Second
	ti.CodeCreateAt = time.Unix(claims.CodeCreateAt, 0)
	ti.ClientID = claims.ClientID
	return &ti, nil
}

func (t TokenStore) GetByAccess(access string) (oauth2.TokenInfo, error) {

	claims, err := generate.ParseToken(access)
	if err != nil {
		return nil, err
	}

	var ti models.Token
	ti.Access = access
	ti.UserID = strconv.Itoa(claims.UserID)
	ti.AccessExpiresIn = time.Duration(claims.ExpiresAt) * time.Second
	ti.AccessCreateAt = time.Unix(claims.AccessCreateAt, 0)
	ti.ClientID = claims.ClientID
	ti.Scope = claims.Scope
	return &ti, nil
}

func (t TokenStore) GetByRefresh(refresh string) (oauth2.TokenInfo, error) {
	claims, err := generate.ParseToken(refresh)
	if err != nil {
		return nil, err
	}

	var ti models.Token
	ti.Refresh = refresh
	ti.UserID = strconv.Itoa(claims.UserID)
	ti.RefreshCreateAt = time.Unix(claims.RefreshCreateAt, 0)
	ti.RefreshExpiresIn = time.Duration(claims.ExpiresAt) * time.Second
	ti.ClientID = claims.ClientID
	ti.Scope = claims.Scope
	return &ti, nil
}
