package v1

import (
	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
)

// Channel 信道
type Channel struct {
	metav1.APIMeta
	Meta metav1.ObjectMeta `json:"meta,omitempty"`

	// 信道所属房间
	RoomRef metav1.ObjectMeta `json:"roomRef,omitempty"`
	// 申请人
	Applicant User `json:"applicant,omitempty"`
}
