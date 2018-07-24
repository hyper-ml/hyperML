package db

import (
  "fmt"
)



// Custom error when record is not found 
type ErrRecNotFound struct {
	Type string
	Key string
}

func (e ErrRecNotFound) Error() string {
	return fmt.Sprintf("%s %s not found", e.Type, e.Key)
}

func IsErrRecNotFound(e error) bool {
  _, ok := e.(ErrRecNotFound)
  return ok
}

// Error when key exists when it shouldnt 
type ErrRecAlreadyExists struct {
  Type string
  Key  string
}

func (e ErrRecAlreadyExists) Error() string {
  return fmt.Sprintf("%s %s not found", e.Type, e.Key)
}


func IsErrRecFound(e error) bool {
  _, ok := e.(ErrRecAlreadyExists)
  return ok
}


