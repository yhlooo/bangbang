package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUID_UnmarshalJSON 测试 UID.UnmarshalJSON 方法
func TestUID_UnmarshalJSON(t *testing.T) {
	a := assert.New(t)

	uid := UID{}
	a.NoError(uid.UnmarshalJSON([]byte(`"12345678-1234-1234-1234-1234567890ab"`)))
	a.Equal("12345678-1234-1234-1234-1234567890ab", uid.String())
}
