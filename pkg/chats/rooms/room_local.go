package rooms

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
	"github.com/yhlooo/bangbang/pkg/chats/keys"
	"github.com/yhlooo/bangbang/pkg/deduplicators"
)

// NewLocalRoom 创建本地房间实例
func NewLocalRoom(key keys.HashKey, ownerUID metav1.UID) RoomWithUpstream {
	return &localRoom{
		uid:          metav1.NewUID(),
		ownerUID:     ownerUID,
		key:          key.Copy(),
		deduplicator: deduplicators.NewBloomFilter(500, 0.001),
	}
}

// localRoom 是 Room 的本地实现
type localRoom struct {
	uid      metav1.UID
	ownerUID metav1.UID
	key      keys.HashKey

	lock sync.RWMutex

	closed               bool
	channels             map[chan *chatv1.Message]struct{}
	upstream             Room
	upstreamDeduplicator deduplicators.Deduplicator
	deduplicator         deduplicators.Deduplicator
}

var _ RoomWithUpstream = (*localRoom)(nil)

// Info 获取房间信息
func (r *localRoom) Info(_ context.Context) (*RoomInfo, error) {
	return &RoomInfo{
		UID:                   r.uid,
		OwnerUID:              r.ownerUID,
		PublishedKeySignature: r.key.PublishedSignature(),
	}, nil
}

// CreateMessage 创建消息
func (r *localRoom) CreateMessage(ctx context.Context, msg *chatv1.Message) error {
	logger := logr.FromContextOrDiscard(ctx)

	r.lock.RLock()
	defer r.lock.RUnlock()

	if r.closed {
		return fmt.Errorf("room already closed")
	}

	if msg.Meta.UID.IsNil() {
		msg.Meta.UID = metav1.NewUID()
	}

	// 去重
	if r.deduplicator.Duplicate(msg.Meta.UID[:]) {
		logger.V(1).Info(fmt.Sprintf("duplicated message: %s", msg.Meta.UID))
		return nil
	}

	for ch := range r.channels {
		select {
		case ch <- msg:
		default:
		}
	}

	// 转发给上游
	if r.upstream != nil && !r.upstreamDeduplicator.Duplicate(msg.Meta.UID[:]) {
		if err := r.upstream.CreateMessage(ctx, msg); err != nil {
			return fmt.Errorf("forward to upstream error: %w", err)
		}
	}

	return nil
}

// Listen 获取监听消息的信道
func (r *localRoom) Listen(_ context.Context) (ch <-chan *chatv1.Message, stop func(), err error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.closed {
		return nil, nil, fmt.Errorf("room already closed")
	}

	if r.channels == nil {
		r.channels = make(map[chan *chatv1.Message]struct{})
	}
	msgCh := make(chan *chatv1.Message, 10)
	r.channels[msgCh] = struct{}{}

	return msgCh, r.stopListen(msgCh), nil
}

// stopListen 停止监听消息
func (r *localRoom) stopListen(ch chan *chatv1.Message) func() {
	return func() {
		r.lock.Lock()
		defer r.lock.Unlock()
		delete(r.channels, ch)
	}
}

// Close 关闭
func (r *localRoom) Close(_ context.Context) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	for ch := range r.channels {
		close(ch)
		delete(r.channels, ch)
	}

	r.closed = true

	return nil
}

// Upstream 返回当前房间的上游
func (r *localRoom) Upstream() Room {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.upstream
}

// SetUpstream 设置上游房间
func (r *localRoom) SetUpstream(ctx context.Context, room Room) error {
	logger := logr.FromContextOrDiscard(ctx)

	r.lock.Lock()
	defer r.lock.Unlock()

	upstream := r.upstream
	if upstream != nil {
		_ = upstream.Close(ctx)
	}

	info, err := room.Info(ctx)
	if err != nil {
		return fmt.Errorf("get upstream room info error: %w", err)
	}

	logger.V(1).Info(fmt.Sprintf("set upstream: %s", info.UID))
	r.upstream = room
	r.upstreamDeduplicator = deduplicators.NewBloomFilter(500, 0.001)
	go r.listenUpstream(ctx)

	return nil
}

// listenUpstream 监听上游房间
func (r *localRoom) listenUpstream(ctx context.Context) {
	logger := logr.FromContextOrDiscard(ctx)

	r.lock.RLock()
	upstream := r.upstream
	upstreamDeduplicator := r.upstreamDeduplicator
	r.lock.RUnlock()

	if upstream == nil {
		return
	}

	defer func() {
		r.lock.Lock()
		defer r.lock.Unlock()
		if r.upstream == upstream {
			r.upstream = nil
		}
		_ = upstream.Close(ctx)
	}()

	ch, stop, err := upstream.Listen(ctx)
	if err != nil {
		logger.Error(err, "listen upstream error")
		return
	}
	defer stop()

	for msg := range ch {
		upstreamDeduplicator.Duplicate(msg.Meta.UID[:])
		if err := r.CreateMessage(ctx, msg); err != nil {
			logger.Error(err, "create message error")
		}
	}
}
