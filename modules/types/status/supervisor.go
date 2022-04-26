package status

import "github.com/zhiting-tech/smartassistant/pkg/errors"

// 与系统管理相关的响应状态码

const (
	FileNotExistErr = iota + 7000
	FileIsDirErr
	ImagePullErr
	GetImageVersionErr
	ParamRequireErr
	ChecksumErr
	FirmwareDownloadErr
	GetFirmwareVersionErr
	BackupTimeLessThenNowErr
	BackupInfoNotExist
	BackupInfoIsNotSuccessErr
	OnBackupExist
)

func init() {
	errors.NewCode(FileNotExistErr, "文件不存在")
	errors.NewCode(FileIsDirErr, "非法访问")
	errors.NewCode(ImagePullErr, "拉取镜像失败")
	errors.NewCode(GetImageVersionErr, "获取版本信息失败")
	errors.NewCode(ParamRequireErr, "缺少必要参数")
	errors.NewCode(ChecksumErr, "固件升级包合法性校验不通过")
	errors.NewCode(FirmwareDownloadErr, "固件升级包下载失败")
	errors.NewCode(GetFirmwareVersionErr, "获取固件版本信息失败")
	errors.NewCode(BackupTimeLessThenNowErr, "定时备份执行时间小于现在时间")
	errors.NewCode(BackupInfoNotExist, "该数据不存在")
	errors.NewCode(BackupInfoIsNotSuccessErr, "该数据执行状态不成功")
	errors.NewCode(OnBackupExist, "有计划备份正在执行")

}
