package url

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/zhiting-tech/smartassistant/modules/config"
)

// BuildQuery 将map中数据转换成 url.Values
func BuildQuery(query map[string]interface{}) url.Values {
	values := url.Values{}
	for k, v := range query {
		values.Add(k, fmt.Sprintf("%v", v))
	}
	return values
}

// Join 将v中的所有键值对拼接成URL
func Join(v url.Values) string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		keyEscaped := k
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}
	return buf.String()
}

// BuildURL 根据请求和相对路径转换成地址
func BuildURL(path string, query map[string]interface{}, req *http.Request) *url.URL {
	values := BuildQuery(query)
	u := url.URL{
		Host:     req.Host,
		Path:     path,
		Scheme:   req.URL.Scheme,
		RawQuery: Join(values), // Go的URL编码默认是将空格编码为"+"号，前端难以确定参数值中的+号原先是不是空格，所以后端只拼接字符串，不再进行编码
	}
	if u.Scheme == "" { // 使用nginx代理则需要配置才能获取scheme
		u.Scheme = req.Header.Get("X-Scheme")
		if u.Scheme == "" {
			u.Scheme = "http"
		}
	}
	return &u
}

// ConcatPath 拼接路径
func ConcatPath(paths ...string) string {
	return strings.Join(paths, "/")
}

// BackendStaticPath 后端静态文件路径
func BackendStaticPath() string {
	return ConcatPath("backend", "static",
		config.GetConf().SmartAssistant.ID, "sa")
}

// ImagePath 图片相对路径
func ImagePath(imgName string) string {
	return ConcatPath(BackendStaticPath(), "img", imgName)
}

// ImageUrl 静态文件夹图片地址
func ImageUrl(req *http.Request, imgName string) string {
	return BuildURL(ImagePath(imgName), nil, req).String()
}

// SAImageUrl SA的Logo地址
func SAImageUrl(req *http.Request) string {
	return ImageUrl(req, "智慧中心.png")
}

func FilePath() string {
	return ConcatPath("file", config.GetConf().SmartAssistant.ID, "sa")
}

// FileUrl 用户上传的文件路径
func FileUrl(req *http.Request, fileName string) string {
	path := ConcatPath(FilePath(), fileName)
	return BuildURL(path, nil, req).String()
}
