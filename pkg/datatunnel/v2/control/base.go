package control

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zhiting-tech/smartassistant/pkg/datatunnel/v2/proto"
)

var (
	ErrInvalidRegisterVersion   = fmt.Errorf("invalid register version")
	ErrRegisterValueNotFunction = fmt.Errorf("register value not function")
	ErrFunctionInOrOutTooLess   = fmt.Errorf("function in or out too less")
	ErrInvalidArgSequence       = fmt.Errorf("invalid arg sequence")
	ErrDoubleRegister           = fmt.Errorf("double register")
	ErrClientMethodNotFound     = fmt.Errorf("client method not found")

	// 服务器相关错误
	ErrUnknownMsgType    = NewControlError(UnknownMsgType)
	ErrForbidden         = NewControlError(Forbidden)
	ErrMethodNotFound    = NewControlError(MethodNotFound)
	ErrVersionNotSupport = NewControlError(VersionNotSupport)
	ErrInvalidArgNum     = NewControlError(InvalidArgNum)
	ErrInvalidArgType    = NewControlError(InvalidArgType)

	ErrSendMsg = fmt.Errorf("send msg error")

	// 客户端相关错误
	ErrDoubleWaitResponse = fmt.Errorf("double wait response")
	ErrConnectionFinish   = fmt.Errorf("connection finish")
	ErrTimeout            = fmt.Errorf("timeout")
)

// BinaryCoder 是rpc调用的参数编解码器
type BinaryCoder interface {
	BinaryEncoder
	BinaryDecoder
}

type BinaryDecoder interface {
	Decode(data []byte, t reflect.Type) (v reflect.Value, err error)
}

type BinaryEncoder interface {
	Encode(v reflect.Value) (data []byte, err error)
}

type ProxyControlStreamSender interface {
	Send(m *proto.ProxyControlStreamMsg) error
}

// ProxyControlStreamContext 用于保存当前连接的控制通道协议的状态上下文
type ProxyControlStreamContext struct {
	sender       ProxyControlStreamSender
	msgID        int64
	waitResponse sync.Map // msgID -> channel
	ctx          context.Context
	close        chan struct{}
}

func NewProxyControlStreamContextWithContext(sender ProxyControlStreamSender, ctx context.Context) *ProxyControlStreamContext {
	return &ProxyControlStreamContext{
		sender: sender,
		ctx:    ctx,
		close:  make(chan struct{}),
	}
}

func (c *ProxyControlStreamContext) Send(msg *proto.ProxyControlStreamMsg) (err error) {
	msgID := msg.Hdr.MessageId
	if msg.Hdr.MessageType == proto.ProxyControlStreamMsgType_REQUEST {
		ch := make(chan *proto.ProxyControlStreamMsg, 1)
		if _, load := c.waitResponse.LoadOrStore(msgID, ch); load {
			err = ErrDoubleWaitResponse
			return
		}
	}

	return c.sender.Send(msg)
}

func (c *ProxyControlStreamContext) Close() {
	type CloseSender interface {
		CloseSend() error
	}
	if closeSender, ok := c.sender.(CloseSender); ok {
		closeSender.CloseSend()
	}

	close(c.close)
}

// UntilResponse 永久等待消息响应
func (c *ProxyControlStreamContext) UntilResponse(msgID int64) (msg *proto.ProxyControlStreamMsg, err error) {
	return c.WaitResponseWithTime(msgID, 0)
}

func (c *ProxyControlStreamContext) WaitResponseWithTime(msgID int64, waitTime time.Duration) (msg *proto.ProxyControlStreamMsg, err error) {
	var (
		value interface{}
		load  bool
	)
	if value, load = c.waitResponse.Load(msgID); !load {
		return
	}
	ch := value.(chan *proto.ProxyControlStreamMsg)

	if waitTime > 0 {
		select {
		case msg = <-ch:
		case <-time.After(waitTime):
			c.waitResponse.Delete(msgID)
			err = ErrTimeout
		case <-c.ctx.Done():
			err = ErrConnectionFinish
		case <-c.close:
			err = ErrConnectionFinish
		}
	} else {
		select {
		case msg = <-ch:
		case <-c.ctx.Done():
			err = ErrConnectionFinish
		case <-c.close:
			err = ErrConnectionFinish
		}
	}
	c.waitResponse.Delete(msgID)

	return
}

// NotifyResponse 通知收到消息id为msgID的相应包
func (c *ProxyControlStreamContext) NotifyResponse(msgID int64, msg *proto.ProxyControlStreamMsg) {
	var (
		value interface{}
		load  bool
	)
	if value, load = c.waitResponse.Load(msgID); !load {
		return
	}

	select {
	case value.(chan *proto.ProxyControlStreamMsg) <- msg:
	default:
		// 若走到这里，说明消息id重复，协议错误
	}

}

func (c *ProxyControlStreamContext) NextMsgID() int64 {
	return atomic.AddInt64(&c.msgID, 1)
}

func (c *ProxyControlStreamContext) Context() context.Context {
	return c.ctx
}

type RemoteCaller struct {
	base *ControlBase
	info *clientMethodInfo
}

// Call 调用远程rpc方法
func (caller *RemoteCaller) Call(c *ProxyControlStreamContext, args ...interface{}) (results []interface{}, err error) {
	return caller.CallWithTimeout(c, 0, args...)
}

func (caller *RemoteCaller) CallWithTimeout(c *ProxyControlStreamContext, wait time.Duration, args ...interface{}) (results []interface{}, err error) {
	var (
		response *proto.ProxyControlStreamMsg
	)
	if response, err = caller.base.SendRequestWithTimeout(
		c,
		caller.info.MethodName,
		caller.info.Version,
		wait,
		args...,
	); err != nil {
		return
	}

	if response.Body.StatusCode != Success {
		err = NewControlErrorWithReason(response.Body.StatusCode, response.Body.Reason)
		return
	}

	t := reflect.TypeOf(caller.info.Method)
	num := t.NumOut() - 1
	if len(response.Body.Values) != num {
		err = ErrFunctionInOrOutTooLess
		return
	}

	var v reflect.Value
	for i := 0; i < t.NumOut()-1; i++ {

		if v, err = caller.base.coder.Decode(response.Body.Values[i], t.Out(i)); err != nil {
			return
		}

		results = append(results, v.Interface())
	}

	return
}

// Notify 通知远程服务器，不需要等待回应
func (caller *RemoteCaller) Notify(c *ProxyControlStreamContext, args ...interface{}) (err error) {
	return caller.base.SendNotify(
		c,
		caller.info.MethodName,
		caller.info.Version,
		args...,
	)
}

type ControlBaseOptionFn func(base *ControlBase)

func (fn ControlBaseOptionFn) apply(base *ControlBase) {
	fn(base)
}

func WithCoder(coder BinaryCoder) ControlBaseOptionFn {
	return func(base *ControlBase) {
		base.coder = coder
	}
}

func WithLogger(logger Logger) ControlBaseOptionFn {
	return func(base *ControlBase) {
		base.logger = logger
	}
}

func WithPermissionMethodFn(fn func(*ProxyControlStreamContext, string) bool) ControlBaseOptionFn {
	return func(base *ControlBase) {
		base.permissionMethodFn = fn
	}
}

type clientMethodInfo struct {
	MethodName string
	Version    int32
	MethodType proto.ProxyControlStreamMsgType
	Method     interface{}
}

type ControlBase struct {
	requestVersionMethods map[string][]interface{}
	notifyVersionMethods  map[string][]interface{}
	clientMethodMap       map[string]*clientMethodInfo
	coder                 BinaryCoder
	logger                Logger
	permissionMethodFn    func(*ProxyControlStreamContext, string) bool
}

func NewControlBase(opts ...ControlBaseOptionFn) *ControlBase {
	base := &ControlBase{
		requestVersionMethods: map[string][]interface{}{},
		notifyVersionMethods:  map[string][]interface{}{},
		clientMethodMap:       map[string]*clientMethodInfo{},
	}

	for _, opt := range opts {
		opt.apply(base)
	}

	return base
}

// RegisterRPC 注册rpc方法
func (base *ControlBase) RegisterRPC(methodName string, version int32,
	methodType proto.ProxyControlStreamMsgType, method interface{},
) (err error) {
	var (
		versionMethod []interface{}
		ok            bool
	)

	if version <= 0 {
		err = ErrInvalidRegisterVersion
		return
	}

	if err = base.checkMethod(method); err != nil {
		return
	}

	if methodType == proto.ProxyControlStreamMsgType_REQUEST {
		if versionMethod, ok = base.requestVersionMethods[methodName]; !ok {
			versionMethod = []interface{}{}
		}
	} else {
		if versionMethod, ok = base.notifyVersionMethods[methodName]; !ok {
			versionMethod = []interface{}{}
		}
	}

	if int32(len(versionMethod)) >= version {

		if versionMethod[version-1] != nil {
			err = ErrDoubleRegister
			return
		}

	} else {
		for count := version - int32(len(versionMethod)); count > 0; count-- {
			versionMethod = append(versionMethod, nil)
		}
	}
	versionMethod[version-1] = method
	if methodType == proto.ProxyControlStreamMsgType_REQUEST {
		base.requestVersionMethods[methodName] = versionMethod
	} else {
		base.notifyVersionMethods[methodName] = versionMethod
	}

	return
}

// RegisterClientMethod 注册客户端调用方法
func (base *ControlBase) RegisterClientMethod(methodName string, version int32,
	methodType proto.ProxyControlStreamMsgType, method interface{},
) (err error) {
	t := reflect.TypeOf(method)
	if t.Kind() != reflect.Func {
		err = ErrRegisterValueNotFunction
		return
	}
	name := fmt.Sprintf("%s_%d", methodName, version)
	if _, ok := base.clientMethodMap[name]; !ok {
		base.clientMethodMap[name] = &clientMethodInfo{
			MethodName: methodName,
			Version:    version,
			MethodType: methodType,
			Method:     method,
		}
	} else {
		err = ErrDoubleRegister
	}

	return

}

// checkMethod 检查rpc方法是否符合规范
func (base *ControlBase) checkMethod(method interface{}) (err error) {

	t := reflect.TypeOf(method)
	if t.Kind() != reflect.Func {
		err = ErrRegisterValueNotFunction
		return
	}

	if t.NumIn() < 1 || t.NumOut() < 1 {
		err = ErrFunctionInOrOutTooLess
		return
	}

	if !reflect.DeepEqual(t.In(0), reflect.TypeOf(&ProxyControlStreamContext{})) {
		err = ErrInvalidArgSequence
		return
	}

	return
}

// NewRemoteCaller 创建远端调用器
func (base *ControlBase) NewRemoteCaller(methodName string, version int32) (caller *RemoteCaller, err error) {
	name := fmt.Sprintf("%s_%d", methodName, version)
	info, ok := base.clientMethodMap[name]
	if !ok {
		err = ErrClientMethodNotFound
		return
	}

	caller = &RemoteCaller{
		base: base,
		info: info,
	}
	return
}

// SendRequest 发送rpc请求
func (base *ControlBase) SendRequest(
	c *ProxyControlStreamContext, methodName string,
	version int32, args ...interface{},
) (response *proto.ProxyControlStreamMsg, err error) {
	return base.SendRequestWithTimeout(c, methodName, version, 0, args...)
}

func (base *ControlBase) SendRequestWithTimeout(
	c *ProxyControlStreamContext, methodName string,
	version int32, wait time.Duration, args ...interface{},
) (response *proto.ProxyControlStreamMsg, err error) {
	var (
		values [][]byte = [][]byte{}
	)

	for _, arg := range args {

		t := reflect.TypeOf(arg)
		base.logger.Tracef("encode request %s type %s.%s, value %v", methodName, t.PkgPath(), t.Name(), arg)

		var value []byte
		if value, err = base.coder.Encode(reflect.ValueOf(arg)); err != nil {
			return
		}

		values = append(values, value)
	}

	msg := &proto.ProxyControlStreamMsg{
		Hdr: &proto.ProxyControlStreamMsgHdr{
			MessageType: proto.ProxyControlStreamMsgType_REQUEST,
			MessageId:   c.NextMsgID(),
		},
		Body: &proto.ProxyControlStreamMsgBody{
			Version: version,
			Method:  methodName,
			Values:  values,
		},
	}

	if err = c.Send(msg); err != nil {
		return
	}

	msgId := msg.Hdr.MessageId
	if wait > 0 {
		if response, err = c.WaitResponseWithTime(msgId, wait); err != nil {
			return
		}
	} else {
		if response, err = c.UntilResponse(msgId); err != nil {
			return
		}
	}

	return
}

// SendNotify 发送通知请求
func (base *ControlBase) SendNotify(
	c *ProxyControlStreamContext, methodName string,
	version int32, args ...interface{},
) (err error) {
	var (
		values [][]byte = [][]byte{}
	)

	for _, arg := range args {

		t := reflect.TypeOf(arg)
		base.logger.Tracef("encode request %s type %s.%s, value %v", methodName, t.PkgPath(), t.Name(), arg)

		var value []byte
		if value, err = base.coder.Encode(reflect.ValueOf(arg)); err != nil {
			return
		}

		values = append(values, value)
	}

	msg := &proto.ProxyControlStreamMsg{
		Hdr: &proto.ProxyControlStreamMsgHdr{
			MessageType: proto.ProxyControlStreamMsgType_NOTIFY,
			MessageId:   c.NextMsgID(),
		},
		Body: &proto.ProxyControlStreamMsgBody{
			Version: version,
			Method:  methodName,
			Values:  values,
		},
	}

	if err = c.Send(msg); err != nil {
		return
	}

	return
}

// handleError 处理所有错误信息
func (base *ControlBase) handleError(c *ProxyControlStreamContext, request *proto.ProxyControlStreamMsg, processErr error) (err error) {
	if processErr == nil {
		return
	}

	base.logger.Warnf("%+v", processErr)
	if errors.Is(ErrSendMsg, processErr) {
		return processErr
	}

	var response *proto.ProxyControlStreamMsg
	switch processErr := processErr.(type) {
	case *ControlError:
		response = base.responseMsgWithError(request, processErr)
	default:
		response = base.responseMsgWithError(request, NewControlError(ServerError))
	}

	if request.Hdr.MessageType == proto.ProxyControlStreamMsgType_REQUEST {
		err = c.Send(response)
	}

	return
}

// HandleProxyControlStreamMsg 处理消息
func (base *ControlBase) HandleProxyControlStreamMsg(c *ProxyControlStreamContext, msg *proto.ProxyControlStreamMsg) (err error) {

	base.logger.Tracef("handle %s message, id %d", proto.ProxyControlStreamMsgType_name[int32(msg.Hdr.MessageType)], msg.Hdr.MessageId)

	switch msg.Hdr.MessageType {
	case proto.ProxyControlStreamMsgType_REQUEST:
		err = base.handleError(c, msg, base.handleRequestMsg(c, msg))
	case proto.ProxyControlStreamMsgType_RESPONSE:
		base.handleResponseMsg(c, msg)
	case proto.ProxyControlStreamMsgType_NOTIFY:
		err = base.handleError(c, msg, base.handleNotifyMsg(c, msg))
	default:
		err = base.handleError(c, msg, ErrUnknownMsgType)
	}

	return
}

// getMethod 获取rpc方法
func (base *ControlBase) getMethod(c *ProxyControlStreamContext, version int32,
	methodName string, methodType proto.ProxyControlStreamMsgType,
) (method interface{}, err error) {
	var (
		methods []interface{}
		ok      bool
	)
	if base.permissionMethodFn != nil && !base.permissionMethodFn(c, methodName) {
		err = ErrForbidden
		return
	}

	if methodType == proto.ProxyControlStreamMsgType_REQUEST {
		if methods, ok = base.requestVersionMethods[methodName]; !ok {
			err = ErrMethodNotFound
			return
		}
	} else {
		if methods, ok = base.notifyVersionMethods[methodName]; !ok {
			err = ErrMethodNotFound
			return
		}
	}

	if len(methods) < int(version) {
		err = ErrVersionNotSupport
		return
	}

	if version <= 0 {
		version = int32(len(methods))
	}

	method = methods[version-1]
	if method == nil {
		err = ErrVersionNotSupport
		return
	}

	return
}

// callMethodByMsg 根据消息调用请求
func (base *ControlBase) callMethodByMsg(c *ProxyControlStreamContext, msg *proto.ProxyControlStreamMsg) (results []reflect.Value, err error) {
	var (
		method interface{}
	)

	if method, err = base.getMethod(c, msg.Body.Version, msg.Body.Method, msg.Hdr.MessageType); err != nil {
		return
	}

	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			err = fmt.Errorf("recover %v", recoverErr)
		}
	}()

	base.logger.Debugf("call method %s", msg.Body.Method)
	if results, err = base.call(c, msg, method); err != nil {
		return
	}

	return
}

// call 调用注册的rpc请求
func (base *ControlBase) call(c *ProxyControlStreamContext, msg *proto.ProxyControlStreamMsg, method interface{}) (results []reflect.Value, err error) {
	t := reflect.TypeOf(method)
	v := reflect.ValueOf(method)

	args := []reflect.Value{reflect.ValueOf(c)}
	num := t.NumIn()

	if (num - 1) != len(msg.Body.Values) {
		err = ErrInvalidArgNum
		return
	}

	if num > 1 {
		for i := 1; i < num; i++ {

			var arg reflect.Value

			argType := t.In(i)
			base.logger.Tracef("decode method %s type %s.%s", msg.Body.Method, argType.PkgPath(), argType.Name())
			if arg, err = base.coder.Decode(msg.Body.Values[i-1], argType); err != nil {
				return
			}
			base.logger.Tracef("after decode method %s value %v", msg.Body.Method, arg.Interface())

			args = append(args, arg)
		}
	}

	results = v.Call(args)
	return
}

func (base *ControlBase) resultsContainError(results []reflect.Value) (has bool, err error) {
	if len(results) == 0 {
		return
	}

	num := len(results)
	err, has = results[num-1].Interface().(error)

	return
}

func (base *ControlBase) responseMsgWithError(requestMsg *proto.ProxyControlStreamMsg, err *ControlError) (responseMsg *proto.ProxyControlStreamMsg) {
	responseMsg = &proto.ProxyControlStreamMsg{
		Hdr: &proto.ProxyControlStreamMsgHdr{
			MessageId:   requestMsg.Hdr.MessageId,
			MessageType: proto.ProxyControlStreamMsgType_RESPONSE,
		},
		Body: &proto.ProxyControlStreamMsgBody{
			Version:    requestMsg.Body.Version,
			Method:     requestMsg.Body.Method,
			StatusCode: err.GetCode(),
			Reason:     err.GetReason(),
		},
	}

	return
}

func (base *ControlBase) resultToResponseMsg(requestMsg *proto.ProxyControlStreamMsg, results []reflect.Value) (responseMsg *proto.ProxyControlStreamMsg, err error) {
	responseMsg = &proto.ProxyControlStreamMsg{
		Hdr: &proto.ProxyControlStreamMsgHdr{
			MessageId:   requestMsg.Hdr.MessageId,
			MessageType: proto.ProxyControlStreamMsgType_RESPONSE,
		},
		Body: &proto.ProxyControlStreamMsgBody{
			Version:    requestMsg.Body.Version,
			Method:     requestMsg.Body.Method,
			StatusCode: Success,
		},
	}

	if len(results) > 0 {
		responseMsg.Body.Values = [][]byte{}
	}

	for _, result := range results {

		t := result.Type()
		base.logger.Tracef("encode method %s result type %s.%s, value %v", requestMsg.Body.Method, t.PkgPath(), t.Name(), result.Interface())

		var value []byte
		if value, err = base.coder.Encode(result); err != nil {
			return
		}

		responseMsg.Body.Values = append(responseMsg.Body.Values, value)
	}
	return
}

// handleRequestMsg 处理rpc调用请求
func (base *ControlBase) handleRequestMsg(c *ProxyControlStreamContext, msg *proto.ProxyControlStreamMsg) (err error) {
	var (
		results     []reflect.Value
		responseMsg *proto.ProxyControlStreamMsg
		has         bool
	)

	base.logger.Tracef("ready to call method %s", msg.Body.Method)
	if results, err = base.callMethodByMsg(c, msg); err != nil {
		return
	}

	if has, err = base.resultsContainError(results); has {
		return
	}
	if responseMsg, err = base.resultToResponseMsg(msg, results[:len(results)-1]); err != nil {
		return
	}

	if err = c.Send(responseMsg); err != nil {
		base.logger.Warnf("send error %v", err)
		err = ErrSendMsg
		return
	}

	return
}

// handleResponseMsg 处理rpc调用返回的结果
func (base *ControlBase) handleResponseMsg(c *ProxyControlStreamContext, msg *proto.ProxyControlStreamMsg) {
	c.NotifyResponse(msg.Hdr.MessageId, msg)
}

// handleNotifyMsg 处理通知类型的消息
func (base *ControlBase) handleNotifyMsg(c *ProxyControlStreamContext, msg *proto.ProxyControlStreamMsg) (err error) {
	var (
		results []reflect.Value
	)

	base.logger.Tracef("ready to notify %s", msg.Body.Method)
	if results, err = base.callMethodByMsg(c, msg); err != nil {
		return
	}

	_, err = base.resultsContainError(results)
	return
}
