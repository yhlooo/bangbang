package v1

// APIMeta API 元信息
type APIMeta struct {
	// API 版本
	Version string `json:"version"`
}

// ObjectMeta 对象元信息
type ObjectMeta struct {
	// 对象唯一 ID
	UID string `json:"uid,omitempty"`
	// 对象名
	Name string `json:"name,omitempty"`
}
