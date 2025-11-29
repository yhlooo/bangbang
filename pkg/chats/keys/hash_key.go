package keys

import (
	"crypto/sha256"
	"fmt"
)

type HashKey []byte

// Verify 验证签名
func (key HashKey) Verify(signature string, published bool) bool {
	expected := ""
	if published {
		expected = key.PublishedSignature()
	} else {
		expected = key.PrivateSignature()
	}
	return signature == expected
}

// PrivateSignature 用于校验的私密签名
func (key HashKey) PrivateSignature() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(append([]byte("secret/"), key...)))
}

// PublishedSignature 用于公布的签名
func (key HashKey) PublishedSignature() string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(append([]byte("published/"), key...)))
}

// Copy 创建一个拷贝
func (key HashKey) Copy() HashKey {
	ret := make(HashKey, len(key))
	copy(ret, key)
	return ret
}
