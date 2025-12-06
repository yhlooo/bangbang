package discovery

import (
	"context"
	"time"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
	"github.com/yhlooo/bangbang/pkg/signatures"
)

// Discoverer 发现器

type Discoverer interface {
	Search(ctx context.Context, key signatures.Key, opts SearchOptions) ([]Room, error)
}

// SearchOptions 搜索选项
type SearchOptions struct {
	// 搜索时长
	Duration time.Duration
	// 请求间隔
	RequestInterval time.Duration
	// 检查可用性
	CheckAvailability bool
}

// Room 房间
type Room struct {
	// 房间信息
	Info chatv1.Room
	// 可用的访问端点
	AvailableEndpoint string
}

// Transponder 应答机
type Transponder interface {
	// Start 开始运行应答机
	Start(ctx context.Context) error
}
