package errors

import (
	"google.golang.org/grpc/codes"
)

/*
	grpc默认错误码请参考"google.golang.org/grpc/codes"
*/

const (
	DeviceOffline             codes.Code = 10001
	DeviceNotExit             codes.Code = 10002
	DeviceAttributeNotExist   codes.Code = 10003
	DeviceGetAttributeFailure codes.Code = 10004
	DeviceSetAttributeFailure codes.Code = 10005
	DeviceAuthFailure         codes.Code = 10006
	// ToDo 补充错误类型
)
