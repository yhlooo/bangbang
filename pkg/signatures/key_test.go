package signatures

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestKey_Copy 测试 Key.Copy
func TestKey_Copy(t *testing.T) {
	a := assert.New(t)

	key1 := Key("hello")
	key2 := key1.Copy()
	key2[2] = 'w'

	a.Equal("hello", string(key1))
	a.Equal("hewlo", string(key2))
}
