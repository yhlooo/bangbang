package v1

import metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"

const (
	KindRoom        = "Room"
	KindRoomRequest = "RoomRequest"
)

// Room 房间
type Room struct {
	metav1.APIMeta
	metav1.ObjectMeta `json:"meta,omitempty"`

	// 房主
	Owner User `json:"owner,omitempty"`
	// 证书签名
	CertSign string `json:"certSign,omitempty"`
	// 访问端点地址
	Endpoints []string `json:"endpoints,omitempty"`
}

var _ metav1.Object = (*Room)(nil)

// DeepCopy 深拷贝
func (obj *Room) DeepCopy() *Room {
	if obj == nil {
		return nil
	}
	var endpoints []string
	if obj.Endpoints != nil {
		endpoints = make([]string, len(obj.Endpoints))
		copy(endpoints, obj.Endpoints)
	}
	return &Room{
		APIMeta:    *obj.APIMeta.DeepCopy(),
		ObjectMeta: *obj.ObjectMeta.DeepCopy(),
		Owner:      *obj.Owner.DeepCopy(),
		CertSign:   obj.CertSign,
		Endpoints:  endpoints,
	}
}

// RoomRequest 房间请求
type RoomRequest struct {
	metav1.APIMeta
	metav1.ObjectMeta `json:"meta,omitempty"`
}

var _ metav1.Object = (*RoomRequest)(nil)

// DeepCopy 深拷贝
func (obj *RoomRequest) DeepCopy() *RoomRequest {
	if obj == nil {
		return nil
	}
	return &RoomRequest{
		APIMeta:    *obj.APIMeta.DeepCopy(),
		ObjectMeta: *obj.ObjectMeta.DeepCopy(),
	}
}
