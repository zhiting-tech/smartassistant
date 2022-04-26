package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/types"
	"github.com/zhiting-tech/smartassistant/pkg/http/httpclient"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"io"
	"io/ioutil"
	"net/http"
)

type SCResp struct {
	Status int                    `json:"status"`
	Reason string                 `json:"reason"`
	Data   map[string]interface{} `json:"data"`
}

type RequestOptions struct {
	isDefaultClient bool
}

type SCRequestOption interface {
	apply(options *RequestOptions)
}

type clientOption bool

func (c clientOption) apply(opts *RequestOptions) {
	opts.isDefaultClient = bool(c)
}

func WithDefaultClient(c bool) SCRequestOption {
	return clientOption(c)
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

// getResponse 通过不同client获取sc返回数据
func getResponse(ctx context.Context, url, method string, content []byte, isDefaultClient bool) (response *http.Response, err error) {
	if isDefaultClient {
		var req *http.Request
		req, err = NewRequestWithContext(ctx, method, url, bytes.NewBuffer(content))
		if err != nil {
			logger.Error("new request error:", err.Error())
			return
		}
		response, err = http.DefaultClient.Do(req)
		if err != nil {
			logger.Error("do request error:", err.Error())
			return
		}
		return
	}
	// 这里使用的clint是用于处理父span不依赖子span的follows_from模式
	request, err := http.NewRequest(method, url, bytes.NewReader(content))
	if err != nil {
		logger.Error("new request error:", err.Error())
		return
	}
	request.Header = GetCloudReqHeader()
	response, err = httpclient.NewHttpClient(httpclient.WithTraceTransport(httpclient.GetTraceSpanOption(ctx))).Do(request)
	if err != nil {
		logger.Error("do request error:", err.Error())
		return
	}
	return
}

// DoWithContext 请求云端接口，并附加链路信息
func DoWithContext(ctx context.Context, path, method string, requestData map[string]interface{}, opts ...SCRequestOption) (resp []byte, err error) {
	options := RequestOptions{
		isDefaultClient: true,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	url := fmt.Sprintf("%s/%s/sa/%s/%s",
		config.GetConf().SmartCloud.URL(), types.SCVersion, config.GetConf().SmartAssistant.ID, path)
	content, _ := json.Marshal(&requestData)
	response, err := getResponse(ctx, url, method, content, options.isDefaultClient)
	if err != nil {
		logger.Error("get response error:", err.Error())
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
