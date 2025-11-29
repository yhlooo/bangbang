package rooms

import (
	"context"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
)

// Room 聊天房间
type Room interface {
	// Info 获取房间信息
	Info(ctx context.Context) (*RoomInfo, error)

	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, msg *chatv1.Message) error
	// Listen 获取监听消息的信道
	Listen(ctx context.Context) (ch <-chan *chatv1.Message, stop func(), err error)

	// Close 关闭
	Close(ctx context.Context) error
}

// RoomInfo 房间信息
type RoomInfo struct {
	UID string
}
