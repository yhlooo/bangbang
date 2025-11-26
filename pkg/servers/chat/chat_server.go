package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
	"github.com/yhlooo/bangbang/pkg/chats/managers"
	"github.com/yhlooo/bangbang/pkg/servers/common"
)

// Server 聊天服务
type Server interface {
	// GetRoom 获取房间信息
	GetRoom(ctx context.Context, req *GetRoomRequest) (*chatv1.Room, error)
	// CreateRoomMember 创建房间成员
	CreateRoomMember(ctx context.Context, req *CreateMemberRequest) (*chatv1.User, error)
	// DeleteRoomMember 删除房间成员
	DeleteRoomMember(ctx context.Context, req *DeleteMemberRequest) (*metav1.Status, error)
	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, req *CreateMessageRequest) (*chatv1.Message, error)
	// ListenMessages 监听消息
	ListenMessages(ctx context.Context, req *GetRoomRequest) (*metav1.Status, error)
	// ListMembers 列出成员
	ListMembers(ctx context.Context, req *GetRoomRequest) (*chatv1.UserList, error)
}

// GetRoomRequest 获取房间请求
type GetRoomRequest struct {
	RoomUID string `uri:"roomUID" binding:"required"`
}

// CreateMessageRequest 创建消息请求
type CreateMessageRequest struct {
	RoomUID string `uri:"roomUID" binding:"required"`
	Message chatv1.Message
}

// Body 返回 body 部分字段
func (req *CreateMessageRequest) Body() interface{} {
	return &req.Message
}

// CreateMemberRequest 创建成员请求
type CreateMemberRequest struct {
	RoomUID string `uri:"roomUID" binding:"required"`
	User    chatv1.User
}

// Body 返回 body 部分字段
func (req *CreateMemberRequest) Body() interface{} {
	return &req.User
}

// DeleteMemberRequest 删除成员请求
type DeleteMemberRequest struct {
	RoomUID   string `uri:"roomUID" binding:"required"`
	MemberUID string `uri:"memberUID" binding:"required"`
}

// NewServer 创建 Server
func NewServer(mgr managers.Manager) Server {
	return &chatServer{
		mgr: mgr,
	}
}

// chatServer Server 的默认实现
type chatServer struct {
	mgr managers.Manager
}

// GetRoom 获取房间信息
func (s *chatServer) GetRoom(ctx context.Context, req *GetRoomRequest) (*chatv1.Room, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info(fmt.Sprintf("get room: %s", req.RoomUID))

	room, err := s.mgr.GetRoom(ctx, req.RoomUID)
	if err != nil {
		return nil, fmt.Errorf("get room %q error: %w", req.RoomUID, err)
	}

	info, err := room.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("get room %q info error: %w", req.RoomUID, err)
	}

	return &chatv1.Room{
		APIMeta: metav1.NewAPIMeta(),
		Meta: metav1.ObjectMeta{
			UID: req.RoomUID,
		},
		Owner: chatv1.User{
			APIMeta: metav1.NewAPIMeta(),
			Meta: metav1.ObjectMeta{
				UID: info.Owner,
			},
		},
	}, nil
}

// CreateRoomMember 创建房间成员
func (s *chatServer) CreateRoomMember(ctx context.Context, req *CreateMemberRequest) (*chatv1.User, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info(fmt.Sprintf("create member in room: %s", req.RoomUID))

	room, err := s.mgr.GetRoom(ctx, req.RoomUID)
	if err != nil {
		return nil, fmt.Errorf("get room %q error: %w", req.RoomUID, err)
	}

	if err := room.Join(ctx, req.User.Meta.UID); err != nil {
		return nil, fmt.Errorf("join to room %q error: %w", req.RoomUID, err)
	}

	return &req.User, nil
}

// ListMembers 列出成员
func (s *chatServer) ListMembers(ctx context.Context, req *GetRoomRequest) (*chatv1.UserList, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info(fmt.Sprintf("list room members: %s", req.RoomUID))

	room, err := s.mgr.GetRoom(ctx, req.RoomUID)
	if err != nil {
		return nil, fmt.Errorf("get room %q error: %w", req.RoomUID, err)
	}

	info, err := room.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("get room %q info error: %w", req.RoomUID, err)
	}

	var items []chatv1.User
	if len(info.Members) > 0 {
		items = make([]chatv1.User, 0, len(info.Members))
		for _, member := range info.Members {
			items = append(items, chatv1.User{
				APIMeta: metav1.NewAPIMeta(),
				Meta: metav1.ObjectMeta{
					UID: member,
				},
			})
		}
	}

	return &chatv1.UserList{
		APIMeta: metav1.NewAPIMeta(),
		Items:   items,
	}, nil
}

// DeleteRoomMember 删除房间成员
func (s *chatServer) DeleteRoomMember(ctx context.Context, req *DeleteMemberRequest) (*metav1.Status, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info(fmt.Sprintf("delete member in room %q: %s", req.RoomUID, req.MemberUID))

	room, err := s.mgr.GetRoom(ctx, req.RoomUID)
	if err != nil {
		return nil, fmt.Errorf("get room %q error: %w", req.RoomUID, err)
	}

	if err := room.Leave(ctx, req.MemberUID); err != nil {
		return nil, fmt.Errorf("leave room %q error: %w", req.RoomUID, err)
	}

	return common.NewOkStatus(ctx), nil
}

// CreateMessage 创建消息
func (s *chatServer) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*chatv1.Message, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info(fmt.Sprintf("create message in room : %s", req.RoomUID))

	room, err := s.mgr.GetRoom(ctx, req.RoomUID)
	if err != nil {
		return nil, fmt.Errorf("get room %q error: %w", req.RoomUID, err)
	}

	if err := room.CreateMessage(ctx, &req.Message); err != nil {
		return nil, fmt.Errorf("create message in room %q error: %w", req.RoomUID, err)
	}

	return &req.Message, nil
}

// ListenMessages 监听消息
func (s *chatServer) ListenMessages(ctx context.Context, req *GetRoomRequest) (*metav1.Status, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info(fmt.Sprintf("listen messages in room: %s", req.RoomUID))

	room, err := s.mgr.GetRoom(ctx, req.RoomUID)
	if err != nil {
		return nil, fmt.Errorf("get room %q error: %w", req.RoomUID, err)
	}

	ch, stop, err := room.Listen(ctx)
	if err != nil {
		return nil, fmt.Errorf("listen message in room %q error: %w", req.RoomUID, err)
	}
	defer stop()

	ginCTX, ok := ctx.(*gin.Context)
	if !ok {
		return nil, fmt.Errorf("require *gin.Context")
	}

	// 写响应头
	logger.Info("start listening messages ...")
	ginCTX.Header("Transfer-Encoding", "chunked")
	ginCTX.Status(http.StatusOK)
	ginCTX.Writer.Flush()

	// 流式传输消息
	for msg := range ch {
		logger.Info(fmt.Sprintf("send message %q to client", msg.Meta.UID))
		raw, err := json.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("marshal message %q to json error: %w", msg.Meta.UID, err)
		}
		_, err = fmt.Fprintln(ginCTX.Writer, string(raw))
		if err != nil {
			return nil, fmt.Errorf("write message %q to response error: %w", msg.Meta.UID, err)
		}
		ginCTX.Writer.Flush()
	}

	return common.NewOkStatus(ctx), nil
}
