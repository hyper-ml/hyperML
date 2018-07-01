package utils

import (
  "hash"
  "crypto/sha512"
  "encoding/hex"
)

//hash for check sums
func NewHash() hash.Hash {
  return sha512.New()
}


func HexToString(bytes []byte) string {
	return hex.EncodeToString(bytes)
}