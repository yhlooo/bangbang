package signatures

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	chatv1 "github.com/yhlooo/bangbang/pkg/apis/chat/v1"
	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
)

// TestHS256SignAPIObject 测试 HS256SignAPIObject
func TestHS256SignAPIObject(t *testing.T) {
	a := assert.New(t)

	uid, _ := uuid.Parse("12345678-1234-1234-1234-1234567890ab")

	obj := &chatv1.Room{
		APIMeta: metav1.NewAPIMeta(chatv1.KindRoom),
		ObjectMeta: metav1.ObjectMeta{
			UID: metav1.UID(uid),
		},
		Owner: chatv1.User{
			APIMeta: metav1.NewAPIMeta(chatv1.KindUser),
			ObjectMeta: metav1.ObjectMeta{
				UID:  metav1.UID(uid),
				Name: "test-user",
			},
		},
		Endpoints: []string{
			"https://192.168.233.6",
		},
	}

	a.NoError(HS256SignAPIObject(Key("test-secret"), obj))
	a.False(obj.SignTime.IsZero())
	a.NotEmpty(obj.Signature)

	a.NoError(HS256VerifyAPIObject(Key("test-secret"), obj, time.Now().Add(-time.Minute), time.Now().Add(time.Minute)))
}

// TestHS256SignObject 测试 HS256SignObject
func TestHS256SignObject(t *testing.T) {
	a := assert.New(t)

	uid, _ := uuid.Parse("12345678-1234-1234-1234-1234567890ab")

	obj := &chatv1.Room{
		APIMeta: metav1.NewAPIMeta(chatv1.KindRoom),
		ObjectMeta: metav1.ObjectMeta{
			UID: metav1.UID(uid),
		},
		Owner: chatv1.User{
			APIMeta: metav1.NewAPIMeta(chatv1.KindUser),
			ObjectMeta: metav1.ObjectMeta{
				UID:  metav1.UID(uid),
				Name: "test-user",
			},
		},
		Endpoints: []string{
			"https://192.168.233.6",
		},
	}

	sign, err := HS256SignObject(Key("test-secret"), obj)
	a.NoError(err)
	a.Equal("hs256:11f3b4b3a71e7c98db8cda725f79a4db53431bb688be319fc2f73cb952fbc983", sign)
}
