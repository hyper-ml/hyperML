package qs

import (
	"fmt"
	"github.com/hyper-ml/hyperml/server/pkg/db"
	"strings"
)

const (

	// UserDiskAlreadyExists :
	UserDiskAlreadyExists = "user disk already exists"

	// EmptyDiskName :
	EmptyDiskName = "user disk name is empty"

	// ZeroSizeDisk :
	ZeroSizeDisk = "user disk size is zero"

	// UserDoesntExist :
	UserDoesntExist = "user doesnt exist"

	// UserDiskDoesntExist :
	UserDiskDoesntExist = "user disk doesnt exist"
)

// IsErrRecNotFound : Custom error returned Database when key doesnt exist
func IsErrRecNotFound(e error) bool {
	return db.IsErrRecNotFound(e)
}

// ErrDiskNameNull : PersistentDisk has no name
func ErrDiskNameNull() error {
	return fmt.Errorf(EmptyDiskName)
}

// ErrDiskSizeZero :PersistentDisk has no size
func ErrDiskSizeZero() error {
	return fmt.Errorf(ZeroSizeDisk)
}

// ErrUserDoesntExist : User does nt exist
func ErrUserDoesntExist() error {
	return fmt.Errorf(UserDoesntExist)
}

// ErrUserDiskAlreadyExists : Error raised when user name exists
func ErrUserDiskAlreadyExists() error {
	return fmt.Errorf(UserDiskAlreadyExists)
}

// ErrUserDiskDoesntExist : Error raised when user disk doesnt exist in metadata
func ErrUserDiskDoesntExist() error {
	return fmt.Errorf(UserDiskDoesntExist)
}

// IsUserDiskDoesntExist :
func IsUserDiskDoesntExist(e error) bool {
	if strings.Contains(e.Error(), UserDiskDoesntExist) {
		return true
	}
	return false
}
