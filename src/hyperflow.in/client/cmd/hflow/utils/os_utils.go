package utils

import (
  "os"
  "io/ioutil"
)

const (
  DefaultFilePerm = 0644
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

func CreateFileContent(path string, content []byte, perm uint32) error {
  if perm == 0 {
    perm = DefaultFilePerm
  }

  if err := ioutil.WriteFile(path, content, os.FileMode(perm)); err != nil {
    return err
  }
  return nil
}

func GetFileContent(path string) ([]byte, error) {
  return ioutil.ReadFile(path)
}


func DirExists(path string) bool {
  if _, err := os.Stat(path); os.IsNotExist(err) {
    return false
  }
  return true
}

func PathExists(path string) bool {
  if _, err := os.Stat(path); os.IsNotExist(err) {
    return false
  }
  return true
}