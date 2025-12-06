package managers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"time"

	"github.com/go-logr/logr"

	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
	"github.com/yhlooo/bangbang/pkg/chats/rooms"
	"github.com/yhlooo/bangbang/pkg/discovery"
	"github.com/yhlooo/bangbang/pkg/servers"
	"github.com/yhlooo/bangbang/pkg/signatures"
)

// Options 运行选项
type Options struct {
	Key signatures.Key
	// 房间所有者 UID
	OwnerUID metav1.UID
	// 房间所有者名
	OwnerName string
	// HTTP 监听地址
	HTTPAddr string
	// 服务发现地址
	DiscoveryAddr string
}

// Validate 校验选项
func (o *Options) Validate() error {
	if len(o.Key) == 0 {
		return errors.New(".Key is required")
	}
	if o.HTTPAddr == "" {
		return errors.New(".HTTPAddr is required")
	}
	if o.DiscoveryAddr == "" {
		return errors.New(".DiscoveryAddr is required")
	}
	return nil
}

// NewManager 创建聊天管理器
func NewManager(opts Options) (Manager, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	return &defaultManager{
		opts:       opts,
		selfRoom:   rooms.NewLocalRoom(opts.Key, opts.OwnerUID, opts.OwnerName),
		discoverer: discovery.NewUDPDiscoverer(opts.DiscoveryAddr),
	}, nil
}

// defaultManager 是 Manager 的默认实现
type defaultManager struct {
	opts Options

	selfRoom   rooms.RoomWithUpstream
	discoverer discovery.Discoverer

	listenAddr net.Addr
	certSign   string
}

var _ Manager = (*defaultManager)(nil)

// SelfRoom 获取自己主持的房间
func (mgr *defaultManager) SelfRoom(_ context.Context) rooms.Room {
	return mgr.selfRoom
}

// StartServer 开始运行 HTTP 服务
func (mgr *defaultManager) StartServer(ctx context.Context) (<-chan struct{}, error) {
	addr, certSign, done, err := servers.RunServer(ctx, servers.Options{
		ListenAddr: mgr.opts.HTTPAddr,
		Room:       mgr.SelfRoom(ctx),
	})
	if err != nil {
		return nil, err
	}
	mgr.listenAddr = addr
	mgr.certSign = certSign
	return done, nil
}

// StartSearchUpstream 开始搜索上游
func (mgr *defaultManager) StartSearchUpstream(ctx context.Context) error {
	logger := logr.FromContextOrDiscard(ctx)

	selfRoom, err := mgr.SelfRoom(ctx).Info(ctx)
	if err != nil {
		return fmt.Errorf("get self room info error: %w", err)
	}

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}

			if mgr.selfRoom.Upstream() != nil {
				// 已经有上游了
				continue
			}

			roomList, err := mgr.discoverer.Search(ctx, mgr.opts.Key, discovery.SearchOptions{
				CheckAvailability: true,
				Exclude:           []metav1.UID{selfRoom.UID},
			})
			if err != nil {
				logger.Error(err, "search rooms error")
				continue
			}

			for _, room := range roomList {
				if room.Info.UID == selfRoom.UID {
					// 跳过自己房间
					continue
				}
				if room.AvailableEndpoint == "" {
					// 跳过不可用的
					logger.V(1).Info(fmt.Sprintf("skip unavailable room: %s", room.Info.UID))
					continue
				}

				if err := mgr.selfRoom.SetUpstream(
					ctx,
					rooms.NewRemoteRoom(room.AvailableEndpoint, room.Info.CertSign),
				); err != nil {
					logger.Error(err, "set upstream error")
					continue
				}
				break
			}
		}
	}()

	return nil
}

// StartTransponder 开始运行应答机
func (mgr *defaultManager) StartTransponder(ctx context.Context) error {
	selfRoom, err := mgr.SelfRoom(ctx).Info(ctx)
	if err != nil {
		return fmt.Errorf("get self room info error: %w", err)
	}

	selfRoom.Endpoints, err = mgr.getEndpoints(ctx)
	if err != nil {
		return fmt.Errorf("get endpoints error: %w", err)
	}
	selfRoom.CertSign = mgr.certSign

	t := discovery.NewUDPTransponder(mgr.opts.DiscoveryAddr, selfRoom, mgr.opts.Key)

	return t.Start(ctx)
}

// getEndpoints 获取可能能访问房间的端点
func (mgr *defaultManager) getEndpoints(ctx context.Context) ([]string, error) {
	logger := logr.FromContextOrDiscard(ctx)

	if mgr.listenAddr == nil {
		return nil, nil
	}

	// 解析当前监听地址
	listenAddr, err := net.ResolveTCPAddr("tcp", mgr.listenAddr.String())
	if err != nil {
		return nil, fmt.Errorf("resolve address %q error: %w", mgr.listenAddr.String(), err)
	}
	port := listenAddr.Port
	allIfaces := false
	if listenAddr.IP == nil || listenAddr.IP.IsUnspecified() {
		allIfaces = true
	}

	// 获取所有网卡地址
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("get interfaces error: %w", err)
	}

	var ips []net.IP
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			logger.Error(err, fmt.Sprintf("get interface %q addresses error", iface.Name))
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if allIfaces || v.Contains(listenAddr.IP) {
					ips = append(ips, v.IP)
				}
			case *net.IPAddr:
				if allIfaces || v.IP.Equal(listenAddr.IP) {
					ips = append(ips, v.IP)
				}
			}
		}
	}

	// 对 IP 排序
	sort.Slice(ips, func(i, j int) bool {
		// IPv4 优先
		if len(ips[i]) != len(ips[j]) {
			return len(ips[i]) < len(ips[j])
		}
		// 私有地址优先
		if ips[i].IsPrivate() != ips[j].IsPrivate() {
			return ips[i].IsPrivate()
		}
		// 本地回环优先
		if ips[i].IsLoopback() != ips[j].IsLoopback() {
			return ips[i].IsLoopback()
		}
		return ips[i].String() < ips[j].String()
	})

	ret := make([]string, len(ips))
	for i, ip := range ips {
		ret[i] = "https://" + (&net.TCPAddr{IP: ip, Port: port}).String()
	}

	return ret, nil
}
