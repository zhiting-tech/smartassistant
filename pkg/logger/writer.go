package logger

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	StatusStart = iota + 1
	StatusConnected
	StatusDisconnect
)

var (
	writer *remoteWriter
	once   sync.Once
)

func getRemoteWriter() *remoteWriter {
	// 防止创建多个writer
	once.Do(func() {
		writer = newRemoteWriter()
		go writer.run(context.Background())
	})

	return writer
}

type remoteWriter struct {
	ch     chan []byte
	status int32
}

func newRemoteWriter() *remoteWriter {
	return &remoteWriter{
		ch:     make(chan []byte, 1024),
		status: StatusStart,
	}
}

func (w *remoteWriter) Write(data []byte) (n int, err error) {
	// 不返回错误
	n = len(data)

	// 传入的data是pool里面的对象，异步处理需要复制一份，否则数据会错乱
	// TODO: 后续使用pool，降低GC
	log := make([]byte, len(data))
	copy(log, data)

	// 连接失败时不发送日志
	if atomic.LoadInt32(&w.status) == StatusDisconnect {
		return
	}

	select {
	case w.ch <- log:
	default:
	}

	return
}

func (w *remoteWriter) run(ctx context.Context) {
	var (
		err error
	)

	for {
		var reDial chan struct{} = make(chan struct{})

		ws := w.untilConnectToRemoteWithContext(ctx)
		if ws == nil {
			return
		}

		clear := func(ws *websocket.Conn) {
			ws.Close()
			atomic.StoreInt32(&w.status, StatusDisconnect)
		}
		go func(ws *websocket.Conn, reDial chan struct{}) {

			for {
				if t, _, err := ws.ReadMessage(); err != nil || t == websocket.CloseMessage {
					close(reDial)
					return
				}
			}
		}(ws, reDial)

	LogLoop:
		for {
			select {
			case <-ctx.Done():
				clear(ws)
				return
			case <-reDial:
				clear(ws)
				break LogLoop
			case data := <-w.ch:
				if err = ws.WriteMessage(websocket.BinaryMessage, data); err != nil {
					clear(ws)
					break LogLoop
				}
			}
		}
	}
}

func (w *remoteWriter) untilConnectToRemoteWithContext(ctx context.Context) (ws *websocket.Conn) {
	var (
		hosts []string = []string{
			// "zt-nginx:9020",
			// "127.0.0.1:9020",
			"127.0.0.1:37966",
			"smartassistant:37966",
		}
		wait = 10 * time.Second
		err  error
	)

	for {
		for _, host := range hosts {
			// 第一次连接时延迟连接，主要处理sa自身端口未启动的情况
			time.Sleep(wait)

			if ws, _, err = websocket.DefaultDialer.DialContext(ctx, fmt.Sprintf("ws://%s/log", host), nil); err == nil {
				atomic.StoreInt32(&w.status, StatusConnected)
				return
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}

		atomic.StoreInt32(&w.status, StatusDisconnect)
	}

}
