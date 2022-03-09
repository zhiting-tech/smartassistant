package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与日志相关错误码
const (
	UploadErr = iota + 9000
	UploadSCErr
	ZipErr
	LogLengthErr
)

func init() {
	errors.NewCode(UploadErr, "上传失败")
	errors.NewCode(UploadSCErr, "传输到远端失败")
	errors.NewCode(ZipErr, "压缩失败")
	errors.NewCode(LogLengthErr, "日志文件大小超过限制")
}
