package auth

import (
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"
)

// GenerateCode 生成6位随机验证码
func GenerateCode() (code string) {
	return fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}

// GenerateState 生成随机token字符
func GenerateState() string {
	b := make([]byte, 32)
	_, err := crand.Read(b)
	if err != nil {
		return "default_state"
	}
	return base64.StdEncoding.EncodeToString(b)
}
