package websocket

import (
	"sync"

	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// bucket 管理websocket连接的桶
type bucket struct {
	clients    sync.Map // key->*client
	broadcast  chan broadcastData
	register   chan *client
	unregister chan *client
}

type broadcastData struct {
	AreaID uint64
	Data   []byte
}

func newBucket() *bucket {
	return &bucket{
		register:   make(chan *client, 2),
		unregister: make(chan *client, 2),
		broadcast:  make(chan broadcastData),
	}
}

func (b *bucket) run() {
	for {
		select {
		case client := <-b.register:
			//logger.Debug("register new websocket client", client.key)
			b.put(client)
		case client := <-b.unregister:
			//logger.Debug("del websocket client", client.key)
			b.remove(client.key)
		case message := <-b.broadcast:
			b.clients.Range(func(key, value interface{}) bool {
				cli := value.(*client)
				if cli.areaID != message.AreaID {
					return true
				}

				//logger.Debug("broadcast clientKey", cli.key, " AreaID ", cli.areaID)
				select {
				case cli.send <- message.Data:
				}
				return true
			})
		}
	}
}

// 断开客户端，停止
func (b *bucket) stop() {
	b.clients.Range(func(key, value interface{}) bool {
		b.clients.Delete(key.(string))
		cli := value.(*client)
		if err := cli.Close(); err != nil {
			logger.Error("close client err:", err.Error())
		}
		return true
	})
}

func (b *bucket) put(cli *client) {
	b.clients.Store(cli.key, cli)
}
func (b *bucket) get(key string) *client {
	cli, ok := b.clients.Load(key)
	if ok {
		return cli.(*client)
	}
	return nil
}

func (b *bucket) remove(key string) {
	v, loaded := b.clients.LoadAndDelete(key)
	if loaded {
		cli := v.(*client)
		if err := cli.Close(); err != nil {
			logger.Error("close client err:", err.Error())
		}
	}
}
