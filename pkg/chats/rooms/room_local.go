package rooms

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
)

// NewLocalRoom 创建本地房间实例
func NewLocalRoom(uid, ownerUID string) Room {
	return &localRoom{
		uid:      uid,
		ownerUID: ownerUID,
	}
}

// localRoom 是 Room 的本地实现
type localRoom struct {
	uid, ownerUID string

	lock sync.RWMutex

	closed bool

	members  map[string]struct{}
	channels map[chan *chatv1.Message]struct{}
}

var _ Room = (*localRoom)(nil)

// Info 获取房间信息
func (r *localRoom) Info(_ context.Context) (*RoomInfo, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	var members []string
	if len(r.members) > 0 {
		members = make([]string, 0, len(r.members))
		for memberUID := range r.members {
			members = append(members, memberUID)
		}
	}
	sort.Strings(members)

	return &RoomInfo{
		UID:     r.uid,
		Owner:   r.ownerUID,
		Members: members,
	}, nil
}

// Join 加入房间
func (r *localRoom) Join(_ context.Context, userUID string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.closed {
		return fmt.Errorf("room already closed")
	}

	if r.members == nil {
		r.members = make(map[string]struct{})
	}
	r.members[userUID] = struct{}{}

	return nil
}

// Leave 离开房间
func (r *localRoom) Leave(_ context.Context, userUID string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.members, userUID)

	return nil
}

// CreateMessage 创建消息
func (r *localRoom) CreateMessage(_ context.Context, msg *chatv1.Message) error {
	r.lock.RLock()
	defer r.lock.RUnlock()

	msg.Meta.UID = uuid.New().String()

	for ch := range r.channels {
		select {
		case ch <- msg:
		default:
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
