package v1

import (
	"fmt"
)

// Version 版本
const Version = "v1"

// APIMeta API 元信息
type APIMeta struct {
	// API 版本
	Version string `json:"version"`
}

// NewAPIMeta 创建 API 元信息
func NewAPIMeta() APIMeta {
	return APIMeta{
		Version: Version,
	}
}

// ObjectMeta 对象元信息
type ObjectMeta struct {
	// 对象唯一 ID
	UID string `json:"uid,omitempty"`
	// 对象名
	Name string `json:"name,omitempty"`
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
