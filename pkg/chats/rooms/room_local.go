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
func NewLocalRoom(key keys.HashKey, ownerUID metav1.UID, ownerName string) RoomWithUpstream {
	return &localRoom{
		uid:          metav1.NewUID(),
		ownerUID:     ownerUID,
		ownerName:    ownerName,
		key:          key.Copy(),
		deduplicator: deduplicators.NewBloomFilter(500, 0.001),
	}
}

// localRoom 是 Room 的本地实现
type localRoom struct {
	uid       metav1.UID
	ownerUID  metav1.UID
	ownerName string
	key       keys.HashKey

	lock sync.RWMutex

	closed       bool
	channels     map[chan *chatv1.Message]*metav1.ObjectMeta
	upstream     Room
	deduplicator deduplicators.Deduplicator
}

var _ RoomWithUpstream = (*localRoom)(nil)

// Info 获取房间信息
func (r *localRoom) Info(_ context.Context) (*RoomInfo, error) {
	return &RoomInfo{
		UID:                   r.uid,
		OwnerUID:              r.ownerUID,
		OwnerName:             r.ownerName,
		PublishedKeySignature: r.key.PublishedSignature(),
	}, nil
}

// CreateMessage 创建消息
func (r *localRoom) CreateMessage(ctx context.Context, msg *chatv1.Message) error {
	logger := logr.FromContextOrDiscard(ctx)

	r.lock.RLock()
	defer r.lock.RUnlock()

	if msg.Meta.UID.IsNil() {
		msg.Meta.UID = metav1.NewUID()
	}

	// 去重
	if r.deduplicator.Duplicate(msg.Meta.UID[:]) {
		logger.V(1).Info(fmt.Sprintf("duplicated message: %s", msg.Meta.UID))
		return nil
	}

	if r.closed {
		return fmt.Errorf("room already closed")
	}

	for ch := range r.channels {
		select {
		case ch <- msg:
		default:
		}
	}

	return nil
}

// Listen 获取监听消息的信道
func (r *localRoom) Listen(
	ctx context.Context,
	user *metav1.ObjectMeta,
) (ch <-chan *chatv1.Message, stop func(), err error) {
	logger := logr.FromContextOrDiscard(ctx)

	r.lock.Lock()

	if r.closed {
		r.lock.Unlock()
		return nil, nil, fmt.Errorf("room already closed")
	}

	if r.channels == nil {
		r.channels = make(map[chan *chatv1.Message]*metav1.ObjectMeta)
	}
	msgCh := make(chan *chatv1.Message, 10)
	r.channels[msgCh] = nil
	if user != nil {
		r.channels[msgCh] = &metav1.ObjectMeta{}
		*r.channels[msgCh] = *user
	}
	r.lock.Unlock()

	if user != nil {
		if err := r.CreateMessage(ctx, &chatv1.Message{
			APIMeta: metav1.NewAPIMeta(chatv1.KindMessage),
			From:    metav1.ObjectMeta{UID: r.uid},
			Content: chatv1.MessageContent{Join: &chatv1.MembersChangeMessageContent{User: *user}},
		}); err != nil {
			logger.Error(err, "send member join message error")
		}
	}

	return msgCh, r.stopListen(msgCh), nil
}

// stopListen 停止监听消息
func (r *localRoom) stopListen(ch chan *chatv1.Message) func() {
	return func() {
		r.lock.Lock()
		user := r.channels[ch]
		delete(r.channels, ch)
		close(ch)
		r.lock.Unlock()
		if user != nil {
			_ = r.CreateMessage(context.Background(), &chatv1.Message{
				APIMeta: metav1.NewAPIMeta(chatv1.KindMessage),
				From:    metav1.ObjectMeta{UID: r.uid},
				Content: chatv1.MessageContent{Leave: &chatv1.MembersChangeMessageContent{User: *user}},
			})
		}
	}
}

// Close 关闭
func (r *localRoom) Close(_ context.Context) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	for ch := range r.channels {
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
	upstreamDeduplicator := deduplicators.NewBloomFilter(500, 0.001)
	done := make(chan struct{})
	go r.listenUpstream(ctx, r.upstream, done, upstreamDeduplicator)
	go r.forwardToUpstream(ctx, r.upstream, done, upstreamDeduplicator)

	return nil
}

// forwardToUpstream 转发消息给上游
func (r *localRoom) forwardToUpstream(
	ctx context.Context,
	upstream Room,
	done <-chan struct{},
	upstreamDeduplicator deduplicators.Deduplicator,
) {
	logger := logr.FromContextOrDiscard(ctx)

	defer func() {
		r.lock.Lock()
		if r.upstream == upstream {
			r.upstream = nil
		}
		r.lock.Unlock()
		_ = upstream.Close(ctx)
	}()

	ch, stop, err := r.Listen(ctx, nil)
	if err != nil {
		logger.Error(err, "listen error")
		return
	}
	defer stop()

	for {
		var msg *chatv1.Message
		var ok bool
		select {
		case <-done:
			return
		case msg, ok = <-ch:
			if !ok {
				return
			}
		}

		if upstreamDeduplicator.Duplicate(msg.Meta.UID[:]) {
			continue
		}
		if err := r.upstream.CreateMessage(ctx, msg); err != nil {
			logger.Error(err, "forward to upstream error")
		}
	}
}

// listenUpstream 监听上游房间
func (r *localRoom) listenUpstream(
	ctx context.Context,
	upstream Room,
	done chan<- struct{},
	upstreamDeduplicator deduplicators.Deduplicator,
) {
	logger := logr.FromContextOrDiscard(ctx)

	defer func() {
		r.lock.Lock()
		if r.upstream == upstream {
			r.upstream = nil
		}
		r.lock.Unlock()
		close(done)
		_ = upstream.Close(ctx)
	}()

	ch, stop, err := upstream.Listen(ctx, &metav1.ObjectMeta{UID: r.ownerUID, Name: r.ownerName})
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
