package v1

import metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"

// User 用户
type User struct {
	metav1.APIMeta
	Meta metav1.ObjectMeta `json:"meta,omitempty"`
}

// UserList 用户列表
type UserList struct {
	metav1.APIMeta

	Items []User `json:"items"`
}
