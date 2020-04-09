package utils

import (
	"github.com/satori/go.uuid"
	"strings"
)

// freshUuid :
func freshUuid() string {
	uuid := uuid.NewV4()
	return uuid.String()
}

// NewUUID :
func NewUUID() string {
	return strings.Replace(freshUuid(), "-", "", -1)
}
