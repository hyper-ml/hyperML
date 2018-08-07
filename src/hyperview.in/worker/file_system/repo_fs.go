package file_system

import (
  "io"
  "path"
  "path/filepath"
  "os"
  "fmt"
  "hyperview.in/worker"
)
 
type repoFs struct {
  repoName string
  commitId string
  dirPath string
  concurrency int
  workerClient *worker.WorkerClient
}


func NewRepoFs(dir string, concurrency int, repoName string, commitId string, wc *worker.WorkerClient) *repoFs {

  return &repoFs {
    workerClient: wc, 
    repoName: repoName,
    commitId: commitId,
    dirPath: dir,
    concurrency: concurrency,
  }
}  

func (fs *repoFs) GetWorkingDir() string {
  return fs.dirPath
}


func (fs *repoFs) GetLocalFilePath(fileName string) string{
  return path.Join(fs.GetWorkingDir(), fileName)
}

func (fs *repoFs) Mount() error {
  // fetch file map
  commit_map, err := fs.workerClient.FetchCommitMap(fs.repoName, fs.commitId)
  if err != nil {
    return err
  }

  fmt.Println("commit_map:", commit_map)

  return nil 
  // download one by one to directory 
}

 
func (fs *repoFs) MakeFile(fpath string, f func(io.Writer) error) (int64, error) {
  file_path := fs.GetLocalFilePath(fpath)

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

func (fs *repoFs) Writer(path string) (io.WriteCloser, error) {
  targetPath := filepath.Join(fs.dirPath, path)

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
  

func (fs *repoFs) Reader(path string, offset uint64, size uint64) (io.ReadCloser, error) {
  return nil, nil
}

func (fs *repoFs) Delete(path string) error {
  return os.Remove(filepath.Join(fs.dirPath, path))
}

func (fs *repoFs) Exists(path string) bool {
  return false  
}

