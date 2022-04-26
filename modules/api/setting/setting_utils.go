package setting

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	errors2 "errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/ssh"
	"gopkg.in/oauth2.v3"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/oauth"

	"github.com/zhiting-tech/smartassistant/modules/api/utils/cloud"
	"github.com/zhiting-tech/smartassistant/modules/entity"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/http/httpclient"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

const (
	HttpRequestTimeout = (time.Duration(30) * time.Second)
	PublicKeyEmail     = "smartassistant@zhitingtech.com"
)

func GetAreaAuthToken(areaID uint64) string {
	req, _ := http.NewRequest("", "", nil)
	scClient, _ := entity.GetSCClient(areaID)
	tgr := oauth2.TokenGenerateRequest{
		ClientID:       scClient.ClientID,
		ClientSecret:   scClient.ClientSecret,
		Scope:          scClient.AllowScope,
		Request:        req,
		AccessTokenExp: 30 * 24 * time.Hour, // 设置为30天有效期
	}

	ti, err := oauth.GetOauthServer().GetAccessToken(oauth2.ClientCredentials, &tgr)
	if err != nil {
		logger.Errorf("get access token failed: (%v)", err)
		return ""
	}

	return ti.GetAccess()
}

// sendAreaAuthToSC 发送认证token给SC
func sendAreaAuthToSC(areaID uint64) {
	if len(config.GetConf().SmartCloud.Domain) <= 0 {
		return
	}
	var err error
	defer func() {
		updates := map[string]interface{}{
			"is_send_auth_to_sc": err == nil,
		}
		if err = entity.UpdateArea(areaID, updates); err != nil {
			logger.Error(err)
		}
	}()
	saID := config.GetConf().SmartAssistant.ID
	scUrl := config.GetConf().SmartCloud.URL()
	url := fmt.Sprintf("%s/sa/%s/areas/%d", scUrl, saID, areaID)
	body := map[string]interface{}{
		"area_token": GetAreaAuthToken(areaID),
	}
	b, _ := json.Marshal(body)
	logger.Debug(url)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(b))
	if err != nil {
		logger.Warnf("NewRequest error %v\n", err)
		return
	}

	req.Header = cloud.GetCloudReqHeader()
	ctx, _ := context.WithTimeout(context.Background(), HttpRequestTimeout)
	req.WithContext(ctx)
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Warnf("request %s error %v\n", url, err)
		return
	}
	if httpResp.StatusCode != http.StatusOK {
		logger.Warnf("request %s error,status:%v\n", url, httpResp.Status)
		err = errors2.New(httpResp.Status)
		return
	}
}

func SendAreaAuthToSC() {
	areas, err := entity.GetAreas()
	if err != nil {
		logger.Errorf("get areas err (%v)", err)
		return
	}

	for _, area := range areas {
		if !area.IsSendAuthToSC {
			sendAreaAuthToSC(area.ID)
		}
	}

}

func SendPrivateKeyToSCWithContext(ctx context.Context, privateKey []byte) (err error) {
	var (
		req  *http.Request
		resp *http.Response
	)
	saID := config.GetConf().SmartAssistant.ID
	scUrl := config.GetConf().SmartCloud.URL()
	url := fmt.Sprintf("%s/sa/%s/private_key", scUrl, saID)
	body := map[string]interface{}{
		"private_key": string(privateKey),
	}
	b, _ := json.Marshal(body)

	ctx, _ = context.WithTimeout(ctx, HttpRequestTimeout)
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b)); err != nil {
		return
	}
	req.Header = cloud.GetCloudReqHeader()

	if resp, err = httpclient.DefaultClient.Do(req); err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		logger.Warnf("request %s error,status:%v\n", url, resp.Status)
		err = errors2.New(resp.Status)
	}
	return
}

func sshKeyGen() (privateKey, publicKey []byte, err error) {
	var (
		rsaPrivateKey *rsa.PrivateKey
		sshPublicKey  ssh.PublicKey
	)

	// 生成rsa公私钥
	if rsaPrivateKey, err = rsa.GenerateKey(rand.Reader, 2048); err != nil {
		return
	}
	// 生成ssh私钥
	privateKey = pem.EncodeToMemory(&pem.Block{
		Bytes: x509.MarshalPKCS1PrivateKey(rsaPrivateKey),
		Type:  "RSA PRIVATE KEY",
	})

	if sshPublicKey, err = ssh.NewPublicKey(&rsaPrivateKey.PublicKey); err != nil {
		return
	}

	// 生成ssh公钥内容

	publicKey = MarshalAuthorizedKey(sshPublicKey)

	return
}

func MarshalAuthorizedKey(key ssh.PublicKey) []byte {
	b := &bytes.Buffer{}
	b.WriteString(key.Type())
	b.WriteByte(' ')
	e := base64.NewEncoder(base64.StdEncoding, b)
	e.Write(key.Marshal())
	e.Close()
	b.WriteByte(' ')
	b.WriteString(PublicKeyEmail)
	b.WriteByte('\n')

	return b.Bytes()
}
