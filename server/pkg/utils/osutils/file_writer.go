package osutils

import (
	"io/ioutil"
	"path/filepath"
)

// SaveFile : Write file at agiven location
func SaveFile(location string, data []byte) error {
	absPath := filepath.Join(AppHome(), location)
	return ioutil.WriteFile(absPath, data, DefaultFilePerm)
}

// ReadFile : Read file from a given location
func ReadFile(location string) ([]byte, error) {
	absPath := filepath.Join(AppHome(), location)
	return ioutil.ReadFile(absPath)
}
