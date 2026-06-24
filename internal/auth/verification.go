package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const verificationTokenBytes = 32

// GenerateVerificationToken 生成一个用于邮箱验证的随机令牌。
func GenerateVerificationToken() (string, error) {
	b := make([]byte, verificationTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
