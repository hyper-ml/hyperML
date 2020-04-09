package requests

import (
	"fmt"
	"strings"
)

const (
	// EmptyPasswordHash : Error Message when password is empty
	EmptyPasswordHash = "empty password hash"

	// UserAlreadyExists :
	UserAlreadyExists = "user name already exists"

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

// ErrEmptyPasswordHash : Error raised when password is empty
func ErrEmptyPasswordHash() error {
	return fmt.Errorf(EmptyPasswordHash)
}

// ErrUserNameAlreadyExists : Error raised when user name exists
func ErrUserNameAlreadyExists() error {
	return fmt.Errorf(UserAlreadyExists)
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
