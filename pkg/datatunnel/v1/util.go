package datatunnel

import (
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	DefaultTCPKeepAlivePeriod = (5 * time.Minute)
	DefaultLingerSecond       = 30
)

func getLogrusLevel(logLevel string) logrus.Level {
	switch logLevel {
	case "trace":
		return logrus.TraceLevel
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "panic":
		return logrus.PanicLevel
	default:
		return logrus.WarnLevel
	}
}

func Proxy(from, to net.Conn) {
	fn := func(from, to net.Conn) {
		defer from.Close()
		defer to.Close()

		io.Copy(from, to)
	}

	go fn(from, to)
	go fn(to, from)
}
