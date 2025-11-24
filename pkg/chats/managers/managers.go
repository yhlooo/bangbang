package managers

import (
	"context"

	"github.com/yhlooo/bangbang/pkg/chats/rooms"
)

// Manager 聊天管理器
type Manager interface {
	// 房主方法

	// CreateRoom 创建房间
	CreateRoom(ctx context.Context) (rooms.Room, error)
	// GetRoom 获取房间
	GetRoom(ctx context.Context, uid string) (rooms.Room, error)
	// DeleteRoom 删除房间
	DeleteRoom(ctx context.Context, uid string) error

	// 客人方法
}
