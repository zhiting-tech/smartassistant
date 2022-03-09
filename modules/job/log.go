package job

import (
	"fmt"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var arr = make([]string, 0)

// LogRemove 定时删除超过七天的日志
func LogRemove() {
	// 检测日志是否过期
	// 七天前日期
	lastDay := time.Unix(time.Now().Unix()-(3600*24*7), 0).Format("20060102")
	// 获取根目录路径
	dir := config.GetConf().SmartAssistant.RuntimePath
	// 拼接文件路径
	path := fmt.Sprintf("%s/log", dir)
	// 获取文件夹下的文件追加进arr切片
	filepath.Walk(path, walkfunc)

	// 循环文件夹
	for k, v := range arr {
		// 切割文件名
		str := strings.FieldsFunc(v, split)
		// 判断文件的切割后的日志文件 并且日志时间大于限制时间的则执行删除
		if len(str) >= 3 && str[0] == "smartassistant" && str[2] == "log" && str[1] <= lastDay {
			// 删除文件
			os.Remove(arr[k])
		}
	}
}

// 检测文件夹下的文件和文件夹
func walkfunc(path string, info os.FileInfo, err error) error {
	arr = append(arr, path)

	return nil
}

// 字符串以.号切割
func split(s rune) bool {
	if s == '.' {
		return true
	}
	return false
}
