package middleware

import (
	"bytes"
	"io/ioutil"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type AccessLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w AccessLogWriter) Write(p []byte) (int, error) {
	if n, err := w.body.Write(p); err != nil {
		return n, err
	}
	return w.ResponseWriter.Write(p)
}

// AccessLog 记录请求日志
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyWriter := &AccessLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyWriter

		spanContext := trace.SpanContextFromContext(c.Request.Context())
		if spanContext.IsValid() {
			c.Writer.Header().Add("Zt-Trace-Id", spanContext.TraceID().String())
		}
		var request string
		if c.Request.Body != nil {
			// 使用ReadAll从body中读取数据，会把Body清空
			data, _ := ioutil.ReadAll(c.Request.Body)
			request = string(data)
			// 重新写入到Body
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		}

		// 格式化时间输出
		beginTime := time.Now().Format("2006/01/02 15:04:05")
		c.Next()
		endTime := time.Now().Format("2006/01/02 15:04:05")

		var response string
		response = bodyWriter.body.String()

		// 不需要记录下面两种情况：
		// 访问静态资源、
		// 非json请求和响应（屏蔽上传或者下载）
		if strings.HasPrefix(c.Request.RequestURI, "/api/static") ||
			c.Writer.Header().Get("Content-Type") != "application/json; charset=utf-8" ||
			strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data") {

			request = ""
			response = ""
		}

		fields := logrus.Fields{
			"client_ip":   c.ClientIP(),
			"module":      "api",
			"request_uri": c.Request.RequestURI,
			"request":     request,
			"response":    response,
		}
		if spanContext.IsValid() {
			fields["trace_id"] = spanContext.TraceID().String()
		}
		l := logger.NewEntry()
		l.WithFields(fields).Debugf(
			"access log: method %s, status_code:% d, begin_time: %s, end_time: %s",
			c.Request.Method,
			bodyWriter.Status(),
			beginTime,
			endTime)
	}
}
