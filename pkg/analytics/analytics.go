package analytics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

type EventType string

const (
	EventTypeDeviceAdd    EventType = "device_add"
	EventTypeDeviceDelete EventType = "device_delete"
	EventTypePluginAdd    EventType = "plugin_add"
	EventTypePluginDelete EventType = "plugin_delete"
)

type Event struct {
	Type        EventType              `json:"type"`
	CreatedAt   time.Time              `json:"created_at"`
	EsTimeStamp int64                  `json:"@timestamp"`
	UserId      int                    `json:"user_id"`
	Values      map[string]interface{} `json:"values"` // 考虑通用性，values根据不同type进行不同的序列化/反序列化
}

var collectCh chan Event
var sendCh chan Event
var configs *config.Options

func Start(conf *config.Options) {
	sendCh = make(chan Event, 100)
	collectCh = make(chan Event, 100)
	configs = conf
	go insertServer()
}

func insertServer() {
	timer := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-timer.C:
			events := collectEvents()
			go send(events)
		case event := <-sendCh:
			collectCh <- event
			if len(collectCh) == cap(collectCh) {
				events := collectEvents()
				go send(events)
			}
		}
	}
}

func collectEvents() []Event {
	events := make([]Event, 0, 100)
	length := len(collectCh)
	for i := 0; i < length; i++ {
		events = append(events, <-collectCh)
	}
	return events
}

func Record(eventType EventType, userId int, value ...map[string]interface{}) {
	now := time.Now()
	event := Event{Type: eventType, UserId: userId, CreatedAt: now, EsTimeStamp: now.Unix()}
	if len(value) != 0 {
		event.Values = value[0]
	}
	sendCh <- event
}

func RecordStruct(eventType EventType, userId int, str interface{}) error {
	values, err := StructToMap(str)
	if err != nil {
		return err
	}
	Record(eventType, userId, values)
	return nil
}

func StructToMap(v interface{}) (map[string]interface{}, error) {
	values := map[string]interface{}{}
	bytes, err := json.Marshal(v)
	if err != nil {
		return values, err
	}
	err = json.Unmarshal(bytes, &values)
	return values, err
}

type reqBody struct {
	Events []Event `json:"events"`
}

func send(events []Event) {
	if len(events) == 0 {
		return
	}
	hostname := configs.SmartCloud.URL()
	url := fmt.Sprintf("%s/log_replay/analytics", hostname)

	data := &reqBody{}
	data.Events = events
	b, err := json.Marshal(data)
	if err != nil {
		logger.Error(err)
		return
	}
	buf := bytes.NewBuffer(b)
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", url, buf)
	if err != nil {
		logger.Error(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(configs.SmartAssistant.ID, configs.SmartAssistant.Key)
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return
	}
	defer resp.Body.Close()
}
