package base

import (
  "fmt"
  "net/http"
)


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
  if err == nil {
    return 200, "OK"
  }

  return http.StatusInternalServerError, err.Error()
}