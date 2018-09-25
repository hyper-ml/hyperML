package fs

import(
  "io"
  "os"
)

func (fs *RepoFs) MakeFile(fpath string, f func(io.Writer) error) (int64, error) {
  file_path := fs.GetLocalFilePath(fpath)

  base.Debug("[RepoFs.MakeFile] Creating File to local file system: ", file_path)

  if err := os.MkdirAll(filepath_pkg.Dir(file_path), DefaultPerm); err != nil {
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