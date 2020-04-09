package base

import (
	"fmt"
)

const (
	// HttpEmptyRespError : Error message for empty response
	HttpEmptyRespError = "Empty Response from Server"

	// NullRepo : Error message for empty repo
	NullRepo = "Repo Not found"

	// MissingFlowAttrs : Error message for missing flow attrs
	MissingFlowAttrs = "Missing Flow attributes"

	// NullModelRepo : Error message for NUll Model Repo
	NullModelRepo = "Flow has no model repo"

	// NullOutput : Error message when flow has no output
	NullOutput = "Flow has no output"
)

// ErrHttpEmptyResponse : Raises Empty response
func ErrHttpEmptyResponse() error {
	return fmt.Errorf(HttpEmptyRespError)
}

// IsErrHttpEmptyResponse : Returns True if err is empty response
func IsErrHttpEmptyResponse(err error) bool {
	if err.Error() == HttpEmptyRespError {
		return true
	}
	return false
}

// ErrNullRepo : Raises Null Repo error
func ErrNullRepo() error {
	return fmt.Errorf(NullRepo)
}

// IsErrNullRepo : Returns true for Null Repo Error
func IsErrNullRepo(err error) bool {
	if err.Error() == NullRepo {
		return true
	}
	return false
}

// ErrMissingFlowAttrs : Raise Missing Flow Attrs Error
func ErrMissingFlowAttrs() error {
	return fmt.Errorf(MissingFlowAttrs)
}

// IsErrMissingFlowAttrs : Returns true when error is missing flow attrs
func IsErrMissingFlowAttrs(err error) bool {
	if err.Error() == MissingFlowAttrs {
		return true
	}
	return false
}

// ErrNullModelRepo : Return error struct for Empty Model Repo
func ErrNullModelRepo() error {
	return fmt.Errorf(NullModelRepo)
}

// IsErrNullModelRepo : Return true for Empty Model Repo
func IsErrNullModelRepo(err error) bool {
	if err.Error() == NullModelRepo {
		return true
	}
	return false
}

// ErrNullOutput : returns error struct for empty output
func ErrNullOutput() error {
	return fmt.Errorf(NullOutput)
}

// IsErrNullOutput : returns true for empty output
func IsErrNullOutput(err error) bool {
	if err.Error() == NullOutput {
		return true
	}
	return false
}
