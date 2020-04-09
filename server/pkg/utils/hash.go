package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"hash"
)

//NewHash : hash for check sums
func NewHash() hash.Hash {
	return sha512.New()
}

// HexToString :
func HexToString(bytes []byte) string {
	return hex.EncodeToString(bytes)
}
