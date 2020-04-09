package osutils

import (
	"os"
)

// GetOsEnvVar :
func GetOsEnvVar(name string) string {
	return os.Getenv(name)
}
