package sdk

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto/v2"
	addr2 "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/utils/addr"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/utils/registry"
)

const Version = "v2.0.0"

type DiscoverFunc func(ctx context.Context, devices chan<- Device)

type MetaData struct {
	SDKVersion string `json:"sdk_version"`
}

func Run(p *Server) error {

	ctx, cancel := context.WithCancel(context.Background())

	go p.discovering()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		select {
		case <-sig:
			cancel()
		}
	}()

	if err := runServer(ctx, p); err != nil {
		logrus.Error(err)
	}
	logrus.Info("shutting down.")
	return nil
}

// mixHandler 同时处理http和grpc请求
func mixHandler(engine *gin.Engine, grpcServer *grpc.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 {
			engine.ServeHTTP(w, r)
			return
		}
		if strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
			return
		}
		return
	}
}

func runServer(ctx context.Context, p *Server) error {

	ln, err := getRangeListener()
	if err != nil {
		return err
	}

	var localIP string
	// localIP 先读环境变量
	localIP = os.Getenv("LOCAL_IP")
	if localIP == "" {
		localIP, err = addr2.LocalIP()
		if err != nil {
			return err
		}
	}

	addr := net.TCPAddr{
		IP:   net.ParseIP(localIP),
		Port: ln.Addr().(*net.TCPAddr).Port,
	}
	// 往etcd注册服务
	target := registry.EndpointsKey(p.Domain)
	endpoint := endpoints.Endpoint{Addr: addr.String(), Metadata: MetaData{SDKVersion: Version}}
	go registry.RegisterService(ctx, target, endpoint)

	// grpc服务
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.ChainUnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	proto.RegisterPluginServer(grpcServer, p)

	// h2c实现了不用tls的http/2
	h1s := http.Server{
		Handler: h2c.NewHandler(mixHandler(p.Router, grpcServer), &http2.Server{}),
	}

	go func() {
		if err = h1s.Serve(ln); err != nil {
			logrus.Errorf("h1s serve err: %+v", err)
		}
	}()
	logrus.Infoln("server started")
	<-ctx.Done()
	logrus.Infoln("server stopped")
	shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	registry.UnregisterService(shutdownContext, target)
	if err = h1s.Shutdown(shutdownContext); err != nil {
		logrus.Errorf("h1s shutdown err: %+v", err)
	}
	return err
}

func getRangeListener() (listener net.Listener, err error) {
	min := 10000
	max := 50000
	for port := min; port < max; port++ {
		listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err == nil {
			return
		}
	}
	listener, err = net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return
	}
	return
}
