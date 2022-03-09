package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type SCResp struct {
	Status int                    `json:"status"`
	Reason string                 `json:"reason"`
	Data   map[string]interface{} `json:"data"`
}

// CloudRequest 请求Cloud SC的方法
// Deprecated: 请使用 DoWithContext 或 cloud.NewRequestWithContext
func CloudRequest(url, method string, requestData map[string]interface{}) (resp []byte, err error) {
	content, _ := json.Marshal(&requestData)
	req, err := NewRequestWithContext(context.Background(), method, url, bytes.NewBuffer(content))
	if err != nil {
		logger.Error("new request error:", err.Error())
		return
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("do request error:", err.Error())
		return
	}

	if response.StatusCode != http.StatusOK {
		logger.Errorf("http status: %s", response.Status)
		return resp, http.ErrNotSupported
	}

	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	var scResp SCResp
	if err = json.Unmarshal(resp, &scResp); err != nil {
		return
	}

	// 0 标识请求成功
	if scResp.Status != 0 {
		logger.Errorf("status %d,reason %s", scResp.Status, scResp.Reason)
		err = errors.New(scResp.Reason)
		return
	}
	return
}

// GetCloudReqHeader 获取请求Cloud SC的Header
// Deprecated: 请使用 cloud.NewRequestWithContext
func GetCloudReqHeader() http.Header {
	saID := config.GetConf().SmartAssistant.ID
	saKey := config.GetConf().SmartAssistant.Key
	header := http.Header{}
	header.Set(types.SAID, saID)
	header.Set(types.SAKey, saKey)
	return header
}

// NewRequestWithContext 增加SmartCloud相关请求头信息
// 请传递链路 context
func NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	header.Set(types.SAID, config.GetConf().SmartAssistant.ID)
	header.Set(types.SAKey, config.GetConf().SmartAssistant.Key)
	req.Header = header
	return req, nil
}

// DoWithContext 请求云端接口，并附加链路信息
func DoWithContext(ctx context.Context, url, method string, requestData map[string]interface{}) (resp []byte, err error) {
	content, _ := json.Marshal(&requestData)
	req, err := NewRequestWithContext(ctx, method, url, bytes.NewBuffer(content))
	if err != nil {
		logger.Error("new request error:", err.Error())
		return
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("do request error:", err.Error())
		return
	}

	if response.StatusCode != http.StatusOK {
		logger.Errorf("http status: %s", response.Status)
		return resp, http.ErrNotSupported
	}

	defer response.Body.Close()
	resp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	var scResp SCResp
	if err = json.Unmarshal(resp, &scResp); err != nil {
		return
	}

	// 0 标识请求成功
	if scResp.Status != 0 {
		logger.Errorf("status %d,reason %s", scResp.Status, scResp.Reason)
		err = errors.New(scResp.Reason)
		return
	}
	return
}
