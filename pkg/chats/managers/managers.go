package managers

import (
	"context"

	"github.com/yhlooo/bangbang/pkg/chats/rooms"
)

// Manager 聊天管理器
type Manager interface {
	// 房主方法

	// SelfRoom 获取自己主持的房间
	SelfRoom(ctx context.Context) rooms.Room

	// 客人方法
}
