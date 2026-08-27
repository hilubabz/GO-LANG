package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() string {
	randomToken := make([]byte, 32)

	_, err := rand.Read(randomToken)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(randomToken)
}