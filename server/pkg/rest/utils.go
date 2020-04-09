package rest

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/auth"
	"github.com/hyper-ml/hyperml/server/pkg/types"
	"net/http"
	"strings"
)

func getUserFromReq(req *http.Request) (*types.User, error) {
	jwt := req.Header.Get("Authorization")
	if jwt == "" {
		return nil, fmt.Errorf("Invalid user")
	}

	jwt = strings.TrimPrefix(jwt, "Bearer ")
	return getUserFromToken(jwt)
}

func getUserFromToken(token string) (*types.User, error) {
	user, err := auth.VerifyToken(token)
	if user == nil {
		return nil, fmt.Errorf("Invalid User")
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}
