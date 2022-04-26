package file

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
)

// UploadServerType 存储服务类型
type UploadServerType int

// FileType 文件类型
type FileType int

// UploadFileRole 上传文件权限
type UploadFileRole int

const (
	UploadFileTypeLocal UploadServerType = iota + 1 // 本地
	UploadFileTypeAliyun
)
const (
	UploadFilePublic  UploadFileRole = iota + 1
	UploadFilePrivate // 私有权限
)

const (
	ImageType FileType = iota + 1 // 图片类型
)

var imageSuffix = []string{".png", ".jpeg", ".jpg", ".gif"}

// SHA256File 文件sha256
func SHA256File(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	h := sha256.New()
	_, err = io.Copy(h, file)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// IsImage 根据后缀判断文件是否是图片
func IsImage(ext string) bool {
	for _, e := range imageSuffix {
		if e == strings.ToLower(ext) {
			return true
		}
	}
	return false
}
