package status



import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与文件相关的响应状态码
const (
	FileHashCheckErr = iota + 11000
	FileTypeNoSupport
)

func init() {
	errors.NewCode(FileHashCheckErr, "文件hash校验不正确")
	errors.NewCode(FileTypeNoSupport, "上传文件类型不支持")
}
