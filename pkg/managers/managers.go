package managers

import (
	"context"

	"github.com/yhlooo/bangbang/pkg/chats/rooms"
)

// Manager 聊天管理器
type Manager interface {
	// SelfRoom 获取自己主持的房间
	SelfRoom(ctx context.Context) rooms.Room
	// StartServer 开始运行 HTTP 服务
	StartServer(ctx context.Context) (<-chan struct{}, error)
	// StartTransponder 开始运行应答机
	StartTransponder(ctx context.Context) error
	// StartSearchUpstream 开始搜索上游
	StartSearchUpstream(ctx context.Context) error
}
