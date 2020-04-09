package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// GetRandomStr : Get a random URL encoded string
func GetRandomStr(length int) (string, error) {

	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(buf), nil
}
