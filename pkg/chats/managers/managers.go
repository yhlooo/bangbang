package managers

import (
	"context"

	"github.com/yhlooo/bangbang/pkg/chats/rooms"
)

// Manager 聊天管理器
type Manager interface {
	// 房主方法

	// CreateLocalRoom 创建房间
	CreateLocalRoom(ctx context.Context) (rooms.Room, error)
	// GetLocalRoom 获取房间
	GetLocalRoom(ctx context.Context, uid string) (rooms.Room, error)
	// DeleteLocalRoom 删除房间
	DeleteLocalRoom(ctx context.Context, uid string) error

	// 客人方法
}

// DefaultRoomID 默认房间 ID
const DefaultRoomID = "default"
