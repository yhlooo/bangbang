package managers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/go-logr/logr"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
	"github.com/yhlooo/bangbang/pkg/chats/keys"
	"github.com/yhlooo/bangbang/pkg/chats/rooms"
	"github.com/yhlooo/bangbang/pkg/servers"
)

// Options 运行选项
type Options struct {
	Key keys.HashKey
	// 房间所有者 UID
	OwnerUID string
	// HTTP 监听地址
	HTTPAddr string
	// 应答器地址
	TransponderAddr string
}

// Validate 校验选项
func (o *Options) Validate() error {
	if len(o.Key) == 0 {
		return errors.New(".Key is required")
	}
	if o.HTTPAddr == "" {
		return errors.New(".HTTPAddr is required")
	}
	if o.TransponderAddr == "" {
		return errors.New(".TransponderAddr is required")
	}
	return nil
}

// NewManager 创建聊天管理器
func NewManager(opts Options) (Manager, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	return &defaultManager{
		opts:     opts,
		selfRoom: rooms.NewLocalRoom(opts.Key, opts.OwnerUID),
	}, nil
}

// defaultManager 是 Manager 的默认实现
type defaultManager struct {
	opts Options

	selfRoom rooms.Room

	listenAddr net.Addr
}

var _ Manager = (*defaultManager)(nil)

// SelfRoom 获取自己主持的房间
func (mgr *defaultManager) SelfRoom(_ context.Context) rooms.Room {
	return mgr.selfRoom
}

// StartServer 开始运行 HTTP 服务
func (mgr *defaultManager) StartServer(ctx context.Context) (<-chan struct{}, error) {
	addr, done, err := servers.RunServer(ctx, servers.Options{
		ListenAddr: mgr.opts.HTTPAddr,
		Room:       mgr.SelfRoom(ctx),
	})
	if err != nil {
		return nil, err
	}
	mgr.listenAddr = addr
	return done, nil
}

// StartTransponder 开始运行应答机
func (mgr *defaultManager) StartTransponder(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	addr, err := net.ResolveUDPAddr("udp", mgr.opts.TransponderAddr)
	if err != nil {
		return fmt.Errorf("resolve udp address %q error: %w", mgr.opts.TransponderAddr, err)
	}

	writeConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("dial udp %q error: %w", addr.String(), err)
	}

	readConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("listen udp %q error: %w", addr.String(), err)
	}

	selfRoom, err := mgr.SelfRoom(ctx).Info(ctx)
	if err != nil {
		return fmt.Errorf("get self room info error: %w", err)
	}

	var endpoints []string

	publishMsg, err := json.Marshal(&chatv1.Room{
		APIMeta: metav1.NewAPIMeta(chatv1.KindRoom),
		Meta:    metav1.ObjectMeta{UID: selfRoom.UID},
		Owner: chatv1.User{
			APIMeta: metav1.NewAPIMeta(chatv1.KindUser),
			Meta:    metav1.ObjectMeta{UID: selfRoom.OwnerUID},
		},
		KeySignature: selfRoom.PublishedKeySignature,
		Endpoints:    endpoints,
	})
	if err != nil {
		return fmt.Errorf("marshal room info to json error: %w", err)
	}
	publishMsg = append(publishMsg, '\n')

	ch := make(chan struct{})

	go func() {
		defer close(ch)

		for {
			decoder := json.NewDecoder(readConn)
			for decoder.More() {
				var req chatv1.RoomRequest
				if err := decoder.Decode(&req); err != nil {
					logger.Error(err, "decode room request error")
					break
				}
				if !req.IsKind(chatv1.KindRoomRequest) {
					continue
				}
				if req.KeySignature != "" && req.KeySignature != selfRoom.PublishedKeySignature {
					continue
				}
				select {
				case <-ctx.Done():
					return
				case ch <- struct{}{}:
				}
			}
		}
	}()

	go func() {
		defer func() {
			_ = readConn.Close()
			_ = writeConn.Close()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
			}

			if _, err := writeConn.Write(publishMsg); err != nil {
				logger.Error(err, "publish error")
			}
		}
	}()
	return nil
}
