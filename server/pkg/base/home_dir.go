package base

import (
	"os"
)

// JustHomeDir : Returns user home directory. Works better on Linux
func JustHomeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

// HomeDir : Returns home directory and Error if not found
func HomeDir() (string, error) {
	return os.UserHomeDir()
}
