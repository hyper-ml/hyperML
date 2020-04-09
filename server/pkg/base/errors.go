package base

import (
	"fmt"
	"net/http"
	"strconv"
)

// Errorf : Prints Error
func Errorf(format string, args ...interface{}) {
	fmt.Println(format, args)
}

// ErrFileNotFound : Custom error when file is not found
type ErrFileNotFound struct {
	RepoName string
	CommitId string
	Fpath    string
}

func (e ErrFileNotFound) Error() string {
	return fmt.Sprintf("%s %s %s not found", e.RepoName, e.CommitId, e.Fpath)
}

// IsErrFileNotFound : Identifies file not found error
func IsErrFileNotFound(e error) bool {
	_, ok := e.(ErrFileNotFound)
	return ok
}

// HTTPError : HTTP Error struct returned by handler methods
type HTTPError struct {
	Status  int
	Message string
}

func (err *HTTPError) Error() string {
	return fmt.Sprintf("%d %s", err.Status, err.Message)
}

// HTTPErrorf :
func HTTPErrorf(status int, format string, args ...interface{}) *HTTPError {
	return &HTTPError{status, fmt.Sprintf(format, args...)}
}

// ErrorToHTTPStatus :
func ErrorToHTTPStatus(err error) (int, string) {
	if err == nil {
		return 200, "OK"
	}

	var errString = err.Error()
	if errString == "" {
		return http.StatusInternalServerError, "Unknown Error"
	}

	status, convErr := strconv.Atoi(errString[0:3])

	if convErr != nil {
		return http.StatusBadRequest, errString
	}

	return status, errString[4:]
}
