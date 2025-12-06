package v1

import metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"

const KindMessage = "Message"

// Message 消息
type Message struct {
	metav1.APIMeta
	metav1.ObjectMeta `json:"meta,omitempty"`

	// 发送人
	From metav1.ObjectMeta `json:"from,omitempty"`
	// 消息内容
	Content MessageContent `json:"content,omitempty"`
}

var _ metav1.Object = (*Message)(nil)

// DeepCopy 深拷贝
func (obj *Message) DeepCopy() *Message {
	if obj == nil {
		return nil
	}
	return &Message{
		APIMeta:    *obj.APIMeta.DeepCopy(),
		ObjectMeta: *obj.ObjectMeta.DeepCopy(),
		From:       *obj.From.DeepCopy(),
		Content:    *obj.Content.DeepCopy(),
	}
}

// MessageContent 消息内容
//
// NOTE: 根据内容类型不同，仅一个属性有值
type MessageContent struct {
	// 文本消息内容
	Text *TextMessageContent `json:"text,omitempty"`
	// 成员加入
	Join *MembersChangeMessageContent `json:"join,omitempty"`
	// 成员离开
	Leave *MembersChangeMessageContent `json:"leave,omitempty"`
}

// DeepCopy 深拷贝
func (obj *MessageContent) DeepCopy() *MessageContent {
	if obj == nil {
		return nil
	}
	return &MessageContent{
		Text:  obj.Text.DeepCopy(),
		Join:  obj.Join.DeepCopy(),
		Leave: obj.Leave.DeepCopy(),
	}
}

// TextMessageContent 文本消息内容
type TextMessageContent struct {
	Content string `json:"content,omitempty"`
}

// DeepCopy 深拷贝
func (obj *TextMessageContent) DeepCopy() *TextMessageContent {
	if obj == nil {
		return nil
	}
	return &TextMessageContent{
		Content: obj.Content,
	}
}

// MembersChangeMessageContent 成员变化消息
type MembersChangeMessageContent struct {
	User metav1.ObjectMeta `json:"user,omitempty"`
}

// DeepCopy 深拷贝
func (obj *MembersChangeMessageContent) DeepCopy() *MembersChangeMessageContent {
	if obj == nil {
		return nil
	}
	return &MembersChangeMessageContent{
		User: *obj.User.DeepCopy(),
	}
}
