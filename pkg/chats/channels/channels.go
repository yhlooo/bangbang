package channels

import (
	"errors"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
)

// Channel 接收聊天消息通道
type Channel interface {
	// Messages 获取接收消息的通道
	Messages() <-chan *chatv1.Message
	// Done 获取关闭或完成通知通道
	Done() <-chan struct{}
	// Close 关闭通道
	Close() error
}

// ChannelWithSender 带发送端的聊天消息通道
type ChannelWithSender interface {
	Channel

	// Send 发送消息到通道
	Send(msg *chatv1.Message) error
}

var (
	// ErrChannelClosed 通道已关闭
	ErrChannelClosed = errors.New("ChannelClosed")
	// ErrChannelBusy 通道忙
	ErrChannelBusy = errors.New("ChannelBusy")
)
