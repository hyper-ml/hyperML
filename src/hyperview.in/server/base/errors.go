package base

import (
  "fmt"
  "net/http"
)

// Custom error when file is not found 
type ErrFileNotFound struct {
  RepoName string
  CommitId string
  Fpath string
}

func (e ErrFileNotFound) Error() string {
  return fmt.Sprintf("%s %s %s not found", e.RepoName, e.CommitId, e.Fpath)
}

func IsErrFileNotFound(e error) bool {
  _, ok := e.(ErrFileNotFound)
  return ok
}


type HTTPError struct {
	Status int
  Message string
}

func (err *HTTPError) Error() string {
  return fmt.Sprintf("%d %s", err.Status, err.Message)
}

func HTTPErrorf(status int, format string, args ...interface{}) *HTTPError {
  return &HTTPError{status, fmt.Sprintf(format, args...)}
}

func ErrorToHTTPStatus(err error) (int, string) {
  Debug("[errors.ErrorToHTTPStatus] :", err)

  if err == nil {
    return 200, "OK"
  }

  return http.StatusInternalServerError, err.Error()
}