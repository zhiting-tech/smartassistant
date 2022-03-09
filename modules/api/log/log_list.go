package log

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/config"
)

type LogInfo struct {
	Level   string `json:"level"`
	Path    string `json:"caller"`
	Message string `json:"msg"`
	Time    string `json:"time"`
}

type ListResponse struct {
	Logs     []LogInfo `json:"logs"`
	LastDay  string    `json:"last_day"`
	LastLine int       `json:"last_line"`
}

// CST
var loc, _ = time.LoadLocation("Asia/Chongqing")

// LogList 日志列表
func LogList(c *gin.Context) {
	var (
		err  error
		resp ListResponse
	)
	// 获取当前时间
	t := time.Now()
	// 筛选日期 获取日期 如果没有则默认使用当天日期作为文件名
	dt, _ := strconv.Atoi(c.DefaultQuery("date", t.Format("20060102")))
	// 错误类型 all=全部  info warning error
	tp := c.DefaultQuery("type", "all")
	// 页数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	// 游标
	lastLine := c.Query("last_line")

	defer func() {
		if resp.Logs == nil {
			resp.Logs = make([]LogInfo, 0)
		}

		if lastLine != "" {
			resp.LastLine = 0
		}

		response.HandleResponse(c, err, &resp)
	}()

	// 获取根目录路径
	dir := config.GetConf().SmartAssistant.RuntimePath
	path := fmt.Sprintf("%s/log/smartassistant.%d.log", dir, dt)

	logList := getLogFile(true)
	resp.LastDay = ""

	// 获取文件最近的一个文件名
	match := matchDay(dt, logList)
	if match != 0 {
		resp.LastDay = strconv.Itoa(match)
	}

	resp.Logs, resp.LastLine, _ = getLog(path, tp, page, lastLine)

	return
}

// getLog 分页获取日志内容和筛选条件
func getLog(path string, tp string, page int, last string) (listLogs []LogInfo, lastLine int, err error) {
	var (
		rep   LogInfo
		field string
	)

	// 最后一行的大小
	lastLine = 0

	// 开始扫描文件内容
	f, err := os.Open(path)
	if err != nil {
		return
	}

	stat, _ := f.Stat()
	defer f.Close()
	if err != nil {
		return
	}

	size := int(stat.Size())

	// 如果客户端未传值则需要返回当前的文件的大小
	if last == "" {
		lastLine = size
	} else {
		// 如果已经传值，则使用传值的大小进行获取
		size, err = strconv.Atoi(last)
		if err != nil {
			return
		}
	}

	scanner := BackScannerNew(f, size)

	fieldArr := map[string]string{
		"info":    "info",
		"warning": "warning",
		"error":   "error",
		"other":   "other",
	}

	if value, ok := fieldArr[tp]; ok {
		field = value
	}

	// 获取总条数
	pagesize := 50
	// 分页筛选
	offset := (page - 1) * pagesize
	// 键
	index := 0

	for {
		rep = LogInfo{}
		// 数量已够
		if len(listLogs) == pagesize {
			break
		}
		// 从后到前一条条扫描
		line, _, err := scanner.Line()
		if err != nil {
			break
		}

		// 筛选类型不为空 AND 类型不为其他 AND 当前类型不匹配
		if field != "" {
			err = json.Unmarshal([]byte(line), &rep)
			if err != nil {
				continue
			}

			if field != "other" {
				if rep.Level != field {
					continue
				}
			} else {
				// 类型不为空 AND 类型等于其他 AND 类型排除这三种的
				if rep.Level == fieldArr["info"] || rep.Level == fieldArr["warning"] || rep.Level == fieldArr["error"] {
					continue
				}
			}
		}

		index += 1
		// 当前键值小于跳过的值
		if index > offset {
			// 如果筛选字段为空
			if field == "" {
				err = json.Unmarshal([]byte(line), &rep)
				if err != nil {
					continue
				}
			}
			// 格式化时间戳
			rep.Time = timeFormat(rep.Time)
			listLogs = append(listLogs, rep)
		}
	}

	return
}

// timeFormat 格式化时间
func timeFormat(timer string) string {
	to, _ := time.Parse(time.RFC3339, timer)
	return to.In(loc).Format("2006-01-02 15:04")
}

// 两个相邻最近日期的日志
func matchDay(num int, logList []string) int {
	var match, min, day int
	min = -1
	for _, v := range logList {
		day, _ = strconv.Atoi(v)
		if day == num {
			continue
		}
		contrast := int(math.Abs(float64(day - num)))
		if contrast < min || min == -1 {
			min = contrast
			match = day
		}
	}

	if match > num {
		match = 0
	}

	return match
}

// 遍历日志文件夹下的日志文件追加到切片中
func getLogFile(day bool) []string {
	// 获取根目录路径
	dir := config.GetConf().SmartAssistant.RuntimePath
	// 获取日志路径
	path := fmt.Sprintf("%s/log/", dir)

	var arr = make([]string, 0)
	var logList = make([]string, 0)
	// 获取文件夹下的文件追加进arr切片
	filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		arr = append(arr, path)
		return nil
	})

	// 循环文件夹
	for _, v := range arr {
		// 切割文件名
		str := strings.FieldsFunc(v, split)
		// 判断文件的切割后的日志文件
		if len(str) >= 3 && str[2] == "log" {
			if day {
				logList = append(logList, str[1])
				continue
			}

			logList = append(logList, v)
		}
	}

	return logList
}

// 字符串以.号切割
func split(s rune) bool {
	return s == '.'
}

// reverse 数组倒序函数
func reverse(arr *[]string, length int) {
	var temp string
	for i := 0; i < length/2; i++ {
		temp = (*arr)[i]
		(*arr)[i] = (*arr)[length-1-i]
		(*arr)[length-1-i] = temp
	}
}
