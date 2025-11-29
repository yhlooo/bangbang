package v1

import metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"

const (
	KindRoom        = "Room"
	KindRoomRequest = "RoomRequest"
)

// Room 房间
type Room struct {
	metav1.APIMeta
	Meta metav1.ObjectMeta `json:"meta,omitempty"`

	// 房主
	Owner User `json:"owner,omitempty"`

	// 密钥签名
	KeySignature string `json:"keySignature,omitempty"`
	// 访问端点地址
	Endpoints []string `json:"endpoints,omitempty"`
}

// RoomRequest 房间请求
type RoomRequest struct {
	metav1.APIMeta
	Meta metav1.ObjectMeta `json:"meta,omitempty"`

	// 密钥签名
	KeySignature string `json:"keySignature,omitempty"`
}
