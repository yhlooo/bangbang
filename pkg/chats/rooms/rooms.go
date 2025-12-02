package rooms

import (
	"context"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
	"github.com/yhlooo/bangbang/pkg/chats/channels"
)

// Room 聊天房间
type Room interface {
	// Info 获取房间信息
	Info(ctx context.Context) (*RoomInfo, error)

	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, msg *chatv1.Message) error
	// Listen 获取监听消息的信道
	Listen(ctx context.Context, user *metav1.ObjectMeta) (channels.Channel, error)

	// Close 关闭
	Close(ctx context.Context) error
}

// RoomWithUpstream 有上游的房间
type RoomWithUpstream interface {
	Room
	// Upstream 返回当前房间的上游
	Upstream() Room
	// SetUpstream 设置上游房间
	SetUpstream(ctx context.Context, room Room) error
}

// RoomInfo 房间信息
type RoomInfo struct {
	UID                   metav1.UID
	OwnerUID              metav1.UID
	OwnerName             string
	PublishedKeySignature string
}
