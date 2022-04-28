// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.1.0
// - protoc             v3.19.0
// source: supervisor.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SupervisorClient is the client API for Supervisor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SupervisorClient interface {
	Restart(ctx context.Context, in *RestartReq, opts ...grpc.CallOption) (*Response, error)
	Backup(ctx context.Context, in *BackupReq, opts ...grpc.CallOption) (*Response, error)
	Restore(ctx context.Context, in *RestoreReq, opts ...grpc.CallOption) (*Response, error)
	Update(ctx context.Context, in *UpdateReq, opts ...grpc.CallOption) (*Response, error)
	UpdateSystem(ctx context.Context, in *UpdateSystemReq, opts ...grpc.CallOption) (*Response, error)
	GetSystemInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetSystemInfoResp, error)
	GetExtensions(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetExtensionsResp, error)
	RemoteHelp(ctx context.Context, in *RemoteHelpReq, opts ...grpc.CallOption) (*Response, error)
	RemoteHelpEnabled(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*RemoteHelpEnabledResp, error)
}

type supervisorClient struct {
	cc grpc.ClientConnInterface
}

func NewSupervisorClient(cc grpc.ClientConnInterface) SupervisorClient {
	return &supervisorClient{cc}
}

func (c *supervisorClient) Restart(ctx context.Context, in *RestartReq, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/supervisor.proto.Supervisor/Restart", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *supervisorClient) Backup(ctx context.Context, in *BackupReq, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/supervisor.proto.Supervisor/Backup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *supervisorClient) Restore(ctx context.Context, in *RestoreReq, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/supervisor.proto.Supervisor/Restore", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *supervisorClient) Update(ctx context.Context, in *UpdateReq, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/supervisor.proto.Supervisor/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *supervisorClient) UpdateSystem(ctx context.Context, in *UpdateSystemReq, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/supervisor.proto.Supervisor/UpdateSystem", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *supervisorClient) GetSystemInfo(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetSystemInfoResp, error) {
	out := new(GetSystemInfoResp)
	err := c.cc.Invoke(ctx, "/supervisor.proto.Supervisor/GetSystemInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *supervisorClient) GetExtensions(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetExtensionsResp, error) {
	out := new(GetExtensionsResp)
	err := c.cc.Invoke(ctx, "/supervisor.proto.Supervisor/GetExtensions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *supervisorClient) RemoteHelp(ctx context.Context, in *RemoteHelpReq, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/supervisor.proto.Supervisor/RemoteHelp", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *supervisorClient) RemoteHelpEnabled(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*RemoteHelpEnabledResp, error) {
	out := new(RemoteHelpEnabledResp)
	err := c.cc.Invoke(ctx, "/supervisor.proto.Supervisor/RemoteHelpEnabled", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SupervisorServer is the server API for Supervisor service.
// All implementations must embed UnimplementedSupervisorServer
// for forward compatibility
type SupervisorServer interface {
	Restart(context.Context, *RestartReq) (*Response, error)
	Backup(context.Context, *BackupReq) (*Response, error)
	Restore(context.Context, *RestoreReq) (*Response, error)
	Update(context.Context, *UpdateReq) (*Response, error)
	UpdateSystem(context.Context, *UpdateSystemReq) (*Response, error)
	GetSystemInfo(context.Context, *emptypb.Empty) (*GetSystemInfoResp, error)
	GetExtensions(context.Context, *emptypb.Empty) (*GetExtensionsResp, error)
	RemoteHelp(context.Context, *RemoteHelpReq) (*Response, error)
	RemoteHelpEnabled(context.Context, *emptypb.Empty) (*RemoteHelpEnabledResp, error)
	mustEmbedUnimplementedSupervisorServer()
}

// UnimplementedSupervisorServer must be embedded to have forward compatible implementations.
type UnimplementedSupervisorServer struct {
}

func (UnimplementedSupervisorServer) Restart(context.Context, *RestartReq) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Restart not implemented")
}
func (UnimplementedSupervisorServer) Backup(context.Context, *BackupReq) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Backup not implemented")
}
func (UnimplementedSupervisorServer) Restore(context.Context, *RestoreReq) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Restore not implemented")
}
func (UnimplementedSupervisorServer) Update(context.Context, *UpdateReq) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (UnimplementedSupervisorServer) UpdateSystem(context.Context, *UpdateSystemReq) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateSystem not implemented")
}
func (UnimplementedSupervisorServer) GetSystemInfo(context.Context, *emptypb.Empty) (*GetSystemInfoResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSystemInfo not implemented")
}
func (UnimplementedSupervisorServer) GetExtensions(context.Context, *emptypb.Empty) (*GetExtensionsResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetExtensions not implemented")
}
func (UnimplementedSupervisorServer) RemoteHelp(context.Context, *RemoteHelpReq) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoteHelp not implemented")
}
func (UnimplementedSupervisorServer) RemoteHelpEnabled(context.Context, *emptypb.Empty) (*RemoteHelpEnabledResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoteHelpEnabled not implemented")
}
func (UnimplementedSupervisorServer) mustEmbedUnimplementedSupervisorServer() {}

// UnsafeSupervisorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SupervisorServer will
// result in compilation errors.
type UnsafeSupervisorServer interface {
	mustEmbedUnimplementedSupervisorServer()
}

func RegisterSupervisorServer(s grpc.ServiceRegistrar, srv SupervisorServer) {
	s.RegisterService(&Supervisor_ServiceDesc, srv)
}

func _Supervisor_Restart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RestartReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).Restart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.proto.Supervisor/Restart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).Restart(ctx, req.(*RestartReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Supervisor_Backup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BackupReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).Backup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.proto.Supervisor/Backup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).Backup(ctx, req.(*BackupReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Supervisor_Restore_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RestoreReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).Restore(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.proto.Supervisor/Restore",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).Restore(ctx, req.(*RestoreReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Supervisor_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.proto.Supervisor/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).Update(ctx, req.(*UpdateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Supervisor_UpdateSystem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateSystemReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).UpdateSystem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.proto.Supervisor/UpdateSystem",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).UpdateSystem(ctx, req.(*UpdateSystemReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Supervisor_GetSystemInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).GetSystemInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.proto.Supervisor/GetSystemInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).GetSystemInfo(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Supervisor_GetExtensions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).GetExtensions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.proto.Supervisor/GetExtensions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).GetExtensions(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Supervisor_RemoteHelp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoteHelpReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).RemoteHelp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.proto.Supervisor/RemoteHelp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).RemoteHelp(ctx, req.(*RemoteHelpReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Supervisor_RemoteHelpEnabled_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SupervisorServer).RemoteHelpEnabled(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/supervisor.proto.Supervisor/RemoteHelpEnabled",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SupervisorServer).RemoteHelpEnabled(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Supervisor_ServiceDesc is the grpc.ServiceDesc for Supervisor service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Supervisor_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "supervisor.proto.Supervisor",
	HandlerType: (*SupervisorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Restart",
			Handler:    _Supervisor_Restart_Handler,
		},
		{
			MethodName: "Backup",
			Handler:    _Supervisor_Backup_Handler,
		},
		{
			MethodName: "Restore",
			Handler:    _Supervisor_Restore_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _Supervisor_Update_Handler,
		},
		{
			MethodName: "UpdateSystem",
			Handler:    _Supervisor_UpdateSystem_Handler,
		},
		{
			MethodName: "GetSystemInfo",
			Handler:    _Supervisor_GetSystemInfo_Handler,
		},
		{
			MethodName: "GetExtensions",
			Handler:    _Supervisor_GetExtensions_Handler,
		},
		{
			MethodName: "RemoteHelp",
			Handler:    _Supervisor_RemoteHelp_Handler,
		},
		{
			MethodName: "RemoteHelpEnabled",
			Handler:    _Supervisor_RemoteHelpEnabled_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "supervisor.proto",
}
