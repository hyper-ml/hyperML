package file_system

import (
  "io"
  "path"
  "io/ioutil"
  "os"
  "fmt"
  "sync/atomic"

  filepath_pkg "path/filepath"
  "golang.org/x/sync/errgroup"

  "hyperview.in/server/base"
  ws "hyperview.in/server/core/workspace"
  
  "hyperview.in/worker/utils"
  "hyperview.in/worker/api_client"
)

const (
  DefaultFileOpLimit = 5
  DefaultOutPath = "/out"
  DefaultOutPerm os.FileMode = 0775
)
 
type RepoFs struct {
  repo *ws.Repo
  commit *ws.Commit
  basePath string
  concurrency int
  commitMap *ws.FileMap
  workerClient *api_client.WorkerClient
}


func NewRepoFs(basePath string, concurrency int, repoName string, commitId string, wc *api_client.WorkerClient) *RepoFs {
  
  var conc_limit int = DefaultFileOpLimit

  if concurrency != 0 {
    conc_limit = concurrency
  }

  return &RepoFs {
    workerClient: wc, 
    repo: &ws.Repo {
        Name: repoName,
      },
    commit: &ws.Commit {
        Id: commitId,
      },
    basePath: basePath,
    concurrency: conc_limit,
  }

}  

func (fs *RepoFs) GetWorkingDir() string {
  return fs.basePath
}


func (fs *RepoFs) GetLocalFilePath(fileName string) string{
  return path.Join(fs.GetWorkingDir(), fileName)
}

// - Method pulls object from remote repo to local FS
// - returns bytes written and error, if any

func (fs *RepoFs) PullObject(filePath string) (int64, error) {
  // TODO: compare read with written 
  // var bytes_read int64
  var bytes_wrtn int64

  f_request := fs.workerClient.ContentIo.Get()
  f_request.Param("commitId", fs.commit.Id)
  f_request.Param("repoName", fs.repo.Name)
  f_request.Param("filePath", filePath)

  file_io, err := f_request.ReadResponse()
  defer file_io.Close()

  if err != nil {
    base.Log("[RepoFs.PullObject] Failed to fetch commit map from the server for repo: ", fs.repo.Name, fs.commit.Id)
    return 0, err  
  }

  buf := utils.GetBuffer()
  defer utils.PutBuffer(buf)

  for {
    // bytes read 
    br, err := file_io.Read(buf)
    if br == 0 && err != nil {
      if err == io.EOF {
        return bytes_wrtn, nil
      }
      return bytes_wrtn, err
    }

    // bytes written in this lap
    bw, err := fs.MakeFile(filePath, func(w io.Writer) error {
      _, err:= w.Write(buf[:br])
      return err})
    
    bytes_wrtn = bytes_wrtn + bw 

    if err != nil {
      return bytes_wrtn, err
    }
  }

  base.Debug("[RepoFs.PullObject] File Name: ", filePath, " bytes written: ", bytes_wrtn)
  return bytes_wrtn, nil
}

func (fs *RepoFs) mountRepo() (repoSize int64, fnerror error) {

  op_limiter := NewOpLimiter(fs.concurrency)
  var repo_size int64
  var eg errgroup.Group

  if fs.commitMap == nil {
    return 0, fmt.Errorf("[RepoFs.mountRepo] The local repo Fs has no knowledge of commit map. Either it is not pulled or doesn't exist on server")
  }

  for file_path, _ := range fs.commitMap.Entries { 
    base.Debug("Downloading File:", file_path)
    op_limiter.Ask()
    
    eg.Go(func() (retError error){
        defer op_limiter.Release() 
        
        downl_bytes, err := fs.PullObject(file_path)
        atomic.AddInt64(&repo_size, downl_bytes) 
        return err
    })

  }
  base.Log("[RepoFs.mountRepo] Cloned Repo size: ", repo_size)
  
  fnerror = eg.Wait()
  repoSize = repo_size

  return
}

func (fs *RepoFs) pullCommitMap() error {
  commit_map, err := fs.workerClient.FetchCommitMap(fs.repo.Name, fs.commit.Id)
  if err != nil {
    return err
  }
  fs.commitMap = commit_map

  return nil
} 

func (fs *RepoFs) Mount() error {

  // fetch commit file map
  if err := fs.pullCommitMap(); err != nil {
      base.Log("[RepoFs.Mount] Failed to pull commit map for repo commit:", fs.repo.Name, fs.commit.Id)
      base.Log(err.Error())
      return err
    }
  
  if repo_size, err := fs.mountRepo(); err != nil {
      base.Log("[RepoFs.Mount] Failed to mount repo on local filesystem:", fs.repo.Name, fs.commit.Id)
      base.Log(err.Error())
      return err
    } else {
      base.Debug("[RepoFs.Mount] Repo Download Size: ", repo_size)
    }

  if err := fs.createOutDir(); err != nil {
    base.Log("[RepoFs.Mount] Failed to create out dir: ", err)
    return err
  }  

  return nil 
}

 
func (fs *RepoFs) MakeFile(fpath string, f func(io.Writer) error) (int64, error) {
  file_path := fs.GetLocalFilePath(fpath)

  base.Debug("[RepoFs.MakeFile] Creating File to local file system: ", file_path)

  if err := os.MkdirAll(filepath_pkg.Dir(file_path), 0700); err != nil {
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

func (fs *RepoFs) Writer(path string) (io.WriteCloser, error) {
  targetPath := filepath_pkg.Join(fs.basePath, path)

  // create dir if missing
  // TODO: test this
  if err := os.MkdirAll(filepath_pkg.Dir(targetPath), 0755); err != nil {
    return nil, err
  }

  f, err := os.Create(targetPath)
  if err != nil {
    return nil, err
  }
  fmt.Println("targetPath:", targetPath)

  return f, nil
}

func (fs *RepoFs) createOutDir() error {
  outPath := filepath_pkg.Join(fs.basePath, DefaultOutPath)

  if err := os.MkdirAll(outPath, DefaultOutPerm); err != nil {
    base.Log("[RepoFs.createOutDir] Failed to create out directory: ", err)
    return err
  }
  base.Debug("[RepoFs.createOutDir] Output directory path: ", outPath)
  return nil
}



func (fs *RepoFs) PushObject(rel_path string) (int64, error) {
  var upld_size int64

  file_path := filepath_pkg.Join(fs.basePath, rel_path)
  base.Debug("[RepoFs.PushObject] Pushing File: ", file_path)

  file_io, err := os.Open(file_path)
  if err != nil {
    return upld_size, err
  } 

  buf := utils.GetBuffer()
  defer utils.PutBuffer(buf)

  //TODO: add this inside loop to send or add multi part writer 
  w, err := fs.workerClient.PutObjectWriter(fs.repo.Name, fs.commit.Id, file_path)
  defer w.Close()

  for {
        
      read_len, err := file_io.Read(buf)
      if read_len == 0 && err != nil {
        if err == io.EOF {
          return upld_size, nil
        }
        return upld_size, err
      }
      
      wrt_len, err := w.Write(buf[:read_len])

      upld_size := upld_size + int64(wrt_len)

      if err != nil {
        return upld_size, err
      }

    }

  return upld_size, nil

}

// Description: Upload files from out directory to commit map
func (fs *RepoFs) PushOutputDir() (size int64, fnError error) {
  var upload_size uint64

  repo_path := fs.basePath
  outPath := filepath_pkg.Join(fs.basePath, DefaultOutPath)

  var eg errgroup.Group

  if err:= filepath_pkg.Walk(outPath, func(current_path string, file_osinfo os.FileInfo, err error) error {
      base.Debug("[RepoFs.PushOutputDir] Found File in outpath: ", current_path)
      // ignore symlinks for this release 

      if err != nil {
        base.Log("[RepoFs.PushOutputDir] Error reading file info from os. ")
        return err
      }

      eg.Go(func() (upldError error){
          
          if current_path == outPath {
            base.Log("[RepoFs.PushOutputDir] Current path and Output path are same. Skipping.. ")
            return nil
          }

          // derive path relative to workspace directory 
          // as this will be updated in commit map
          relative_path, err := filepath_pkg.Rel(repo_path, current_path)
          if err != nil {
            base.Log("[RepoFs.PushOutputDir] Failed to find relative path of current path to repo :", current_path, repo_path)
            base.Log("[RepoFs.PushOutputDir] ", err)
            return err 
          }

          if (file_osinfo.Mode() & file_osinfo.Mode() > os.ModeNamedPipe) {
            base.Log("[RepoFs.PushOutputDir] Found Named pipe. Skipping.. ", current_path)
            return nil
          }

          if (file_osinfo.Mode() & file_osinfo.Mode() > os.ModeSymlink) {
            base.Log("[RepoFs.PushOutputDir] Found Named Symlink. Skipping.. ", current_path)
            return nil
          }

          file_len, err := fs.PushObject(relative_path)
          atomic.AddUint64(&upload_size, uint64(file_len))

          if err != nil {
            return err
          }
          return nil
          // call function upload path to commit map

        })  

      return nil
    }); err != nil {
    base.Log("[RepoFs.PushOutputDir] Failed to walk through out directory. ", err)
  }
  fnError = eg.Wait()
  size = int64(upload_size)
  return
}
 

func (fs *RepoFs) Reader(path string, offset uint64, size uint64) (io.ReadCloser, error) {
  return nil, nil
}

// TODO: remove from cloud too??
func (fs *RepoFs) Delete(path string) error {
  return os.Remove(filepath_pkg.Join(fs.basePath, path))
}

func (fs *RepoFs) Exists(path string) bool {
  return false  
}

func (fs *RepoFs) ListFiles(rel_path string, r bool) ([]os.FileInfo, error) {
  return fs.listFiles(rel_path, r, true)
}

// set r = true  for recursive search
func (fs *RepoFs) listFiles(rel_path string, r bool, printFlag bool) ([]os.FileInfo, error) {
  var search_path string = fs.basePath

  if rel_path != "" {
    if rel_path[:1] != "/" {
      search_path = search_path + "/"+ rel_path
    } else {
      search_path = search_path + rel_path
    }
  }  

  files, err := ioutil.ReadDir(search_path)
  if err != nil {
    base.Debug("[RepoFs.ListFiles] Failed to list files from :", search_path)
    return nil, err
  }

  if printFlag && len(files) > 0 {
    base.Log("[RepoFs.ListFiles] File List: ")
    for _, f := range files {
      base.Log(f.Name())
    }
  }

  return files, nil
}