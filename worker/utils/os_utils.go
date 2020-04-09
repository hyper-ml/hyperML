package utils

import (
  "os"
  "io"
  "io/ioutil"
  filepath_pkg "path/filepath"
  "github.com/hyper-ml/hyperml/server/pkg/base"

)

func MkDirAll(path string, perm os.FileMode) error {
   if err := os.MkdirAll(path, perm); err != nil {
    return err
  }
  return nil
}

func WriteFile(path string, data []byte, perm os.FileMode) (error) {
  return ioutil.WriteFile(path, data, perm)
}

func Open(path string) (*os.File, error) {
  return os.Open(path)
}


func PathExists(path string) bool {
  if _, err := os.Stat(path); os.IsNotExist(err) {
    return false
  }
  return true
}

func IsPathEmpty(path string) bool {
  d, err := os.Open(path)
  if err != nil {
    return true
  }

  defer d.Close()
  _, err = d.Readdirnames(1)

  if err == io.EOF {
    return true
  }

  return false
}

func ListFiles(path string) error {
  files, err := ioutil.ReadDir(path)
  if err != nil {
    return err
  } else {
    base.Println("Listing files from [" + filepath_pkg.Base(path) + "]:")
    for _, f := range files {
      base.Println("  " + f.Name())
    }
  }  
  return nil
}
 