package managers

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/yhlooo/bangbang/pkg/chats/rooms"
)

// NewManager 创建聊天管理器
func NewManager() Manager {
	return &defaultManager{
		uid: uuid.New().String(),
	}
}

// defaultManager 是 Manager 的默认实现
type defaultManager struct {
	uid string

	lock  sync.RWMutex
	rooms map[string]rooms.Room
}

var _ Manager = (*defaultManager)(nil)

// CreateRoom 创建房间
func (mgr *defaultManager) CreateRoom(_ context.Context) (rooms.Room, error) {
	uid := uuid.New().String()
	room := rooms.NewRoom(uid, mgr.uid)

	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	mgr.rooms[uid] = room

	return room, nil
}

// GetRoom 获取房间
func (mgr *defaultManager) GetRoom(_ context.Context, uid string) (rooms.Room, error) {
	mgr.lock.RLock()
	defer mgr.lock.RUnlock()

	room, ok := mgr.rooms[uid]
	if !ok {
		return nil, fmt.Errorf("room %s not found", uid)
	}

	return room, nil
}

// DeleteRoom 删除房间
func (mgr *defaultManager) DeleteRoom(ctx context.Context, uid string) error {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	room, ok := mgr.rooms[uid]
	if !ok {
		return fmt.Errorf("room %s not found", uid)
	}

	if err := room.Close(ctx); err != nil {
		return fmt.Errorf("close room %s error: %w", uid, err)
	}
	delete(mgr.rooms, uid)

	return nil
}
