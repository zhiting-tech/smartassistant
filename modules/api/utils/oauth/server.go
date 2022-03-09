package oauth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth/generate"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth/models"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"strings"
	"sync"
	"time"
)

var (
	serverOnce sync.Once

	oauthServer *server.Server
)

func GetOauthServer() *server.Server {
	serverOnce.Do(func() {
		oauthServer = newOauthServer()
	})
	return oauthServer
}

func newOauthServer() *server.Server {
	manager := manage.NewDefaultManager()

	// generate jwt access token
	manager.MapAccessGenerate(generate.NewJWTAccessGenerate(jwt.SigningMethodHS256))
	manager.MapAuthorizeGenerate(generate.NewAuthorizeGenerate())
	manager.MapClientStorage(models.NewClientStore())
	manager.MapTokenStorage(NewTokenStore())

	// refreshToken Config
	manager.SetRefreshTokenCfg(&manage.RefreshingConfig{
		AccessTokenExp:    24 * time.Hour,
		IsGenerateRefresh: true,
	})

	implicitTokenConfig := &manage.Config{
		AccessTokenExp: 24 * 365 * 10 * time.Hour,
	}
	manager.SetImplicitTokenCfg(implicitTokenConfig)

	config := &manage.Config{
		AccessTokenExp:    24 * time.Hour,
		RefreshTokenExp:   24 * time.Hour * 7,
		IsGenerateRefresh: true,
	}
	manager.SetPasswordTokenCfg(config)
	manager.SetAuthorizeCodeTokenCfg(config)

	clientTokenConfig := &manage.Config{
		AccessTokenExp: 24 * time.Hour,
	}
	manager.SetClientTokenCfg(clientTokenConfig)

	srv := server.NewDefaultServer(manager)
	srv.SetInternalErrorHandler(internalErrorHandler)
	srv.SetResponseErrorHandler(responseErrorHandler)
	srv.SetClientAuthorizedHandler(clientAuthorizedHandler)
	srv.SetClientScopeHandler(clientScopeHandler)
	return srv

}

func internalErrorHandler(err error) (re *errors.Response) {
	logger.Error("Internal Error:", err.Error())
	return
}

func responseErrorHandler(re *errors.Response) {
	logger.Error("Response Error:", re.Error.Error())
}

// clientAuthorizedHandler check the client allows to use this authorization grant type
func clientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
	client, err := entity.GetClientByClientID(clientID)
	if err != nil {
		return false, err
	}
	return strings.Contains(client.GrantType, string(grant)), nil
}

// clientScopeHandler check the client allows to use scope
func clientScopeHandler(clientID string, scope string) (allowed bool, err error) {
	client, _ := entity.GetClientByClientID(clientID)

	scopeList := strings.Split(scope, ",")
	if len(scopeList) == 0 {
		return false, nil
	}
	for _, scope := range scopeList {
		if _, ok := types.Scopes[scope]; !ok {
			return false, nil
		}
		if !strings.Contains(client.AllowScope, scope) {
			return false, nil
		}
	}
	return true, nil
}
