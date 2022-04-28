package extension

import (
	"context"
	"net"

	jsoniter "github.com/json-iterator/go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zhiting-tech/smartassistant/modules/api/auth"
	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/modules/entity"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	pb "github.com/zhiting-tech/smartassistant/pkg/extension/proto"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
)

// ExtensionServer 扩展服务
type ExtensionServer struct {
	NotifyChans map[chan pb.SAEventInfo]struct{}
}

// newGRPCError 返回一个包含sa错误信息的GRPC错误
func newGRPCError(err error) error {
	switch v := err.(type) {
	case errors.Error:
		code := v.Code.Status
		reason := v.Code.Reason
		return status.Error(codes.Code(code), reason)
	default:
		return status.Convert(err).Err()
	}
}

// GetUserInfo 获取用户数据
func (es *ExtensionServer) GetUserInfo(ctx context.Context, req *pb.GetAreaInfoReq) (resp *pb.GetUserInfoResp, err error) {
	resp = new(pb.GetUserInfoResp)

	user, err := auth.GetUserByToken(req.Token)
	if err != nil {
		err = newGRPCError(err)
		return
	}
	area, err := entity.GetAreaByID(user.AreaID)
	if err != nil {
		err = newGRPCError(err)
		return
	}
	departmentInfo, err := entity.GetDepartmentsByUser(user)
	if err != nil {
		err = newGRPCError(err)
		return
	}
	resp.UserInfo = &pb.UserInfo{
		UserId:      int32(user.ID),
		AccountName: user.AccountName,
		NickName:    user.Nickname,
		IsOwner:     entity.IsOwner(user.ID),
	}
	resp.AreaInfo = &pb.Area{
		AreaId:   area.ID,
		Name:     area.Name,
		AreaType: pb.AreaType(area.AreaType),
	}
	for _, d := range departmentInfo {
		info := &pb.DepartmentBaseInfo{
			DepartmentId: int32(d.ID),
			Name:         d.Name,
		}
		if d.IsManager {
			info.CompanyRole = pb.CompanyRole_manager_role
		} else {
			info.CompanyRole = pb.CompanyRole_member_role
		}
		resp.DepartmentInfos = append(resp.DepartmentInfos, info)
	}
	return
}

// GetDepartmentUsers 获取部门下所有成员
func (es *ExtensionServer) GetDepartmentUsers(ctx context.Context, req *pb.GetDepartmentUsersReq) (resp *pb.GetDepartmentUsersResp, err error) {
	resp = &pb.GetDepartmentUsersResp{
		DepartmentUsers: make(map[int32]*pb.DepartmentUsers),
	}
	user, err := auth.GetUserByToken(req.Token)
	if err != nil {
		err = newGRPCError(err)
		return
	}
	_, err = entity.GetAreaByID(user.AreaID)
	if err != nil {
		err = newGRPCError(err)
		return
	}
	owner, err := entity.GetAreaOwner(user.AreaID)
	if err != nil {
		err = newGRPCError(err)
		return
	}

	// 获取请求部门的用户数据
	for _, departmentID := range req.DepartmentIds {
		var (
			department entity.Department
			users      []entity.User
		)
		department, err = entity.GetDepartmentByID(int(departmentID))
		if err != nil {
			err = newGRPCError(err)
			return
		}
		if department.AreaID != user.AreaID {
			err = newGRPCError(errors.New(errors.BadRequest))
			return
		}

		users, err = entity.GetDepartmentUsers(int(departmentID))
		if err != nil {
			err = newGRPCError(err)
			return
		}

		isManager := department.ManagerID != nil && *department.ManagerID == user.ID || owner.ID == user.ID
		userInfo := &pb.DepartmentUsers{}
		for _, u := range users {
			if !isManager && u.ID != user.ID {
				continue
			}
			userInfo.Users = append(userInfo.Users, &pb.UserInfo{
				UserId:      int32(u.ID),
				AccountName: u.AccountName,
				NickName:    u.Nickname,
				IsOwner:     owner.ID == u.ID,
			})
		}

		resp.DepartmentUsers[departmentID] = userInfo
	}
	return
}

// GetBaseUserInfos 通过请求userID获取基础用户数据
func (es *ExtensionServer) GetBaseUserInfos(ctx context.Context, req *pb.BaseUserInfosReq) (resp *pb.DepartmentUsers, err error) {
	resp = new(pb.DepartmentUsers)
	currentUser, err := auth.GetUserByToken(req.Token)
	if err != nil {
		err = newGRPCError(err)
		return
	}
	owner, err := entity.GetAreaOwner(currentUser.AreaID)
	if err != nil {
		err = newGRPCError(err)
		return
	}

	if len(req.UserIds) == 0 {
		var users []entity.User
		if users, err = entity.GetUsers(currentUser.AreaID); err != nil {
			err = newGRPCError(err)
			return
		}
		for _, u := range users {
			resp.Users = append(resp.Users, &pb.UserInfo{
				UserId:      int32(u.ID),
				AccountName: u.AccountName,
				NickName:    u.Nickname,
				IsOwner:     owner.ID == u.ID,
			})
		}
		return
	}
	for _, userID := range req.UserIds {
		var user entity.User
		user, _ = entity.GetUserByID(int(userID))
		if user.AreaID != currentUser.AreaID {
			continue
		}
		resp.Users = append(resp.Users, &pb.UserInfo{
			UserId:      int32(user.ID),
			AccountName: user.AccountName,
			NickName:    user.Nickname,
			IsOwner:     owner.ID == user.ID,
		})
	}
	return
}

// GetDepartments 获取公司下的部门
func (es *ExtensionServer) GetDepartments(ctx context.Context, req *pb.GetAreaInfoReq) (resp *pb.DepartmentsResp, err error) {
	resp = &pb.DepartmentsResp{}
	user, err := auth.GetUserByToken(req.Token)
	if err != nil {
		err = newGRPCError(err)
		return
	}
	_, err = entity.GetAreaByID(user.AreaID)
	if err != nil {
		err = newGRPCError(err)
		return
	}
	var departments []entity.Department
	if departments, err = entity.GetDepartments(user.AreaID); err != nil {
		err = newGRPCError(err)
		return
	}
	for _, d := range departments {
		var users []entity.User
		users, err = entity.GetDepartmentUsers(d.ID)
		if err != nil {
			err = newGRPCError(err)
			return
		}
		resp.Departments = append(resp.Departments, &pb.DepartmentBaseInfo{
			DepartmentId: int32(d.ID),
			Name:         d.Name,
			Sort:         int32(d.Sort),
			UserCount:    int32(len(users)),
		})
	}
	return
}

// SANotifyEvent Sa通知事件
func (es *ExtensionServer) SANotifyEvent(req *pb.EmptyReq, server pb.Extension_SANotifyEventServer) error {
	nc := make(chan pb.SAEventInfo, 20)
	es.Subscribe(nc)
	defer es.Unsubscribe(nc)
	for {
		select {
		case <-server.Context().Done():
			return nil
		case n := <-nc:
			server.Send(&n)
		}
	}
}

// Subscribe 注册通知服务
func (es *ExtensionServer) Subscribe(notify chan pb.SAEventInfo) {
	es.NotifyChans[notify] = struct{}{}
}

// Unsubscribe 解除通知服务
func (es *ExtensionServer) Unsubscribe(notify chan pb.SAEventInfo) {
	delete(es.NotifyChans, notify)
}

// Notify 通知正在监听的服务
func (es *ExtensionServer) Notify(notifyType pb.SAEvent, content map[string]interface{}) {
	data, _ := jsoniter.Marshal(content)
	n := pb.SAEventInfo{
		Event: notifyType,
		Data:  data,
	}
	for ch := range es.NotifyChans {
		select {
		case ch <- n:
		default:
		}
	}
	logger.Infof("extension notify: %d, %v\n", notifyType, content)
}

func (es *ExtensionServer) Run(ctx context.Context) {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	pb.RegisterExtensionServer(server, es)
	lis, err := net.Listen("tcp", config.GetConf().Extension.GRPCAddress())
	if err != nil {
		logger.Error(err)
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(r)
			}
		}()
		err = server.Serve(lis)
		if err != nil {
			logger.Error(err)
			return
		}
	}()
	<-ctx.Done()
	server.Stop()
	logger.Warning("extension server stopped")
}
