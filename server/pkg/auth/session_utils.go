package auth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/hyper-ml/hyperml/server/pkg/types"
)

var jwtClaimIssuer string = "hyperflow"

// UserClaims : Session tokens
type UserClaims struct {
	User *types.User `json:"user"`
	jwt.StandardClaims
}

var hmacSampleSecret []byte = []byte("hfsecret")

// GenerateToken :
func GenerateToken(user *types.User) string {
	claims := UserClaims{
		user,
		jwt.StandardClaims{
			//ExpiresAt: 15000000,
			Issuer: jwtClaimIssuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, _ := token.SignedString(hmacSampleSecret)
	return tokenString
}

// VerifyToken : check token validity
func VerifyToken(tokenString string) (*types.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return hmacSampleSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims.User, nil
	}

	return nil, fmt.Errorf("Unknown error")
}
