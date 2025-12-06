package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/go-logr/logr"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
	"github.com/yhlooo/bangbang/pkg/chats/rooms"
	"github.com/yhlooo/bangbang/pkg/signatures"
)

// NewUDPDiscoverer 创建基于 UDP 的发现器
func NewUDPDiscoverer(addr string) *UDPDiscoverer {
	return &UDPDiscoverer{
		addr: addr,
	}
}

// UDPDiscoverer 基于 UDP 的发现器
type UDPDiscoverer struct {
	addr string
}

var _ Discoverer = (*UDPDiscoverer)(nil)

// Search 搜索房间
func (d *UDPDiscoverer) Search(ctx context.Context, key signatures.Key, opts SearchOptions) ([]Room, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("discoverer")
	ctx = logr.NewContext(ctx, logger)

	if opts.Duration == 0 {
		opts.Duration = 3 * time.Second
	}
	if opts.RequestInterval == 0 {
		opts.RequestInterval = time.Second
	}

	addr, err := net.ResolveUDPAddr("udp", d.addr)
	if err != nil {
		return nil, fmt.Errorf("resolve udp address error: %w", err)
	}

	readConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen udp %q error: %w", addr.String(), err)
	}
	writeConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("dial udp %q error: %w", addr.String(), err)
	}

	go d.runSender(ctx, writeConn, key.Copy(), int(opts.Duration/opts.RequestInterval), opts.RequestInterval)

	logger.V(1).Info("listening rooms ...")
	ret, err := d.runListener(ctx, readConn, key.Copy(), opts.RequestInterval, opts.Exclude)
	if err != nil {
		return nil, fmt.Errorf("run listener error: %w", err)
	}
	logger.V(1).Info(fmt.Sprintf("found %d rooms", len(ret)))

	if opts.CheckAvailability {
		logger.V(1).Info("checking availability for rooms ...")
		d.checkAvailability(ctx, key.Copy(), ret)
	}

	return ret, nil
}

// runListener 运行监听器
func (d *UDPDiscoverer) runListener(
	ctx context.Context,
	conn *net.UDPConn,
	key signatures.Key,
	timeout time.Duration,
	exclude []metav1.UID,
) ([]Room, error) {
	logger := logr.FromContextOrDiscard(ctx).WithName("listener")
	ctx = logr.NewContext(ctx, logger)

	roomMap := map[metav1.UID]chatv1.Room{}

	go func() {
		select {
		case <-ctx.Done():
		case <-time.After(timeout):
		}
		_ = conn.Close()
	}()
	_ = conn.SetReadBuffer(1 << 20)

	buffer := make([]byte, 8<<10)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if n == 0 {
				break
			}
			logger.Error(err, "read udp packet error")
		}
		if n == 0 {
			continue
		}

		var room chatv1.Room
		if err := json.Unmarshal(buffer[:n], &room); err != nil {
			logger.Error(err, fmt.Sprintf("decode room error: %s", string(buffer[:n])))
			continue
		}

		if !room.IsKind(chatv1.KindRoom) {
			continue
		}
		if slices.Contains(exclude, room.UID) {
			continue
		}
		if key != nil {
			now := time.Now()
			if err := signatures.HS256VerifyAPIObject(
				key, &room,
				now.Add(-10*time.Minute), now.Add(10*time.Minute),
			); err != nil {
				logger.V(1).Info(fmt.Sprintf("signature verification error: %s", err))
				continue
			}
		}

		logger.V(1).Info(fmt.Sprintf("found room %q", room.UID))
		roomMap[room.UID] = room
	}

	if len(roomMap) == 0 {
		return nil, nil
	}

	ret := make([]Room, 0, len(roomMap))
	for _, room := range roomMap {
		ret = append(ret, Room{
			Info: room,
		})
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Info.UID.String() < ret[j].Info.UID.String()
	})

	return ret, nil
}

// runSender 运行发送器
func (d *UDPDiscoverer) runSender(
	ctx context.Context,
	conn *net.UDPConn,
	key signatures.Key,
	n int,
	interval time.Duration,
) {
	req := &chatv1.RoomRequest{
		APIMeta:    metav1.NewAPIMeta(chatv1.KindRoomRequest),
		ObjectMeta: metav1.ObjectMeta{UID: metav1.NewUID()},
	}
	var reqRaw []byte
	if key == nil {
		reqRaw, _ = json.Marshal(req)
		reqRaw = append(reqRaw, '\n')
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger := logr.FromContextOrDiscard(ctx).WithName("sender")
	for i := 0; i < n; i++ {
		if key != nil {
			_ = signatures.HS256SignAPIObject(key, req)
			reqRaw, _ = json.Marshal(req)
			reqRaw = append(reqRaw, '\n')
		}

		logger.V(1).Info("sending room request")
		_, err := conn.Write(reqRaw)
		if err != nil {
			logger.Error(err, "send room request error")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// checkAvailability 检查房间可访问性
func (d *UDPDiscoverer) checkAvailability(ctx context.Context, key signatures.Key, roomList []Room) {
	logger := logr.FromContextOrDiscard(ctx)
	for i, room := range roomList {
		available := ""
		for _, endpoint := range room.Info.Endpoints {
			subCTX, cancel := context.WithTimeout(ctx, time.Second)
			info, err := rooms.NewRemoteRoom(endpoint, room.Info.CertSign).Info(subCTX)
			cancel()
			if err != nil {
				logger.V(1).Info(fmt.Sprintf(
					"endpoint %q for room %q not available: %v",
					endpoint, room.Info.UID, err,
				))
				continue
			}
			if info.UID != room.Info.UID {
				logger.V(1).Info(fmt.Sprintf(
					"endpoint %q for room %q not available: uid not match: %q",
					endpoint, room.Info.UID, info.UID,
				))
				continue
			}
			if key != nil {
				now := time.Now()
				if err := signatures.HS256VerifyAPIObject(
					key, info,
					now.Add(-10*time.Minute), now.Add(10*time.Minute),
				); err != nil {
					logger.V(1).Info(fmt.Sprintf(
						"endpoint %q for room %q not available, signature verification error: %s",
						endpoint, room.Info.UID, err.Error(),
					))
					continue
				}
			}
			available = endpoint
			break
		}

		if available != "" {
			roomList[i].AvailableEndpoint = available
			logger.V(1).Info(fmt.Sprintf("room %q available on %q", room.Info.UID, available))
		} else {
			logger.V(1).Info(fmt.Sprintf("room %q has no available endpoint", room.Info.UID))
		}
	}
}

// NewUDPTransponder 创建基于 UDP 的应答机
func NewUDPTransponder(addr string, room *chatv1.Room, key signatures.Key) *UDPTransponder {
	return &UDPTransponder{
		addr: addr,
		room: room.DeepCopy(),
		key:  key.Copy(),
	}
}

// UDPTransponder 基于 UDP 的应答机
type UDPTransponder struct {
	once sync.Once

	addr string
	room *chatv1.Room
	key  signatures.Key

	readConn  *net.UDPConn
	writeConn *net.UDPConn
}

var _ Transponder = (*UDPTransponder)(nil)

// Start 开始运行应答机
func (t *UDPTransponder) Start(ctx context.Context) error {
	finalErr := fmt.Errorf("already started")
	t.once.Do(func() {
		addr, err := net.ResolveUDPAddr("udp", t.addr)
		if err != nil {
			finalErr = fmt.Errorf("resolve udp address error: %w", err)
			return
		}

		t.readConn, err = net.ListenUDP("udp", addr)
		if err != nil {
			finalErr = fmt.Errorf("listen udp %q error: %w", addr.String(), err)
			return
		}
		t.writeConn, err = net.DialUDP("udp", nil, addr)
		if err != nil {
			finalErr = fmt.Errorf("dial udp %q error: %w", addr.String(), err)
			return
		}

		finalErr = nil
		ch := make(chan struct{})

		go t.runListener(ctx, ch)
		go t.runSender(ctx, ch, t.room.DeepCopy())

		return
	})
	return finalErr
}

// runListener 运行监听器
func (t *UDPTransponder) runListener(ctx context.Context, ch chan<- struct{}) {
	logger := logr.FromContextOrDiscard(ctx).WithName("transponder.listener")

	defer close(ch)

	buffer := make([]byte, 8<<10)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n, _, err := t.readConn.ReadFromUDP(buffer)
		if err != nil {
			logger.Error(err, "read udp packet error")
		}
		if n == 0 {
			continue
		}

		var req chatv1.RoomRequest
		if err := json.Unmarshal(buffer[:n], &req); err != nil {
			logger.Error(err, "decode room request error: %s", string(buffer[:n]))
		}
		if !req.IsKind(chatv1.KindRoomRequest) {
			continue
		}
		if req.Signature != "" {
			now := time.Now()
			if err := signatures.HS256VerifyAPIObject(
				t.key, &req,
				now.Add(-10*time.Minute), now.Add(10*time.Minute),
			); err != nil {
				logger.V(1).Info(fmt.Sprintf("signature verification error: %s", err))
				continue
			}
		}
		select {
		case <-ctx.Done():
			return
		case ch <- struct{}{}:
		}
	}
}

// runSender 运行发送器
func (t *UDPTransponder) runSender(ctx context.Context, ch <-chan struct{}, room *chatv1.Room) {
	logger := logr.FromContextOrDiscard(ctx).WithName("transponder.sender")

	defer func() {
		_ = t.readConn.Close()
		_ = t.writeConn.Close()
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

		if err := signatures.HS256SignAPIObject(t.key, room); err != nil {
			logger.Error(err, "sign room info error")
			continue
		}
		publishMsg, err := json.Marshal(room)
		if err != nil {
			logger.Error(err, "marshal room info to json error")
			continue
		}
		publishMsg = append(publishMsg, '\n')

		if _, err := t.writeConn.Write(publishMsg); err != nil {
			logger.Error(err, "publish error")
		}
	}
}
