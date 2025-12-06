package signatures

// Key 密钥
type Key []byte

// Copy 拷贝
func (k Key) Copy() Key {
	if k == nil {
		return nil
	}
	out := make([]byte, len(k))
	copy(out, k)
	return out
}
