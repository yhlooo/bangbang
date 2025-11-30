package v1

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

const (
	// Version 版本
	Version = "v1"
	// KindStatus 状态
	KindStatus = "Status"
)

// APIMeta API 元信息
type APIMeta struct {
	// API 版本
	Version string `json:"version"`
	// 类型
	Kind string `json:"kind"`
}

// IsKind 判断是否指定类型
func (m APIMeta) IsKind(kind string) bool {
	if m.Version != Version {
		return false
	}
	return m.Kind == kind
}

// NewAPIMeta 创建 API 元信息
func NewAPIMeta(kind string) APIMeta {
	return APIMeta{
		Version: Version,
		Kind:    kind,
	}
}

// ObjectMeta 对象元信息
type ObjectMeta struct {
	// 对象唯一 ID
	UID UID `json:"uid,omitempty"`
	// 对象名
	Name string `json:"name,omitempty"`
}

// NewUID 创建一个 UID
func NewUID() UID {
	return UID(uuid.New())
}

// UID 唯一 ID
type UID uuid.UUID

// MarshalJSON 序列化为 JSON
//
//goland:noinspection GoMixedReceiverTypes
func (uid *UID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uid.String())
}

// UnmarshalJSON 从 JSON 反序列化
//
//goland:noinspection GoMixedReceiverTypes
func (uid *UID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	ret, err := uuid.Parse(s)
	if err != nil {
		return err
	}
	*uid = UID(ret)
	return nil
}

// IsNil 判断是否零值
func (uid UID) IsNil() bool {
	return uuid.UUID(uid) == uuid.Nil
}

// String 返回字符串形式
func (uid UID) String() string {
	return uuid.UUID(uid).String()
}

// Short 返回短字符串形式
func (uid UID) Short() string {
	if uid.IsNil() {
		return ""
	}
	sum := sha1.Sum(uid[:])
	return base32.StdEncoding.EncodeToString(sum[:5])
}

// Status 接口状态
type Status struct {
	APIMeta
	Meta ObjectMeta `json:"meta,omitempty"`

	// HTTP 状态码
	Code int `json:"code,omitempty"`
	// 可枚举的原因
	Reason string `json:"reason,omitempty"`
	// 人类可读的描述
	Message string `json:"message,omitempty"`
}

// Error 返回字符串形式的错误描述
func (s *Status) Error() string {
	return fmt.Sprintf("%s(%d): %s (uid:%s)", s.Reason, s.Code, s.Message, s.Meta.UID)
}
