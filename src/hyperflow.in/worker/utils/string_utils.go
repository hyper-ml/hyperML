package utils

import(
  "time"
  "math/rand"
)

var randIntVar *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

const (
  charset = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func RandomStr(size int) string {
  b := make([]byte, size) 
  for i := range b {
    b[i] = charset[randIntVar.Intn(len(charset))]
  }

  return string(b)
}
