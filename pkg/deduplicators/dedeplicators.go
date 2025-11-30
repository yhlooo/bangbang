package deduplicators

// Deduplicator 去重器
type Deduplicator interface {
	// Duplicate 校验是否重复的并记录下该内容
	Duplicate(data []byte) bool
}
