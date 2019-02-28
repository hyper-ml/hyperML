package auth

import(
  "fmt"
  "strings"
)


const (
  NoDataFound = "not found"
  EmptyPasswordHash = "empty password hash"
)

func IsNoDataFoundErr(err error) bool {
  if strings.Contains(err.Error(), NoDataFound) {
    return true
  }
  return false
}


func ErrEmptyPasswordHash() error {
  return fmt.Errorf(EmptyPasswordHash)
}