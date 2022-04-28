package trace

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync/atomic"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// Environment variable names
// go.opentelemetry.io/otel/exporters/jaeger/env.go
const (
	// Hostname for the Jaeger agent, part of address where exporter sends spans
	// i.e.	"localhost"
	envAgentHost = "OTEL_EXPORTER_JAEGER_AGENT_HOST"
	// Port for the Jaeger agent, part of address where exporter sends spans
	// i.e. 6831
	envAgentPort = "OTEL_EXPORTER_JAEGER_AGENT_PORT"
	// The HTTP endpoint for sending spans directly to a collector,
	// i.e. http://jaeger-collector:14268/api/traces.
	envEndpoint = "OTEL_EXPORTER_JAEGER_ENDPOINT"

	AgentAdminHTTP = 14271
)

var traceEnable int32 = 0

func Enable() bool {
	return atomic.LoadInt32(&traceEnable) != 0
}

func envOr(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return defaultValue
}

// Init 初始化链路跟踪
// 可通过环境变量配置，默认情况下会尝试常用情况
func Init(service string, opts ...sdktrace.TracerProviderOption) {
	go func() {
		var exp *jaeger.Exporter
		var err error
		_, ok := os.LookupEnv(envEndpoint)
		if ok {
			exp, err = jaeger.New(jaeger.WithCollectorEndpoint())
		}
		if !ok || err != nil {
			host := envOr(envAgentHost, "jaeger-agent")
			port := envOr(envAgentPort, "6831")
			// 如果用户指定地址合法，或者 docker-compose.yaml 中已配置 jaeger-agent，则直接使用
			_, err = net.ResolveUDPAddr("udp", net.JoinHostPort(host, port))
			if err != nil {
				// 测试 localhost agent 是否运行
				host = "localhost"
				err = checkAgentAdminStatus(host)
				if err != nil {
					return
				}
			}
			exp, err = jaeger.New(
				jaeger.WithAgentEndpoint(jaeger.WithAgentHost(host),
					jaeger.WithAgentPort(port)))
			if err != nil {
				return
			}
		}
		if exp == nil {
			return
		}
		options := []sdktrace.TracerProviderOption{
			// Always be sure to batch in production.
			sdktrace.WithBatcher(exp),
			// Record information about this application in a Resource.
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(service),
			)),
		}
		options = append(options, opts...)
		tp := sdktrace.NewTracerProvider(options...)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
		http.DefaultClient.Transport = otelhttp.NewTransport(http.DefaultTransport)
		atomic.StoreInt32(&traceEnable, 1)
	}()
}

func checkAgentAdminStatus(host string) error {
	resp, err := http.Get(fmt.Sprintf("http://%v:%v", host, AgentAdminHTTP))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("abnormal value of http status code: %v", resp.StatusCode)
	}
	return nil
}
