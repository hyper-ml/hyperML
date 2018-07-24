package fs

import (
  "io"
  "path"
  "path/filepath"
  "os"
  "fmt"
)

type FS interface {
  GetWorkingDir() string 
  MakeFile(fpath string, f func(io.Writer) error) (int64, error) 
  Writer(path string) (io.WriteCloser, error)
  Reader(path string, offset uint64, size uint64) (io.ReadCloser, error)
  Delete(path string) error
  Exists(path string) bool
}


type RepoFS struct {
  // repo or project directory    
  dir string

  //to control concurrent read or wrtiers  
  concurrency int
  //TODO: add stats   
}

type CountWriter struct {
  w io.Writer
  size int64
}

func (c *CountWriter) Write(b []byte) (int, error) {
  n, err := c.w.Write(b)
  c.size += int64(n)
  return n, err
}


func NewFS(dir string, concurrency int) *RepoFS {
  return &RepoFS {
    dir: dir,
    concurrency: concurrency,
  }
}


func (fs *RepoFS) GetWorkingDir() string {
  return fs.dir
}

func (fs *RepoFS) GetLocalPath(fileName string) string{
  return path.Join(fs.GetWorkingDir(), fileName)
}

func (fs *RepoFS) MakeFile(fpath string, f func(io.Writer) error) (int64, error) {
  file_path := fs.GetLocalPath(fpath)

  if err := os.MkdirAll(filepath.Dir(file_path), 0700); err != nil {
    return 0, err
  }

  file, err := os.Create(file_path)
  if err != nil {
    return 0, err
  }

  defer func() {
    if err = file.Close(); err != nil{
      fmt.Println("Error closing file", err)
      return
    }
  }()

  w := &CountWriter{w: file}
  if err := f(w); err != nil {
    return 0, err
  }

  return w.size, nil
}

func (fs *RepoFS) Writer(path string) (io.WriteCloser, error) {
  targetPath := filepath.Join(fs.dir, path)

  // create dir if missing
  // TODO: test this
  if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
    return nil, err
  }

  f, err := os.Create(targetPath)
  if err != nil {
    return nil, err
  }
  fmt.Println("targetPath:", targetPath)

  return f, nil
}

func (fs *RepoFS) Reader(path string, offset uint64, size uint64) (io.ReadCloser, error) {
  return nil, nil
}

func (fs *RepoFS) Delete(path string) error {
  return os.Remove(filepath.Join(fs.dir, path))
}

func (fs *RepoFS) Exists(path string) bool {
  return false  
}