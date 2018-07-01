package utils

import (
  "strings"
	"github.com/satori/go.uuid"
)

func freshUuid() string {
  uuid, _ := uuid.NewV4()
  return uuid.String()
}

func NewUUID() string {
  return strings.Replace(freshUuid(), "-", "", -1)
}