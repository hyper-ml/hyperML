package base

import (
	"fmt"
	"os"
)

// SetEnvVar : Set OS Env Vars
func SetEnvVar(name, value string) error {
	os.Setenv(name, value)
	fmt.Println("Environment value: ", name, os.Getenv(name))

	return nil
}

// GetEnvVar : Get OS Env Var
func GetEnvVar(name string) string {
	return os.Getenv(name)
}

// GetEnv : Get Env var
func GetEnv(name string) string {
	return os.Getenv(name)
}
