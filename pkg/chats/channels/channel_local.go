package channels

import (
	"sync"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
)

// NewLocalChannel 创建基于内存的 Channel
func NewLocalChannel(bufSize int) ChannelWithSender {
	return &localChannel{
		ch:   make(chan *chatv1.Message, bufSize),
		done: make(chan struct{}),
	}
}

// localChannel 基于内存的 Channel 实现
type localChannel struct {
	lock   sync.RWMutex
	closed bool
	ch     chan *chatv1.Message
	done   chan struct{}
}

var _ ChannelWithSender = (*localChannel)(nil)

// Send 发送消息到通道
func (ch *localChannel) Send(msg *chatv1.Message) error {
	ch.lock.RLock()
	defer ch.lock.RUnlock()
	if ch.closed {
		return ErrChannelClosed
	}
	select {
	case ch.ch <- msg:
	default:
		return ErrChannelBusy
	}
	return nil
}

// Messages 获取接收消息的通道
func (ch *localChannel) Messages() <-chan *chatv1.Message {
	return ch.ch
}

// Done 获取关闭或完成通知通道
func (ch *localChannel) Done() <-chan struct{} {
	return ch.done
}

// Close 关闭通道
func (ch *localChannel) Close() error {
	ch.lock.Lock()
	defer ch.lock.Unlock()
	if ch.closed {
		return ErrChannelClosed
	}
	close(ch.ch)
	close(ch.done)
	ch.closed = true
	return nil
}
