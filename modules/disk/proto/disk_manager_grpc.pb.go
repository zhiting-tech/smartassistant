// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.1.0
// - protoc             v3.19.0
// source: disk_manager.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DiskManagerClient is the client API for DiskManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DiskManagerClient interface {
	// 物理分区列表
	PhysicalVolumeList(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*PhysicalVolumeListResp, error)
	// 储存池，以及储存池下的物理分区，逻辑分区
	VolumeGroupList(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*VolumeGroupListResp, error)
	// 选择物理分区，创建存储池
	VolumeGroupCreate(ctx context.Context, in *VolumeGroupCreateOrExtendReq, opts ...grpc.CallOption) (*VolumeGroupResp, error)
	// 添加物理分区到存储池
	VolumeGroupExtend(ctx context.Context, in *VolumeGroupCreateOrExtendReq, opts ...grpc.CallOption) (*VolumeGroupResp, error)
	// 重命名存储池
	VolumeGroupRename(ctx context.Context, in *VolumeGroupRenameReq, opts ...grpc.CallOption) (*VolumeGroupResp, error)
	// 删除存储池
	VolumeGroupRemove(ctx context.Context, in *VolumeGroupRemoveReq, opts ...grpc.CallOption) (*Empty, error)
	// 创建逻辑分区
	LogicalVolumeCreate(ctx context.Context, in *LogicalVolumeCreateReq, opts ...grpc.CallOption) (*LogicalVolumeResp, error)
	// 修改逻辑分区名称
	LogicalVolumeRename(ctx context.Context, in *LogicalVolumeRenameReq, opts ...grpc.CallOption) (*LogicalVolumeResp, error)
	// 增大逻辑分区大小；请使用足够长的超时时间
	LogicalVolumeExtend(ctx context.Context, in *LogicalVolumeExtendReq, opts ...grpc.CallOption) (*LogicalVolumeResp, error)
	// 删除逻辑分区
	LogicalVolumeRemove(ctx context.Context, in *LogicalVolumeRemoveReq, opts ...grpc.CallOption) (*Empty, error)
	// 外部已被挂载物理分区列表
	PhysicalMountedList(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*PhysicalVolumeListResp, error)
	// 挂载外部硬盘
	MountPhysical(ctx context.Context, in *MountPhysicalReq, opts ...grpc.CallOption) (*Empty, error)
	// 弹出挂载硬盘
	UnmountPhysical(ctx context.Context, in *MountPhysicalReq, opts ...grpc.CallOption) (*Empty, error)
}

type diskManagerClient struct {
	cc grpc.ClientConnInterface
}

func NewDiskManagerClient(cc grpc.ClientConnInterface) DiskManagerClient {
	return &diskManagerClient{cc}
}

func (c *diskManagerClient) PhysicalVolumeList(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*PhysicalVolumeListResp, error) {
	out := new(PhysicalVolumeListResp)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/PhysicalVolumeList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) VolumeGroupList(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*VolumeGroupListResp, error) {
	out := new(VolumeGroupListResp)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/VolumeGroupList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) VolumeGroupCreate(ctx context.Context, in *VolumeGroupCreateOrExtendReq, opts ...grpc.CallOption) (*VolumeGroupResp, error) {
	out := new(VolumeGroupResp)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/VolumeGroupCreate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) VolumeGroupExtend(ctx context.Context, in *VolumeGroupCreateOrExtendReq, opts ...grpc.CallOption) (*VolumeGroupResp, error) {
	out := new(VolumeGroupResp)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/VolumeGroupExtend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) VolumeGroupRename(ctx context.Context, in *VolumeGroupRenameReq, opts ...grpc.CallOption) (*VolumeGroupResp, error) {
	out := new(VolumeGroupResp)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/VolumeGroupRename", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) VolumeGroupRemove(ctx context.Context, in *VolumeGroupRemoveReq, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/VolumeGroupRemove", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) LogicalVolumeCreate(ctx context.Context, in *LogicalVolumeCreateReq, opts ...grpc.CallOption) (*LogicalVolumeResp, error) {
	out := new(LogicalVolumeResp)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/LogicalVolumeCreate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) LogicalVolumeRename(ctx context.Context, in *LogicalVolumeRenameReq, opts ...grpc.CallOption) (*LogicalVolumeResp, error) {
	out := new(LogicalVolumeResp)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/LogicalVolumeRename", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) LogicalVolumeExtend(ctx context.Context, in *LogicalVolumeExtendReq, opts ...grpc.CallOption) (*LogicalVolumeResp, error) {
	out := new(LogicalVolumeResp)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/LogicalVolumeExtend", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) LogicalVolumeRemove(ctx context.Context, in *LogicalVolumeRemoveReq, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/LogicalVolumeRemove", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) PhysicalMountedList(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*PhysicalVolumeListResp, error) {
	out := new(PhysicalVolumeListResp)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/PhysicalMountedList", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) MountPhysical(ctx context.Context, in *MountPhysicalReq, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/MountPhysical", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *diskManagerClient) UnmountPhysical(ctx context.Context, in *MountPhysicalReq, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/proto.DiskManager/UnmountPhysical", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DiskManagerServer is the server API for DiskManager service.
// All implementations must embed UnimplementedDiskManagerServer
// for forward compatibility
type DiskManagerServer interface {
	// 物理分区列表
	PhysicalVolumeList(context.Context, *Empty) (*PhysicalVolumeListResp, error)
	// 储存池，以及储存池下的物理分区，逻辑分区
	VolumeGroupList(context.Context, *Empty) (*VolumeGroupListResp, error)
	// 选择物理分区，创建存储池
	VolumeGroupCreate(context.Context, *VolumeGroupCreateOrExtendReq) (*VolumeGroupResp, error)
	// 添加物理分区到存储池
	VolumeGroupExtend(context.Context, *VolumeGroupCreateOrExtendReq) (*VolumeGroupResp, error)
	// 重命名存储池
	VolumeGroupRename(context.Context, *VolumeGroupRenameReq) (*VolumeGroupResp, error)
	// 删除存储池
	VolumeGroupRemove(context.Context, *VolumeGroupRemoveReq) (*Empty, error)
	// 创建逻辑分区
	LogicalVolumeCreate(context.Context, *LogicalVolumeCreateReq) (*LogicalVolumeResp, error)
	// 修改逻辑分区名称
	LogicalVolumeRename(context.Context, *LogicalVolumeRenameReq) (*LogicalVolumeResp, error)
	// 增大逻辑分区大小；请使用足够长的超时时间
	LogicalVolumeExtend(context.Context, *LogicalVolumeExtendReq) (*LogicalVolumeResp, error)
	// 删除逻辑分区
	LogicalVolumeRemove(context.Context, *LogicalVolumeRemoveReq) (*Empty, error)
	// 外部已被挂载物理分区列表
	PhysicalMountedList(context.Context, *Empty) (*PhysicalVolumeListResp, error)
	// 挂载外部硬盘
	MountPhysical(context.Context, *MountPhysicalReq) (*Empty, error)
	// 弹出挂载硬盘
	UnmountPhysical(context.Context, *MountPhysicalReq) (*Empty, error)
	mustEmbedUnimplementedDiskManagerServer()
}

// UnimplementedDiskManagerServer must be embedded to have forward compatible implementations.
type UnimplementedDiskManagerServer struct {
}

func (UnimplementedDiskManagerServer) PhysicalVolumeList(context.Context, *Empty) (*PhysicalVolumeListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PhysicalVolumeList not implemented")
}
func (UnimplementedDiskManagerServer) VolumeGroupList(context.Context, *Empty) (*VolumeGroupListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VolumeGroupList not implemented")
}
func (UnimplementedDiskManagerServer) VolumeGroupCreate(context.Context, *VolumeGroupCreateOrExtendReq) (*VolumeGroupResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VolumeGroupCreate not implemented")
}
func (UnimplementedDiskManagerServer) VolumeGroupExtend(context.Context, *VolumeGroupCreateOrExtendReq) (*VolumeGroupResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VolumeGroupExtend not implemented")
}
func (UnimplementedDiskManagerServer) VolumeGroupRename(context.Context, *VolumeGroupRenameReq) (*VolumeGroupResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VolumeGroupRename not implemented")
}
func (UnimplementedDiskManagerServer) VolumeGroupRemove(context.Context, *VolumeGroupRemoveReq) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VolumeGroupRemove not implemented")
}
func (UnimplementedDiskManagerServer) LogicalVolumeCreate(context.Context, *LogicalVolumeCreateReq) (*LogicalVolumeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LogicalVolumeCreate not implemented")
}
func (UnimplementedDiskManagerServer) LogicalVolumeRename(context.Context, *LogicalVolumeRenameReq) (*LogicalVolumeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LogicalVolumeRename not implemented")
}
func (UnimplementedDiskManagerServer) LogicalVolumeExtend(context.Context, *LogicalVolumeExtendReq) (*LogicalVolumeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LogicalVolumeExtend not implemented")
}
func (UnimplementedDiskManagerServer) LogicalVolumeRemove(context.Context, *LogicalVolumeRemoveReq) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LogicalVolumeRemove not implemented")
}
func (UnimplementedDiskManagerServer) PhysicalMountedList(context.Context, *Empty) (*PhysicalVolumeListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PhysicalMountedList not implemented")
}
func (UnimplementedDiskManagerServer) MountPhysical(context.Context, *MountPhysicalReq) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MountPhysical not implemented")
}
func (UnimplementedDiskManagerServer) UnmountPhysical(context.Context, *MountPhysicalReq) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnmountPhysical not implemented")
}
func (UnimplementedDiskManagerServer) mustEmbedUnimplementedDiskManagerServer() {}

// UnsafeDiskManagerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DiskManagerServer will
// result in compilation errors.
type UnsafeDiskManagerServer interface {
	mustEmbedUnimplementedDiskManagerServer()
}

func RegisterDiskManagerServer(s grpc.ServiceRegistrar, srv DiskManagerServer) {
	s.RegisterService(&DiskManager_ServiceDesc, srv)
}

func _DiskManager_PhysicalVolumeList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).PhysicalVolumeList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/PhysicalVolumeList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).PhysicalVolumeList(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_VolumeGroupList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).VolumeGroupList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/VolumeGroupList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).VolumeGroupList(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_VolumeGroupCreate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VolumeGroupCreateOrExtendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).VolumeGroupCreate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/VolumeGroupCreate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).VolumeGroupCreate(ctx, req.(*VolumeGroupCreateOrExtendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_VolumeGroupExtend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VolumeGroupCreateOrExtendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).VolumeGroupExtend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/VolumeGroupExtend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).VolumeGroupExtend(ctx, req.(*VolumeGroupCreateOrExtendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_VolumeGroupRename_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VolumeGroupRenameReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).VolumeGroupRename(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/VolumeGroupRename",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).VolumeGroupRename(ctx, req.(*VolumeGroupRenameReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_VolumeGroupRemove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VolumeGroupRemoveReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).VolumeGroupRemove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/VolumeGroupRemove",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).VolumeGroupRemove(ctx, req.(*VolumeGroupRemoveReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_LogicalVolumeCreate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogicalVolumeCreateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).LogicalVolumeCreate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/LogicalVolumeCreate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).LogicalVolumeCreate(ctx, req.(*LogicalVolumeCreateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_LogicalVolumeRename_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogicalVolumeRenameReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).LogicalVolumeRename(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/LogicalVolumeRename",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).LogicalVolumeRename(ctx, req.(*LogicalVolumeRenameReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_LogicalVolumeExtend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogicalVolumeExtendReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).LogicalVolumeExtend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/LogicalVolumeExtend",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).LogicalVolumeExtend(ctx, req.(*LogicalVolumeExtendReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_LogicalVolumeRemove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogicalVolumeRemoveReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).LogicalVolumeRemove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/LogicalVolumeRemove",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).LogicalVolumeRemove(ctx, req.(*LogicalVolumeRemoveReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_PhysicalMountedList_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).PhysicalMountedList(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/PhysicalMountedList",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).PhysicalMountedList(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_MountPhysical_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MountPhysicalReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).MountPhysical(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/MountPhysical",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).MountPhysical(ctx, req.(*MountPhysicalReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _DiskManager_UnmountPhysical_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MountPhysicalReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DiskManagerServer).UnmountPhysical(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.DiskManager/UnmountPhysical",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DiskManagerServer).UnmountPhysical(ctx, req.(*MountPhysicalReq))
	}
	return interceptor(ctx, in, info, handler)
}

// DiskManager_ServiceDesc is the grpc.ServiceDesc for DiskManager service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DiskManager_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.DiskManager",
	HandlerType: (*DiskManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PhysicalVolumeList",
			Handler:    _DiskManager_PhysicalVolumeList_Handler,
		},
		{
			MethodName: "VolumeGroupList",
			Handler:    _DiskManager_VolumeGroupList_Handler,
		},
		{
			MethodName: "VolumeGroupCreate",
			Handler:    _DiskManager_VolumeGroupCreate_Handler,
		},
		{
			MethodName: "VolumeGroupExtend",
			Handler:    _DiskManager_VolumeGroupExtend_Handler,
		},
		{
			MethodName: "VolumeGroupRename",
			Handler:    _DiskManager_VolumeGroupRename_Handler,
		},
		{
			MethodName: "VolumeGroupRemove",
			Handler:    _DiskManager_VolumeGroupRemove_Handler,
		},
		{
			MethodName: "LogicalVolumeCreate",
			Handler:    _DiskManager_LogicalVolumeCreate_Handler,
		},
		{
			MethodName: "LogicalVolumeRename",
			Handler:    _DiskManager_LogicalVolumeRename_Handler,
		},
		{
			MethodName: "LogicalVolumeExtend",
			Handler:    _DiskManager_LogicalVolumeExtend_Handler,
		},
		{
			MethodName: "LogicalVolumeRemove",
			Handler:    _DiskManager_LogicalVolumeRemove_Handler,
		},
		{
			MethodName: "PhysicalMountedList",
			Handler:    _DiskManager_PhysicalMountedList_Handler,
		},
		{
			MethodName: "MountPhysical",
			Handler:    _DiskManager_MountPhysical_Handler,
		},
		{
			MethodName: "UnmountPhysical",
			Handler:    _DiskManager_UnmountPhysical_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "disk_manager.proto",
}