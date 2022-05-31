package control

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	Success = iota
	ServerError
	UnknownMsgType
	Forbidden
	MethodNotFound
	VersionNotSupport
	InvalidArgNum
	InvalidArgType
)

const (
	CustomCodeStart = iota + 1025
	DuplicateRegisterService
)

var (
	codeReasonMap = map[int32]string{
		ServerError:              "server error",
		UnknownMsgType:           "unknown message type",
		Forbidden:                "forbidden",
		MethodNotFound:           "method not found",
		VersionNotSupport:        "version not support",
		InvalidArgNum:            "invalid arg num",
		InvalidArgType:           "invalid arg type",
		DuplicateRegisterService: "duplicate register service",
	}
)

type ControlError struct {
	err    error
	code   int32
	reason string
}

func NewControlError(code int32) *ControlError {
	return NewControlErrorWithReason(code, "")
}

func NewControlErrorWithReason(code int32, reason string) *ControlError {

	if reason == "" {
		if code < CustomCodeStart {
			reason = codeReasonMap[code]
		}
	}
	return &ControlError{
		code:   code,
		reason: reason,
		err:    errors.New(reason),
	}
}

func Wrap(err error, code int32) *ControlError {
	return WrapWithReason(err, code, "")
}

func WrapWithReason(err error, code int32, reason string) *ControlError {
	controlErr := NewControlErrorWithReason(code, reason)
	controlErr.err = errors.Wrap(err, controlErr.Error())
	return controlErr
}

func (e *ControlError) Error() string {
	reason := e.GetReason()
	if reason == "" {
		reason = ""
	}
	return fmt.Sprintf("Code : %d, Reason : %s", e.GetCode(), reason)
}

func (e *ControlError) GetCode() int32 {
	return e.code
}

func (e *ControlError) GetReason() string {
	return e.reason
}

func (e *ControlError) Unwrap() error {
	return e.err
}

func (e *ControlError) Format(f fmt.State, verb rune) {
	// io.WriteString(f, e.Error())
	type formater interface {
		Format(f fmt.State, verb rune)
	}

	e.err.(formater).Format(f, verb)
}

// func (e *ControlError) GetErrStack() errors.StackTrace {
// 	// 获取错误调用栈(跳过New,Newf,Wrapf,Wrap调用栈)
// 	type stackTracer interface {
// 		StackTrace() errors.StackTrace
// 	}
// 	st := e.err.(stackTracer)
// 	stackTrace := st.StackTrace()

// 	filterStack := errors.StackTrace{}
// 	filterFuncRegex, _ := regexp.Compile(`/control\.(New|Wrap)f?`)

// 	for _, f := range stackTrace[:2] {
// 		stackText, _ := f.MarshalText()
// 		if !filterFuncRegex.MatchString(string(stackText)) {
// 			filterStack = append(filterStack, f)
// 		}
// 	}
// 	filterStack = append(filterStack, stackTrace[2:]...)
// 	return filterStack

// }
