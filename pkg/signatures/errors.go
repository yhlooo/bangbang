package signatures

import "errors"

var (
	// ErrNoSignature 没有签名
	ErrNoSignature = errors.New("NoSignature")
	// ErrSignatureExpired 签名过期了
	ErrSignatureExpired = errors.New("SignatureExpired")
	// ErrInvalidSignTime 非法的签名时间
	ErrInvalidSignTime = errors.New("InvalidSignTime")
	// ErrSignatureMismatch 签名不匹配
	ErrSignatureMismatch = errors.New("SignatureMismatch")
)
