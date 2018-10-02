package base

import (
  "fmt"
)

const (
  HttpEmptyRespError = "Empty Response from Server"
  NullRepo = "Repo Not found"
  MissingFlowAttrs = "Missing Flow attributes"
  NullModelRepo ="Flow has no model repo"
  NullOutput = "Flow has no output"
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



func ErrMissingFlowAttrs() error{
  return fmt.Errorf(MissingFlowAttrs)
}

func IsErrMissingFlowAttrs(err error) bool {
  if err.Error() == MissingFlowAttrs {
    return true
  }
  return false
}



func ErrNullModelRepo() error{
  return fmt.Errorf(NullModelRepo)
}

func IsErrNullModelRepo(err error) bool {
  if err.Error() == NullModelRepo {
    return true
  }
  return false
}



func ErrNullOutput() error{
  return fmt.Errorf(NullOutput)
}

func IsErrNullOutput(err error) bool {
  if err.Error() == NullOutput {
    return true
  }
  return false
}