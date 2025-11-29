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
	"github.com/yhlooo/bangbang/pkg/chats/rooms"
	"github.com/yhlooo/bangbang/pkg/servers/common"
)

// Server 聊天服务
type Server interface {
	// GetInfo 获取房间信息
	GetInfo(ctx context.Context, req *EmptyRequest) (*chatv1.Room, error)
	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, req *CreateMessageRequest) (*chatv1.Message, error)
	// ListenMessages 监听消息
	ListenMessages(ctx context.Context, _ *EmptyRequest) (*metav1.Status, error)
}

type EmptyRequest struct{}

// CreateMessageRequest 创建消息请求
type CreateMessageRequest struct {
	Message chatv1.Message
}

// Body 返回 body 部分字段
func (req *CreateMessageRequest) Body() interface{} {
	return &req.Message
}

// NewServer 创建 Server
func NewServer(room rooms.Room) Server {
	return &chatServer{
		room: room,
	}
}

// chatServer Server 的默认实现
type chatServer struct {
	room rooms.Room
}

// GetInfo 获取房间信息
func (s *chatServer) GetInfo(ctx context.Context, _ *EmptyRequest) (*chatv1.Room, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("get room info")

	info, err := s.room.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("get room info error: %w", err)
	}
	return &chatv1.Room{
		APIMeta: metav1.NewAPIMeta(chatv1.KindRoom),
		Meta:    metav1.ObjectMeta{UID: info.UID},
		Owner: chatv1.User{
			APIMeta: metav1.NewAPIMeta(chatv1.KindUser),
			Meta:    metav1.ObjectMeta{UID: info.OwnerUID},
		},
		KeySignature: info.PublishedKeySignature,
	}, nil
}

// CreateMessage 创建消息
func (s *chatServer) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*chatv1.Message, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("create message in room")

	if err := s.room.CreateMessage(ctx, &req.Message); err != nil {
		return nil, fmt.Errorf("create message in room error: %w", err)
	}

	return &req.Message, nil
}

// ListenMessages 监听消息
func (s *chatServer) ListenMessages(ctx context.Context, _ *EmptyRequest) (*metav1.Status, error) {
	logger := logr.FromContextOrDiscard(ctx)
	logger.Info("listen messages in room")

	ch, stop, err := s.room.Listen(ctx)
	if err != nil {
		return nil, fmt.Errorf("listen message in room error: %w", err)
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
