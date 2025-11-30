package rooms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-logr/logr"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
)

// NewRemoteRoom 创建远程房间实例
func NewRemoteRoom(endpoint string) Room {
	return &remoteRoom{endpoint: endpoint}
}

// remoteRoom 是基于 API 访问远程房间的实现的 Room
type remoteRoom struct {
	endpoint string

	lock      sync.RWMutex
	closed    bool
	stopFuncs []func()
}

var _ Room = (*remoteRoom)(nil)

// Info 获取房间信息
func (r *remoteRoom) Info(ctx context.Context) (*RoomInfo, error) {
	info := &chatv1.Room{}
	if err := r.doRequest(ctx, http.MethodGet, "/info", nil, info); err != nil {
		return nil, err
	}
	return &RoomInfo{
		UID:                   info.Meta.UID,
		OwnerUID:              info.Owner.Meta.UID,
		OwnerName:             info.Owner.Meta.Name,
		PublishedKeySignature: info.KeySignature,
	}, nil
}

// CreateMessage 创建消息
func (r *remoteRoom) CreateMessage(ctx context.Context, msg *chatv1.Message) error {
	r.lock.RLock()
	if r.closed {
		r.lock.RUnlock()
		return fmt.Errorf("room already closed")
	}
	r.lock.RUnlock()
	return r.doRequest(ctx, http.MethodPost, "/messages", msg, msg)
}

// Listen 获取监听消息的信道
func (r *remoteRoom) Listen(
	ctx context.Context,
	user *metav1.ObjectMeta,
) (ch <-chan *chatv1.Message, stop func(), err error) {
	logger := logr.FromContextOrDiscard(ctx)

	r.lock.Lock()
	defer r.lock.Unlock()

	if r.closed {
		return nil, nil, fmt.Errorf("room already closed")
	}

	uri := "/messages"
	if user != nil {
		uri += "?" + url.Values{
			"userUID":  {user.UID.String()},
			"userName": {user.Name},
		}.Encode()
	}

	// 构造请求
	ctx, cancel := context.WithCancel(ctx)
	resp, err := r.doGetStreamRequest(ctx, uri)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("make request error: %w", err)
	}

	msgCh := make(chan *chatv1.Message, 10)
	go func() {
		defer func() {
			close(msgCh)
			_ = resp.Body.Close()
			cancel()
		}()

		decoder := json.NewDecoder(resp.Body)
		for decoder.More() {
			msg := &chatv1.Message{}
			if err := decoder.Decode(msg); err != nil {
				logger.Error(err, "decode message error")
				return
			}
			if !msg.IsKind(chatv1.KindMessage) {
				// TODO: metav1.Status 应该反序列化出来
				logger.Info("invalid message")
				continue
			}

			select {
			case <-ctx.Done():
				return
			case msgCh <- msg:
			}
		}
	}()

	r.stopFuncs = append(r.stopFuncs, cancel)
	return msgCh, cancel, nil
}

// Close 关闭
func (r *remoteRoom) Close(_ context.Context) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.closed = true
	for _, stop := range r.stopFuncs {
		stop()
	}
	return nil
}

// doRequest 发送请求
func (r *remoteRoom) doRequest(ctx context.Context, method, uri string, reqData, respData interface{}) error {
	// 构造请求
	req, err := r.makeRequest(ctx, method, uri, reqData)
	if err != nil {
		return fmt.Errorf("make request error: %w", err)
	}

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 读取响应
	respBodyRaw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read response body error: %w", err)
	}
	if resp.ContentLength > 1<<20 {
		return fmt.Errorf(
			"response body too large: %d Bytes (> 1MiB), first 1MiB: %s",
			resp.ContentLength, string(respBodyRaw),
		)
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := metav1.Status{}
		if err := json.Unmarshal(respBodyRaw, &apiErr); err != nil {
			return fmt.Errorf("unexpected status code: %d (!= 200), body: %s", resp.StatusCode, string(respBodyRaw))
		}
		return &apiErr
	}

	// 反序列化
	if respData != nil {
		if err := json.Unmarshal(respBodyRaw, respData); err != nil {
			return fmt.Errorf("decode response body from json erron: %w, body: %s", err, string(respBodyRaw))
		}
	}

	return nil
}

// doGetStreamRequest 发送获取流请求
func (r *remoteRoom) doGetStreamRequest(ctx context.Context, uri string) (*http.Response, error) {
	// 构造请求
	req, err := r.makeRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("make request error: %w", err)
	}

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		respBodyRaw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
		return nil, fmt.Errorf(
			"unexpected status code: %d (!= 200), body: %s",
			resp.StatusCode, string(respBodyRaw),
		)
	}

	return resp, nil
}

// makeRequest 构造请求
func (r *remoteRoom) makeRequest(ctx context.Context, method, uri string, reqData interface{}) (*http.Request, error) {
	var reqBody io.Reader
	if reqData != nil {
		reqDataRaw, err := json.Marshal(reqData)
		if err != nil {
			return nil, fmt.Errorf("encode request data to json error: %w", err)
		}
		reqBody = bytes.NewReader(reqDataRaw)
	}
	return http.NewRequestWithContext(
		ctx,
		method,
		fmt.Sprintf("%s/chat/v1%s", r.endpoint, uri),
		reqBody,
	)
}
