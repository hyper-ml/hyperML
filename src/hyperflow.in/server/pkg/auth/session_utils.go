package auth

import(
  "fmt"
 
  "github.com/dgrijalva/jwt-go"
)

var jwtClaimIssuer string = "hyperflow"

type UserClaims struct {
    User *User `json:"user"`
    jwt.StandardClaims
}

var hmacSampleSecret []byte = []byte("hfsecret")

func GenerateToken(user *User) (string) {
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

func VerifyToken(tokenString string) (*User, error) {
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