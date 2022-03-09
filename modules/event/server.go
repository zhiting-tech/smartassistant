package event

import (
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"sync"
)

var eventServer *Server
var eventServerOnce sync.Once

type Server struct {
	evenHandlers map[EventType][]HandleFunc
	m            sync.Mutex
}

func GetServer() *Server {
	eventServerOnce.Do(func() {
		eventServer = newEventServer()
	})
	return eventServer
}

func newEventServer() *Server {
	return &Server{
		evenHandlers: make(map[EventType][]HandleFunc),
	}
}

func (s *Server) RegisterHandler(eventType EventType, handlerFunctions ...HandleFunc) {
	s.evenHandlers[eventType] = handlerFunctions
}

func (s *Server) Notify(em *EventMessage) {
	var fn = func(msg EventMessage, handler HandleFunc) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(r)
			}
		}()
		if err := handler(msg); err != nil {
			logger.Error("handler err: ", err)
		}
	}

	s.m.Lock()
	defer s.m.Unlock()
	handlerFunctions, ok := s.evenHandlers[em.EventType]
	if ok {
		for _, h := range handlerFunctions {
			go fn(*em, h)
		}
	}
}
