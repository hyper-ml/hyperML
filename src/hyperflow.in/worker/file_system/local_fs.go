package file_system

import (
  "io"
)

type LocalFs interface {
  GetWorkingDir() string 
  MakeFile(fpath string, f func(io.Writer) error) (int64, error) 
  Writer(path string) (io.WriteCloser, error)
  Reader(path string, offset uint64, size uint64) (io.ReadCloser, error)
  Delete(path string) error
  Exists(path string) bool
}
