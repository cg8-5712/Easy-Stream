package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateStreamKey 生成推流码
func GenerateStreamKey() string {
	timestamp := time.Now().Unix()
	random := make([]byte, 4)
	_, _ = rand.Read(random)
	return fmt.Sprintf("stream_%d_%s", timestamp, hex.EncodeToString(random))
}

// GenerateToken 生成随机 Token
func GenerateToken(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
