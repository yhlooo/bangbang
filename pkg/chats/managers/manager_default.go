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
	mgrID := uuid.New().String()
	defaultRoomID := uuid.New().String()
	defaultRoom := rooms.NewLocalRoom(defaultRoomID, mgrID)
	mgr := &defaultManager{
		uid:           mgrID,
		defaultRoomID: defaultRoomID,
		rooms: map[string]rooms.Room{
			defaultRoomID: defaultRoom,
		},
	}
	return mgr
}

// defaultManager 是 Manager 的默认实现
type defaultManager struct {
	uid string

	lock          sync.RWMutex
	defaultRoomID string
	rooms         map[string]rooms.Room
}

var _ Manager = (*defaultManager)(nil)

// CreateLocalRoom 创建房间
func (mgr *defaultManager) CreateLocalRoom(_ context.Context) (rooms.Room, error) {
	uid := uuid.New().String()
	room := rooms.NewLocalRoom(uid, mgr.uid)

	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	if mgr.rooms == nil {
		mgr.rooms = make(map[string]rooms.Room)
	}
	mgr.rooms[uid] = room

	return room, nil
}

// GetLocalRoom 获取房间
func (mgr *defaultManager) GetLocalRoom(_ context.Context, uid string) (rooms.Room, error) {
	if uid == DefaultRoomID {
		uid = mgr.defaultRoomID
	}

	mgr.lock.RLock()
	defer mgr.lock.RUnlock()

	room, ok := mgr.rooms[uid]
	if !ok {
		return nil, fmt.Errorf("room %s not found", uid)
	}

	return room, nil
}

// DeleteLocalRoom 删除房间
func (mgr *defaultManager) DeleteLocalRoom(ctx context.Context, uid string) error {
	if uid == DefaultRoomID || uid == mgr.defaultRoomID {
		return fmt.Errorf("cannot delete default room")
	}

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
