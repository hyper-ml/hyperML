package db

import (
	"fmt"
)

// ErrRecNotFound : Custom error when record is not found
type ErrRecNotFound struct {
	Type string
	Key  string
}

func (e ErrRecNotFound) Error() string {
	return fmt.Sprintf("%s %s not found", e.Type, e.Key)
}

// IsErrRecNotFound :
func IsErrRecNotFound(e error) bool {
	_, ok := e.(ErrRecNotFound)
	return ok
}

// ErrRecAlreadyExists : Error when key exists when it shouldnt
type ErrRecAlreadyExists struct {
	Type string
	Key  string
}

func (e ErrRecAlreadyExists) Error() string {
	return fmt.Sprintf("%s %s not found", e.Type, e.Key)
}

// IsErrRecFound :
func IsErrRecFound(e error) bool {
	_, ok := e.(ErrRecAlreadyExists)
	return ok
}
