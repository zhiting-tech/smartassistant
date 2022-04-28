package httpclient

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	saTrace "github.com/zhiting-tech/smartassistant/pkg/trace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const TraceDomainFileEnv = ""

var traceDomains []string = []string{
	"*.zhitingtech.com",
}

var filterRegex []*regexp.Regexp = []*regexp.Regexp{}

func init() {
	var (
		traceDomainFile string
		exist           bool
		content         []byte
		err             error
	)
	if traceDomainFile, exist = os.LookupEnv(TraceDomainFileEnv); exist {
		if content, err = os.ReadFile(traceDomainFile); err != nil {
			return
		}

		lines := strings.Split(string(content), string(filepath.Separator))
		traceDomains = append(traceDomains, lines...)
	}

	for _, domain := range traceDomains {
		var reg *regexp.Regexp
		expr := strings.ReplaceAll(strings.ReplaceAll(domain, ".", "\\."), "*", ".*")
		if reg, err = regexp.Compile(expr); err != nil {
			continue
		}

		filterRegex = append(filterRegex, reg)
	}
}

func newTraceTransport(tr http.RoundTripper, opts ...otelhttp.Option) *traceTransport {
	return &traceTransport{
		tr:   tr,
		opts: opts,
	}
}

type traceTransport struct {
	tr   http.RoundTripper
	opts []otelhttp.Option
}

func (t *traceTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	tr := t.tr
	if saTrace.Enable() {
		// 防止重复嵌套
		if _, ok := t.tr.(*otelhttp.Transport); !ok {
			tr = otelhttp.NewTransport(t.tr, t.opts...)
		}

	}

	return tr.RoundTrip(r)
}

type TraceOption struct {
	c *http.Client
	traceOption []otelhttp.Option
}

var DefaultClient = NewHttpClient()

type HttpClientOption interface {
	apply(t *TraceOption)
}

type HttpClientOptionFn func(t *TraceOption)

func (fn HttpClientOptionFn) apply(t *TraceOption) {
	fn(t)
}

type CheckRedirectFn func(req *http.Request, via []*http.Request) error

func WithTransport(rt http.RoundTripper) HttpClientOption {
	return HttpClientOptionFn(
		func(t *TraceOption) {
			t.c.Transport = rt
		},
	)
}

func WithTimeout(timeout time.Duration) HttpClientOption {
	return HttpClientOptionFn(
		func(t *TraceOption) {
			t.c.Timeout = timeout
		},
	)
}

func WithCheckRedirect(checkRedirect CheckRedirectFn) HttpClientOption {
	return HttpClientOptionFn(
		func(t *TraceOption) {
			t.c.CheckRedirect = checkRedirect
		},
	)
}

func WithCookieJar(jar http.CookieJar) HttpClientOption {
	return HttpClientOptionFn(
		func(t *TraceOption) {
			t.c.Jar = jar
		},
	)
}

func WithTraceTransport(opt ...otelhttp.Option) HttpClientOption {
	return HttpClientOptionFn(
		func(t *TraceOption) {
			t.traceOption = opt
		},
	)
}

func NewHttpClient(options ...HttpClientOption) *http.Client {
	t := &TraceOption{
		c: &http.Client{
			Transport: http.DefaultTransport,
		},
	}

	for _, opt := range options {
		opt.apply(t)
	}

	// 防止重复嵌套
	t.traceOption = append(t.traceOption, GetTraceFilterOption())
	t.c.Transport = newTraceTransport(t.c.Transport,t.traceOption...)
	return t.c
}

func GetTraceFilterOption() otelhttp.Option {
	return otelhttp.WithFilter(traceFilter)
}

func GetTraceSpanOption(ctx context.Context) otelhttp.Option {
	return otelhttp.WithSpanOptions(
		trace.WithLinks(
			trace.LinkFromContext(
				ctx,
				attribute.String(
					"opentracing.ref_type",
					"follows_from",
				),
			),
		),
	)
}

func traceFilter(req *http.Request) bool {

	hostname := req.URL.Hostname()
	if ip := net.ParseIP(hostname); ip != nil {
		return filterIP(ip)
	}

	return filterDomain(hostname)
}

func filterIP(ip net.IP) bool {
	// 过滤通过ip访问的地址
	return true
}

func filterDomain(domain string) bool {
	for _, reg := range filterRegex {
		if reg.Match([]byte(domain)) {
			return true
		}
	}
	return false
}
