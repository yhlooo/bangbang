package v1

import metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"

// Room 房间
type Room struct {
	metav1.APIMeta
	Meta metav1.ObjectMeta `json:"meta,omitempty"`

	// 房主
	Owner User `json:"owner,omitempty"`
}
