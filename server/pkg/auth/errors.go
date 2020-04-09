package auth

import (
	"strings"
)

const (
	// NoDataFound : NDF error message text
	NoDataFound = "not found"
)

// IsNoDataFoundErr : Check if given error is NDF
func IsNoDataFoundErr(err error) bool {
	if strings.Contains(err.Error(), NoDataFound) {
		return true
	}
	return false
}
