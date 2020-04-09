package auth

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"github.com/hyper-ml/hyperml/server/pkg/types"
)

func validatePassword(userType types.UserType, plaintxt string) error {
	if plaintxt == "" {
		return fmt.Errorf("Password can not be empty")
	}

	return nil
}

func hashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)

	if err != nil {
		return false
	}

	return true
}
