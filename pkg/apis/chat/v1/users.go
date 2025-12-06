package v1

import metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"

const (
	KindUser     = "User"
	KindUserList = "UserList"
)

// User 用户
type User struct {
	metav1.APIMeta
	metav1.ObjectMeta `json:"meta,omitempty"`
}

var _ metav1.Object = (*User)(nil)

// DeepCopy 深拷贝
func (obj *User) DeepCopy() *User {
	if obj == nil {
		return nil
	}
	return &User{
		APIMeta:    *obj.APIMeta.DeepCopy(),
		ObjectMeta: *obj.ObjectMeta.DeepCopy(),
	}
}

// UserList 用户列表
type UserList struct {
	metav1.APIMeta

	Items []User `json:"items"`
}

// DeepCopy 深拷贝
func (obj *UserList) DeepCopy() *UserList {
	if obj == nil {
		return nil
	}
	var items []User
	if obj.Items != nil {
		items = make([]User, len(obj.Items))
		for i, item := range obj.Items {
			items[i] = *item.DeepCopy()
		}
	}
	return &UserList{
		APIMeta: *obj.APIMeta.DeepCopy(),
		Items:   items,
	}
}
