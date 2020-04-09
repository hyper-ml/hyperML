package auth

import (
	"github.com/hyper-ml/hyperml/server/pkg/types"
)

// LoginRequest :
type LoginRequest struct {
	UserName string
	Password string
}

// LoginResponse :
type LoginResponse struct {
	Jwt       string
	UserAttrs *types.UserAttrs
}
