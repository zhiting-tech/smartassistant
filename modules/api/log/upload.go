package log

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"io"
	"io/fs"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Resp struct {
	Status int    `json:"status"`
	Reason string `json:"reason"`
}

// Upload 主动上传日志
func Upload(c *gin.Context) {
	var (
		err error
		rep *http.Response
	)

	defer func() {
		response.HandleResponse(c, err, nil)
	}()

	// 获取根目录路径
	dir := config.GetConf().SmartAssistant.RuntimePath
	// 拼接文件路径
	path := fmt.Sprintf("%s/log", dir)

	logZipPath := path + "/log.zip"

	// 压缩日志文件夹里的日志
	lastDay, err := com(path, "log.zip")
	if err != nil {
		err = errors.Wrap(err, status.ZipErr)
		return
	}
	// 删除压缩文件
	defer os.Remove(logZipPath)

	url := config.GetConf().SmartCloud.URL()
	if url == "" {
		errors.New(status.UploadErr)
		return
	}

	// 判断文件是否存在
	exists, err := PathExists(logZipPath)
	if err == nil && exists {
		// 发送日志到SC
		rep, err = postFile(logZipPath, url+"/log_replay/zip")
		if err != nil {
			err = errors.Wrap(err, status.UploadSCErr)
			return
		}
		repContent, _ := ioutil.ReadAll(rep.Body)
		var repJson Resp

		err = json.Unmarshal(repContent, &repJson)
		if err != nil || repJson.Status != 0 {
			err = errors.Wrap(err, status.UploadSCErr)
			return
		}

		lastDayPath := path + "/log_last_day"
		// 写入文件记录最后一个文件日期和行数
		ioutil.WriteFile(lastDayPath, []byte(lastDay), 0644)
	}

	return
}

// 循环文件夹调用压缩，记录文件日期等操作
func com(path string, comPath string) (lastDay string, err error) {
	var arr = make([]string, 0)
	// 获取文件夹下的文件追加进arr切片
	filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		arr = append(arr, path)
		return nil
	})

	var files []*os.File

	lastDayPath := path + "/log_last_day"

	lastDayFile, err := os.Open(lastDayPath)
	defer lastDayFile.Close()

	last, err := ioutil.ReadAll(lastDayFile)

	if err != nil {
		return "", err
	}

	lastArray := strings.Split(string(last), ",")

	lastDayString := 0
	lastDayLine := 0

	if lastArray[0] != "" && lastArray[1] != "" {
		lastDayString, err = strconv.Atoi(lastArray[0])
		lastDayLine, err = strconv.Atoi(lastArray[1])
	}

	if err != nil {
		return "", err
	}

	// 键
	index := 0

	logPath := path + "/output.log"
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
		return
	}
	//及时关闭file句柄
	defer file.Close()
	defer os.Remove(logPath)

	//写入文件时，使用带缓存的 *Writer
	write := bufio.NewWriter(file)
	check := false

	// 循环文件夹
	for k, v := range arr {
		// 切割文件名
		str := strings.FieldsFunc(v, split)

		// 判断文件的切割后的日志文件
		if len(str) >= 3 && str[0] == path+"/smartassistant" && str[2] == "log" {
			day, _ := strconv.Atoi(str[1])
			// 判断文件日期是否大于上传记录的最大日期
			if lastDayString != 0 && lastDayString > day {
				continue
			}
			f, _ := os.Open(arr[k])
			defer f.Close()

			scanner := bufio.NewScanner(f)

			index = 0

			for scanner.Scan() {
				index += 1
				// 当前键值小于跳过的值
				if lastDayString == day && index <= lastDayLine {
					continue
				}

				write.WriteString(scanner.Text() + "\n")
				if !check {
					check = true
				}
			}

			// 文件日期,文件行数
			lastDay = str[1] + "," + strconv.Itoa(index)
		}
	}

	if check {
		//Flush将缓存的文件真正写入到文件中
		write.Flush()

		openLogPath, _ := os.Open(logPath)
		defer openLogPath.Close()

		files = append(files, openLogPath)
		// 压缩文件
		err = Compress(files, path+"/"+comPath)
		if err != nil {
			return "", nil
		}
	}

	return lastDay, nil
}

// Compress 压缩文件
// files 文件数组，可以是不同dir下的文件或者文件夹
// dest 压缩文件存放地址
func Compress(files []*os.File, dest string) error {
	// 创建压缩文件
	d, _ := os.Create(dest)
	defer d.Close()
	// NewWriter创建并返回一个将zip文件写入w的*Writer。
	w := zip.NewWriter(d)
	defer w.Close()

	for _, file := range files {
		err := compress(file, "logs", w)
		if err != nil {
			return err
		}
	}
	return nil
}

// 执行压缩
func compress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		if len(prefix) == 0 {
			prefix = info.Name()
		} else {
			prefix = prefix + "/" + info.Name()
		}
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		// FileInfoHeader返回一个根据fi填写了部分字段的Header。
		header, err := zip.FileInfoHeader(info)
		if len(prefix) != 0 {
			header.Name = prefix + "/" + header.Name
		}
		if err != nil {
			return err
		}
		header.Method = zip.Deflate
		// 使用给出的*FileHeader来作为文件的元数据添加一个文件进zip文件。
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		// 将文件复制到目标
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// post传输文件到远程
func postFile(filename string, targetUrl string) (*http.Response, error) {
	bodyBuf := bytes.NewBufferString("")
	bodyWriter := multipart.NewWriter(bodyBuf)

	// use the body_writer to write the Part headers to the buffer
	_, err := bodyWriter.CreateFormFile("zip", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return nil, err
	}

	// the file data will be the second part of the body
	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return nil, err
	}
	// need to know the boundary to properly close the part myself.
	boundary := bodyWriter.Boundary()
	//close_string := fmt.Sprintf("\r\n--%s--\r\n", boundary)
	closeBuf := bytes.NewBufferString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	// use multi-reader to defer the reading of the file data until
	// writing to the socket buffer.
	requestReader := io.MultiReader(bodyBuf, fh, closeBuf)
	fi, err := fh.Stat()
	if err != nil {
		fmt.Printf("Error Stating file: %s", filename)
		return nil, err
	}

	req, err := http.NewRequest("POST", targetUrl, requestReader)
	if err != nil {
		return nil, err
	}

	// 请求SC授权信息
	username := config.GetConf().SmartAssistant.ID
	password := config.GetConf().SmartAssistant.Key
	req.SetBasicAuth(username, password)

	// Set headers for multipart, and Content Length
	req.Header.Add("Content-Type", "multipart/form-data; boundary="+boundary)
	req.ContentLength = fi.Size() + int64(bodyBuf.Len()) + int64(closeBuf.Len())

	// 设置请求超时
	var c = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second * 60,
		},
	}

	return c.Do(req)
}

/*
   判断文件或文件夹是否存在
   如果返回的错误为nil,说明文件或文件夹存在
   如果返回的错误类型使用os.IsNotExist()判断为true,说明文件或文件夹不存在
   如果返回的错误为其它类型,则不确定是否在存在
*/
func PathExists(path string) (bool, error) {

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
