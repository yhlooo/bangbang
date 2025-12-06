package signatures

import (
	"crypto/sha256"
	"fmt"
)

// SignCert 对证书签名
func SignCert(cert []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(cert))
}
