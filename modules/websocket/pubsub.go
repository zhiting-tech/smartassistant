package websocket

import (
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

func NewPubSub() *PubSubTrie {
	return &PubSubTrie{}
}

// PubSubTrie 支持主题前缀匹配
// TODO 考虑持久化
type PubSubTrie struct {
	children sync.Map // int32 -> *PubSubTrie

	subscribers sync.Map // uuid -> Subscriber
}

type SubscribeFunc func(msg interface{}) error

type Subscriber struct {
	n                *PubSubTrie
	UUID             string
	fn               SubscribeFunc
	activeWithPrefix bool
}

func (s *Subscriber) Unsubscribe() {
	s.n.subscribers.Delete(s.UUID)
}
func (t *PubSubTrie) getChild(word string) *PubSubTrie {
	node := t
	for _, ch := range word {
		if ch == '*' {
			continue
		}
		node = node.child(ch)
	}
	return node
}

func (t *PubSubTrie) child(ch int32) *PubSubTrie {
	trie := &PubSubTrie{}
	if v, ok := t.children.LoadOrStore(ch, trie); ok {
		trie = v.(*PubSubTrie)
	}
	return trie
}

// Subscribe 订阅主题，支持前缀匹配
func (t *PubSubTrie) Subscribe(word string, fn SubscribeFunc) Subscriber {
	node := t.getChild(word)
	s := Subscriber{n: node, UUID: uuid.New().String(), fn: fn}
	if strings.HasSuffix(word, "*") {
		s.activeWithPrefix = true
	}
	node.subscribers.Store(s.UUID, s)
	return s
}

// Publish 发布消息
func (t *PubSubTrie) Publish(topic string, msg interface{}) {
	node := t
	length := len(topic)
	logger.Debugf("publish %s", topic)
	for i, ch := range topic {
		v, ok := node.children.Load(ch)
		if !ok {
			return
		}
		// logger.Debugf("node %s exec, %d children, %d subscribers, isEnd %v",
		// 	string(ch), len(node.children), len(node.subscribers), length == i+1)
		node, ok = v.(*PubSubTrie)
		if !ok {
			return
		}
		node.exec(msg, length == i+1)
	}
}

// exec 发布时执行订阅者的任务
func (t *PubSubTrie) exec(msg interface{}, isEnd bool) {
	node := t

	node.subscribers.Range(func(key, value interface{}) bool {
		tfn := value.(Subscriber)
		if isEnd || tfn.activeWithPrefix {
			go func(tfn Subscriber) {
				defer func() {
					if r := recover(); r != nil {
						logger.Errorf("pubsub func panic %s", r)
					}
				}()
				tfn.fn(msg) // TODO optimize
			}(tfn)
		}
		return true
	})
}
