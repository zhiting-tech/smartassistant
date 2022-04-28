package file

import (
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/utils/file"
	"sync"
)

var (
	once              sync.Once
	uploadFileManager *UploadFileManager
)

type initUploadServerFunc func(option InitUploadServerOption) (server UploadServer)

type UploadFileManager struct {
	uploadServersFunc map[string]initUploadServerFunc
}

// RegisterFunc 注册不同上传服务初始化函数
func (um *UploadFileManager) RegisterFunc(driver string, initFunc initUploadServerFunc) {
	if driver == "local" {
		return
	}
	um.uploadServersFunc[driver] = initFunc
}

func (um *UploadFileManager) GetUploadServer(option InitUploadServerOption) UploadServer {
	if serverFunc, ok := um.uploadServersFunc[config.GetConf().Oss.Driver]; ok {
		return serverFunc(option)
	}
	return &Loc{
		req:      option.Req,
		Storage:  int(file.UploadFileTypeLocal),
		HashStr:  option.Hash,
		File:     option.Open,
		FileAuth: int(file.UploadFilePrivate),
		FileName: option.FileName,
	}
}

func GetUploadFileManager() *UploadFileManager {
	once.Do(func() {
		uploadFileManager = &UploadFileManager{
			uploadServersFunc: make(map[string]initUploadServerFunc),
		}
	})
	return uploadFileManager
}
