package v1

import metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"

// Message 消息
type Message struct {
	metav1.APIMeta
	Meta metav1.ObjectMeta `json:"meta,omitempty"`

	// 发送人
	From metav1.ObjectMeta `json:"from,omitempty"`
	// 消息内容
	Content MessageContent `json:"content,omitempty"`
}

// MessageContent 消息内容
//
// NOTE: 根据内容类型不同，仅一个属性有值
type MessageContent struct {
	// 文本消息内容
	Text *TextMessageContent `json:"text,omitempty"`
}

// TextMessageContent 文本消息内容
type TextMessageContent struct {
	Content string `json:"content,omitempty"`
}
