package utils

import (
  "os"
)

func MkDirAll(path string, perm os.FileMode) error {
   if err := os.MkdirAll(path, perm); err != nil {
    return err
  }
  return nil
}

func Open(path string) (*os.File, error) {
  return os.Open(path)
}