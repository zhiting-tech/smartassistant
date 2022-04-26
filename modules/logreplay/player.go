package logreplay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jinzhu/now"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/http/httpclient"
	"gorm.io/gorm"
)

var (
	player *LogPlayer
	once   sync.Once
)

const (
	disable = iota
	enable
)

func GetLogPlayer() *LogPlayer {
	once.Do(func() {
		player = &LogPlayer{
			ch:          make(chan []byte, 1024),
			toFile:      make(chan []byte, 1024),
			bufferLimit: 81920,
			upload:      disable,
			save:        disable,
		}

		player.init()
	})

	return player
}

type LogPlayer struct {
	save        int32
	upload      int32
	toFile      chan []byte
	ch          chan []byte
	bufferLimit int
	logFile     *os.File
	beforeDay   time.Time
}

func (p *LogPlayer) init() {
	p.checkUpload()
	p.beforeDay = time.Now()
}

func (p *LogPlayer) checkUpload() {
	var (
		area           entity.Area
		err            error
		defaultSetting = entity.GetDefaultLogSetting()
	)

	defer func() {
		if defaultSetting.LogSwitch {
			p.EnableUpload()
		} else {
			p.DisableUpload()
		}
	}()

	if err = entity.GetDB().Order("created_at asc").First(&area).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		fmt.Fprintf(os.Stderr, "[ %s ] check upload get area error %v\n", time.Now().String(), err)
		return
	}

	if err = entity.GetSetting(entity.LogSwitch, &defaultSetting, area.ID); err != nil {
		fmt.Fprintf(os.Stderr, "[ %s ] get area log setting error %v\n", time.Now().String(), err)
		return
	}

}

func (p *LogPlayer) EnableUpload() {
	atomic.StoreInt32(&p.upload, enable)
}

func (p *LogPlayer) DisableUpload() {
	atomic.StoreInt32(&p.upload, disable)
}

func (p *LogPlayer) isUpload() bool {
	return atomic.LoadInt32(&p.upload) != disable
}

func (p *LogPlayer) EnableSave() {
	atomic.StoreInt32(&p.save, enable)
}

func (p *LogPlayer) DisableSave() {
	atomic.StoreInt32(&p.save, disable)
}

func (p *LogPlayer) isEnableSave() bool {
	return atomic.LoadInt32(&p.save) != disable
}

func (p *LogPlayer) openLog() (err error) {
	n := time.Now()
	filePath := path.Join(config.GetConf().SmartAssistant.RuntimePath, "log", fmt.Sprintf("smartassistant.%s.log", n.Format("20060102")))
	os.MkdirAll(path.Dir(filePath), 0777)
	if _, err = os.Stat(filePath); err == nil {
		p.logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ %s ] Open %s error %v\n", time.Now().String(), filePath, err)
		}
	} else {
		if os.IsNotExist(err) {
			err = nil
			if p.logFile, err = os.Create(filePath); err != nil {
				fmt.Fprintf(os.Stderr, "[ %s ] Create %s error %v\n", time.Now().String(), filePath, err)
			}
		}
	}
	p.beforeDay = n

	return
}

func (p *LogPlayer) Recv(conn *websocket.Conn) {
	var (
		t    int
		data []byte
		err  error
	)
	for {
		if t, data, err = conn.ReadMessage(); err != nil || t == websocket.CloseMessage {
			conn.Close()
			return
		}

		if p.isEnableSave() {
			select {
			case p.ch <- data:
			default:
			}
		}
	}
}

func (p *LogPlayer) isNextDay() bool {
	before := now.New(p.beforeDay)
	return before.BeginningOfDay().Unix() < now.BeginningOfDay().Unix()
}

func (p *LogPlayer) appendLogToFile(data []byte) {
	var (
		n   int
		err error
	)

	if p.isNextDay() {
		p.logFile.Close()
		p.openLog()
	}

	if p.logFile == nil {
		if err := p.openLog(); err != nil {
			fmt.Fprintf(os.Stderr, "[ %s ] Append log error %v\n", time.Now().String(), err)
			return
		}
	}

	if n, err = p.logFile.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "[ %s ] Append log error %v\n", time.Now().String(), err)
	} else if n != len(data) {
		fmt.Fprintf(os.Stderr, "[ %s ] Append log too short %d <-> %d\n", time.Now().String(), n, len(data))
	}

}

func (p *LogPlayer) runLogToFile(ctx context.Context) {
	wait := 10 * time.Second
	flush := time.NewTimer(wait)
	for {
		select {
		case <-ctx.Done():
			if p.logFile != nil {
				p.logFile.Close()
			}
			return
		case <-flush.C:
			// 定时同步文件
			if p.logFile != nil {
				p.logFile.Sync()
			}
			// 重置定时器
			flush.Reset(wait)
			select {
			case <-flush.C:
			default:
			}
		case data := <-p.toFile:
			p.appendLogToFile(data)
		}
	}
}

func (p *LogPlayer) Run(ctx context.Context) {
	var (
		timePending, finish bool
		buffer              = bytes.NewBuffer(nil)
		waitTime            = 10 * time.Second
		t                   = time.NewTimer(waitTime)
	)

	defer t.Stop()
	go p.runLogToFile(ctx)

	for {
		select {
		case data := <-p.ch:
			if p.isEnableSave() {
				if p.isUpload() {
					buffer.Write(data)
				}
				p.toFile <- data
			}
		case <-t.C:
			timePending = true
		case <-ctx.Done():
			finish = true
		}

		// 缓冲区超过大小、定时触发、进程停止时触发上送
		if buffer.Len() > p.bufferLimit || finish || timePending {
			if buffer.Len() > 0 {
				p.toRemote(ctx, buffer)
				buffer.Reset()
			}
		}

		if timePending {
			timePending = false
			// 重置定时器
			t.Reset(waitTime)
			select {
			case <-t.C:
			default:
			}
		}

		if finish {
			return
		}
	}
}

func (p *LogPlayer) toRemote(ctx context.Context, reader io.Reader) {
	var (
		url  string
		req  *http.Request
		resp *http.Response
		data []byte
		err  error
	)
	url = fmt.Sprint(config.GetConf().SmartCloud.URL(), "/log_replay")
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, reader); err != nil {
		fmt.Fprintf(os.Stderr, "[ %s ] New request %s error %v\n", time.Now().String(), url, err)
		return
	}
	req.SetBasicAuth(config.GetConf().SmartAssistant.ID, config.GetConf().SmartAssistant.Key)

	if resp, err = httpclient.DefaultClient.Do(req); err != nil {
		log.Print()
		fmt.Fprintf(os.Stderr, "[ %s ] Request %s error %v\n", time.Now().String(), url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "[ %s ] Http response code %d not ok\n", time.Now().String(), resp.StatusCode)
		return
	}

	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	type SCResponse struct {
		Status int    `json:"status"`
		Reason string `json:"reason"`
	}

	scResp := SCResponse{}
	if err = json.Unmarshal(data, &scResp); err != nil {
		return
	}

	if scResp.Status != 0 {
		fmt.Fprintf(os.Stderr, "[ %s ] Response error code:%d, reason:%s\n", time.Now().String(), scResp.Status, scResp.Reason)
		return
	}

}
