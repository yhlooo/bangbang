package signatures

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	metav1 "github.com/yhlooo/bangbang/pkg/apis/meta/v1"
)

// HS256SignAPIObject 对 metav1.Object 签名
func HS256SignAPIObject(key Key, obj metav1.Object) error {
	// 重置被签名对象
	meta := obj.GetMeta()
	meta.Signature = ""
	meta.SignTime = time.Now()

	// 签名
	var err error
	meta.Signature, err = HS256SignObject(key, obj)
	return err
}

// HS256VerifyAPIObject 校验 metav1.Object 签名
func HS256VerifyAPIObject(key Key, obj metav1.Object, allowSince, allowUntil time.Time) error {
	meta := obj.GetMeta()
	if meta.Signature == "" {
		return ErrNoSignature
	}

	if !allowSince.IsZero() && meta.SignTime.Before(allowSince) {
		return fmt.Errorf("%w: sign time: %q (expected after %q)", ErrSignatureExpired, meta.SignTime, allowSince)
	}
	if allowUntil.IsZero() {
		// 默认不接受未来的签名
		allowUntil = time.Now()
	}
	if meta.SignTime.After(allowUntil) {
		return fmt.Errorf("%w: sign time: %q (expected before %q)", ErrInvalidSignTime, meta.SignTime, allowUntil)
	}

	signature := meta.Signature
	meta.Signature = ""
	err := HS256VerifyObject(key, signature, obj)
	meta.Signature = signature
	return err
}

// HS256VerifyObject 校验可 JSON 序列化的对象的签名
func HS256VerifyObject(key Key, expected string, obj interface{}) error {
	sign, err := HS256SignObject(key, obj)
	if err != nil {
		return fmt.Errorf("sign object error: %w", err)
	}

	if sign != expected {
		return fmt.Errorf("%w: signature: %q (expected %q)", ErrSignatureMismatch, sign, expected)
	}

	return nil
}

// HS256SignObject 对任意可 JSON 序列化的对象进行签名
func HS256SignObject(key Key, obj interface{}) (string, error) {
	// JSON 序列化
	raw, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("marshal object to json error: %w", err)
	}

	// 签名
	return HS256Sign(key, raw)
}

// HS256Sign 对数据 hmac + sha256 签名
func HS256Sign(key Key, data []byte) (string, error) {
	hash := hmac.New(sha256.New, key)
	_, err := hash.Write(data)
	if err != nil {
		return "", err
	}
	sign := hash.Sum(nil)
	signStr := fmt.Sprintf("hs256:%x", sign)

	return signStr, nil
}
