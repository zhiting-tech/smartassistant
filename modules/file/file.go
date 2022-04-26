package file

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	"github.com/zhiting-tech/smartassistant/modules/utils/file"
	urlutils "github.com/zhiting-tech/smartassistant/modules/utils/url"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// UploadFileOption 上传文件参数
type UploadFileOption struct {
	InitUploadServerOption
	UploadUserID int
	FileType     file.FileType
}

// InitUploadServerOption 初始化上传文件服务参数
type InitUploadServerOption struct {
	FileAuth int
	Req      *http.Request
	Hash     string
	FileName string
	Open     multipart.File
}

type UploadServer interface {
	Upload() (string, error)
	GetUrl(info entity.FileInfo) (string, error)
	GetStorage() int
	GetFileAuth() int
}

type Loc struct {
	req      *http.Request
	Storage  int
	HashStr  string
	File     multipart.File
	FileAuth int
	FileName string
}

func (loc *Loc) Upload() (url string, err error) {
	// 本地存储
	// 存文件
	fileDir := path.Join(config.GetConf().SmartAssistant.RuntimePath, "run", "smartassistant", "file")
	if err = os.MkdirAll(fileDir, os.ModePerm); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return "", err
	}
	saveFileName := fmt.Sprintf("%s%s", loc.HashStr, filepath.Ext(loc.FileName))
	targetPath := path.Join(fileDir, saveFileName)
	targetFile, err := os.Create(targetPath)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return "", err
	}
	_, err = io.Copy(targetFile, loc.File)
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return "", err
	}
	targetFile.Close()
	//进行hash校验
	hash := file.SHA256File(targetPath)
	if hash != loc.HashStr {
		os.Remove(targetPath)
		err = errors.New(status.FileHashCheckErr)
		return "", err
	}

	return urlutils.FileUrl(loc.req, saveFileName), nil
}

func (loc *Loc) GetUrl(fileInfo entity.FileInfo) (string, error) {
	return urlutils.FileUrl(loc.req, fmt.Sprintf("%s%s", fileInfo.Hash, fileInfo.Extension)), nil
}

func (loc *Loc) GetStorage() int {
	return loc.Storage
}

func (loc *Loc) GetFileAuth() int {
	return loc.FileAuth
}

func UploadFile(option UploadFileOption) (id int, url string, err error) {
	uploadServer := GetUploadFileManager().GetUploadServer(option.InitUploadServerOption)
	url, err = uploadServer.Upload()
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	ext := filepath.Ext(option.FileName)
	fileInfo := &entity.FileInfo{
		Hash:        option.Hash,
		Extension:   ext,
		StorageType: uploadServer.GetStorage(),
		FileAuth:    uploadServer.GetFileAuth(),
		FileType:    int(option.FileType),
		Name:        option.FileName,
		UserID:      option.UploadUserID,
	}

	if err = entity.CreateFileInfo(fileInfo); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	id = fileInfo.ID
	return
}

// GetFileUrl 获取文件url
func GetFileUrl(c *gin.Context, fileId int) (string, error) {
	fileInfo, err := entity.GetFileInfo(fileId)
	if err != nil {
		logger.Warnf("get file err %s", err)
		return "", err
	}
	uploadServer := GetUploadFileManager().GetUploadServer(InitUploadServerOption{
		FileAuth: fileInfo.FileAuth,
		Req:      c.Request,
	})
	return uploadServer.GetUrl(fileInfo)
}
