package managers

import (
	"context"

	"github.com/yhlooo/bangbang/pkg/chats/rooms"
)

// NewManager 创建聊天管理器
func NewManager() Manager {
	mgr := &defaultManager{
		selfRoom: rooms.NewLocalRoom(),
	}
	return mgr
}

// defaultManager 是 Manager 的默认实现
type defaultManager struct {
	selfRoom rooms.Room
}

var _ Manager = (*defaultManager)(nil)

// SelfRoom 获取自己主持的房间
func (m *defaultManager) SelfRoom(_ context.Context) rooms.Room {
	return m.selfRoom
}
