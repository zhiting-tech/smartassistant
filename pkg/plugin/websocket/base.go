package websocket

import (
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

type Conn struct {
	*websocket.Conn
}

func init() {
	log.SetFlags(0)
}

// GetConn 获取连接
func GetConn(host string, rawQuery string) (c Conn, err error) {
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws", RawQuery: rawQuery, ForceQuery: true}
	log.Printf("connecting to %s", u.String())
	c.Conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	return
}

// Write 发送消息
func (c *Conn) Write(msg string) error {
	return c.WriteMessage(websocket.TextMessage, []byte(msg))
}

// Read 读取消息
func (c *Conn) Read() (msg []byte, err error) {
	_, msg, err = c.ReadMessage()
	return
}
