package utils

import (
  "fmt"
)

const (
  HttpEmptyRespError = "Empty Response from Server"
  NullRepo = "Repo Not found"
  MissingFlowAttrs = "Missing Flow attributes"
)

func ErrHttpEmptyResponse() error{
  return fmt.Errorf(HttpEmptyRespError)
}

func IsErrHttpEmptyResponse(err error) bool {
  if err.Error() == HttpEmptyRespError {
    return true
  }
  return false
}


func ErrNullRepo() error{
  return fmt.Errorf(NullRepo)
}

func IsErrNullRepo(err error) bool {
  if err.Error() == NullRepo {
    return true
  }
  return false
}


func ErrMissingFlowAttrs() error {
  return fmt.Errorf(MissingFlowAttrs)
}

