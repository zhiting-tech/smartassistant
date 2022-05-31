package file

import (
	errors2 "errors"
	"github.com/gin-gonic/gin"
	"github.com/zhiting-tech/smartassistant/modules/api/utils/response"
	"github.com/zhiting-tech/smartassistant/modules/file"
	"github.com/zhiting-tech/smartassistant/modules/types/status"
	fileutils "github.com/zhiting-tech/smartassistant/modules/utils/file"
	"github.com/zhiting-tech/smartassistant/modules/utils/session"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"path/filepath"
)

type fileType string

var (
	fileTypeImg fileType = "img" // 图片
)

type fileUploadReq struct {
	FileHash string
	FileType fileType // 文件类型
}

type fileUploadResp struct {
	FileId  int    `json:"file_id"`
	FileUrl string `json:"file_url"`
}

func FileUpload(c *gin.Context) {
	var (
		req  fileUploadReq
		resp fileUploadResp
		err  error
	)
	defer func() {
		response.HandleResponse(c, err, resp)
	}()

	req.GetFormParam(c)
	user := session.Get(c)
	if user == nil {
		err = errors.Wrap(err, status.InvalidUserCredentials)
		return
	}

	if err = req.validateRequest(); err != nil {
		return
	}

	switch req.FileType {
	case fileTypeImg:
		if err = req.validateImg(c); err != nil {
			return
		}
	default:
		err = errors.New(status.FileTypeNoSupport)
		return
	}
	resp.FileId, resp.FileUrl, err = req.uploadFile(c)
	if err != nil {
		return
	}
	return
}

func (req *fileUploadReq) GetFormParam(c *gin.Context) {
	req.FileHash = c.Request.FormValue("file_hash")
	req.FileType = fileType(c.Request.FormValue("file_type"))
	return
}

func (req *fileUploadReq) validateRequest() (err error) {
	if req.FileHash == "" {
		err = errors.New(errors.BadRequest)
		return
	}
	return
}

func (req *fileUploadReq) validateImg(c *gin.Context) (err error) {
	fileUpload, err := c.FormFile("file_upload")
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	if !fileutils.IsImage(filepath.Ext(fileUpload.Filename)) {
		return errors.Wrap(errors2.New("invalid img file format"), errors.BadRequest)
	}
	return
}

func (req *fileUploadReq) uploadFile(c *gin.Context) (id int, url string, err error) {
	fileUpload, err := c.FormFile("file_upload")
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	open, err := fileUpload.Open()
	if err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}
	defer open.Close()
	return file.UploadFile(file.UploadFileOption{
		InitUploadServerOption: file.InitUploadServerOption{
			Req:      c.Request,
			Hash:     req.FileHash,
			FileName: fileUpload.Filename,
			Open:     open,
		},
		UploadUserID: session.Get(c).UserID,
		FileType:     fileutils.ImageType,
	})
}
