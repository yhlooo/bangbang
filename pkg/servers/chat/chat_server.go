package chat

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
)

// Server 聊天服务
type Server interface {
	// GetRoom 获取房间信息
	GetRoom(ctx context.Context, req *GetRoomRequest) (*chatv1.Room, error)
	// CreateChannel 创建信道（加入房间）
	CreateChannel(ctx context.Context, req *CreateChannelRequest) (*chatv1.Channel, error)
	// ListMembers 列出成员
	ListMembers(ctx context.Context, req *ListMembersRequest) (*chatv1.UserList, error)
	// GetMember 获取成员信息
	GetMember(ctx context.Context, req *GetMemberRequest) (*chatv1.User, error)
}

// GetRoomRequest 获取房间信息请求
type GetRoomRequest struct {
	RoomUID string `uri:"roomUID" binding:"required"`
}

// CreateChannelRequest 创建信道请求
type CreateChannelRequest struct {
	RoomUID string `uri:"roomUID" binding:"required"`
	Channel chatv1.Channel
}

// Body 返回 body 的指针
func (req *CreateChannelRequest) Body() interface{} {
	return &req.Channel
}

// ListMembersRequest 列出成员请求
type ListMembersRequest struct {
	RoomUID string `uri:"roomUID" binding:"required"`
}

// GetMemberRequest 获取成员信息请求
type GetMemberRequest struct {
	RoomUID   string `uri:"roomUID" binding:"required"`
	MemberUID string `uri:"memberUID" binding:"required"`
}

// NewServer 创建 Server
func NewServer() Server {
	return &chatServer{}
}

// chatServer Server 的默认实现
type chatServer struct{}

// GetRoom 获取房间信息
func (s *chatServer) GetRoom(ctx context.Context, req *GetRoomRequest) (*chatv1.Room, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info(fmt.Sprintf("get room: %s", req.RoomUID))
	return &chatv1.Room{
		APIMeta: metav1.APIMeta{
			Version: "v1",
		},
		Meta: metav1.ObjectMeta{
			UID: req.RoomUID,
		},
		Owner: chatv1.User{
			APIMeta: metav1.APIMeta{
				Version: "v1",
			},
			Meta: metav1.ObjectMeta{
				UID:  "123",
				Name: "hhh",
			},
		},
	}, nil
}

// CreateChannel 创建信道（加入房间）
func (s *chatServer) CreateChannel(ctx context.Context, req *CreateChannelRequest) (*chatv1.Channel, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info(fmt.Sprintf("create channel: room: %s", req.RoomUID))
	return &req.Channel, nil
}

// ListMembers 列出成员
func (s *chatServer) ListMembers(ctx context.Context, req *ListMembersRequest) (*chatv1.UserList, error) {
	panic("implement me")
}

// GetMember 获取成员信息
func (s *chatServer) GetMember(ctx context.Context, req *GetMemberRequest) (*chatv1.User, error) {
	panic("implement me")
}
