package fs

import (
  "io"
  "path"
  "os" 
  "sync/atomic"
  filepath_pkg "path/filepath"
  "golang.org/x/sync/errgroup"
  "hyperview.in/client/api_client"
  "hyperview.in/client/utils" 
   "hyperview.in/server/base"
  ws "hyperview.in/server/core/workspace"
)


const (
  DefaultFileOpLimit = 5
  DefaultPerm os.FileMode = 0775

  DefaultModelPath = "/saved_models"
  DefaultModelPerm os.FileMode = 0775

  DefaultOutPath = "/out"
  DefaultOutPerm os.FileMode = 0775


  DefaultDataPath = "/data"
  DefaultDataPerm os.FileMode = 0775  
)
 
type RepoFs struct {
  repo *ws.Repo
  branch *ws.Branch
  commit *ws.Commit
  basePath string
  concurrency int
  fileMap *ws.FileMap
  api *api_client.ApiClient
}

func NewRepoFs(basePath string, parallel int, repoName string, branchName string, commitId string, c *client.ApiClient) *RepoFs {
  
  var limit int = DefaultFileOpLimit

  if parallel != 0 {
    limit = parallel
  }

  return &RepoFs {
    api: c, 
    repo: &ws.Repo {
        Name: repoName,
      },
    branch: &ws.Branch {
      Name: branchName,
    },
    commit: &ws.Commit {
        Id: commitId,
      },
    basePath: basePath,
    concurrency: limit,
  }

}  
 

func (fs *RepoFs) CreateRepoDir() error {

  repo_dir := fs.basePath 
  if err := os.MkdirAll(repo_dir, DefaultPerm); err != nil {
    base.Log("[RepoFs.CreateRepoDir] Failed to create repo directory: ", err)
    return err
  }
  base.Debug("[RepoFs.CreateRepoDir] Repo directory path: ", repo_dir)
  return nil
} 

func (fs *RepoFs) GetWorkingDir() string {
  return fs.basePath
}


func (fs *RepoFs) GetLocalFilePath(fileName string) string{
  return path.Join(fs.GetWorkingDir(), fileName)
}


func (fs *RepoFs) SwitchBranch(name string) error {
  return nil
} 

func (fs *RepoFs) syncFileMap() error {
  f_map, err := fs.api.GetCommitMap(fs.repo.Name, fs.commit.Id)
  if err != nil {
    base.Log("[RepoFs.syncFileMap] Failed to sync file map for this repo: ", fs.repo.Name, fs.commit.Id)
    return err
  }
  fs.fileMap = f_map
  return nil
}

func (fs *RepoFs) Clone() error {

  op_limiter := NewOpLimiter(fs.concurrency)
  var repo_size int64
  var eg errgroup.Group

  if err := fs.CreateRepoDir(); err != nil {
    base.Error("[RepoFs.Clone] Failed to create repo directory. ", err)
    return err
  }

  if fs.fileMap == nil {
    if err:= fs.syncFileMap(); err != nil {
      return err
    }
  }

  if len(fs.fileMap.Entries) == 0 {
    base.Debug("[RepoFs.Clone] No files in map")
    return nil
  }

  for file_path, _ := range fs.fileMap.Entries { 
    base.Info("Downloading File:", file_path)
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


func (fs *RepoFs) pushCode(fullFilePath string) (int64, error) {
  
  var upld_size int64
  repo_name := fs.repo.Name
  branch_name:= fs.branch.Name
  commit_id := fs.commit.Id

  base.Debug("[RepoFs.PushObject] Pushing File: ", fullFilePath)
  repo_path:= fs.GetWorkingDir()
  rel_path, err := filepath_pkg.Rel(repo_path, fullFilePath)

  file_io, err := os.Open(fullFilePath)
  if err != nil {
    return upld_size, err
  } 

  buf := utils.GetBuffer()
  defer utils.PutBuffer(buf)

  //TODO: add this inside loop to send or add multi part writer 
  w, err := fs.api.PutObjectWriter(repo_name, branch_name, commit_id, rel_path)
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
    upld_size = upld_size + int64(wrt_len)
    if err != nil {
        return upld_size, err
    }
  }

  return upld_size, nil
}

func (fs *RepoFs) pushCodeUpdates() error {
  var upload_size uint64
  var file_len int64
  var eg errgroup.Group

  repo_name := fs.repo.Name
  branch_name := fs.branch.Name
  commit_id := fs.commit.Id

  repo_path := fs.GetWorkingDir()

  commit_map, err := fs.api.fetchCommitMap(repo_name, commit_id) 
  if err != nil {
    base.Log("[RepoFs.pushCodeUpdates] Error :", err)
    return err
  }

  if err:= filepath_pkg.Walk(repo_path, func(current_path string, file_osinfo os.FileInfo, err error) error {

    if (file_osinfo.Mode() & file_osinfo.Mode() > os.ModeNamedPipe) {
      base.Log("[RepoFs.pushCodeUpdates] Found Named pipe. Skipping.. ", current_path)
      return nil
    }

    if (file_osinfo.Mode() & file_osinfo.Mode() > os.ModeSymlink) {
      base.Log("[RepoFs.pushCodeUpdates] Found Named Symlink. Skipping.. ", current_path)
      return nil
    }

    file_commit_info, ok := commit_map.Entries[current_path]
    _ = file_commit_info

    if !ok {

      // file doesnt exist. Push
      eg.Go(func() (upldError error){  
        bytes_wrt, err := fs.pushCode(current_path) 
        if err != nil {
          return err
        }

        atomic.AddUint64(&upload_size, uint64(bytes_wrt))
        return nil   
      })    
       
    }

    //TODO: if (file_commit_info.size == file_osinfo.Size()) {
      // file not changed. Skip the file 
    //  return nil
    //}

    // update code 
    eg.Go(func() (upldError error){  
      bytes_wrt, err := fs.pushCode(current_path)
      if err != nil {
        return err
      }
      atomic.AddUint64(&upload_size, uint64(bytes_wrt))
      return nil
    })
    return nil 

  }); err != nil {
    base.Log("[RepoFs.pushCodeUpdates] Walkthrough completed with an error: ", err)
  }

  fnError := eg.Wait()
  // TODO: upload size 
  base.Log("[RepoFs.pushCodeUpdates] Upload size: ", upload_size)
  return fnError
}

func (fs *RepoFs) PushRepo() (commit *ws.Commit, fnError error) {
  repo_name := fs.repo.Name
  branch_name := fs.branch.Name
  commit_id := fs.commit.Id

  commit, err = fs.api.GetOrCreateCommit(repo_name, branch_name, commit_id)
  if err != nil {
    return commit, err
  }
  fs.commit = commit 

  err =  fs.pushCodeUpdates()
  if err != nil {
    return commit, err
  }

  return commit, nil
}

func (fs *RepoFs) PullObject(filePath string) (int64, error) {

  var bytes_wrtn int64
  if fs.repoName == "" || fs.commitId == "" || filePath == "" {
    return 0, fmt.Errorf("[RepoFs.PullObject] Either one or more parameters are missing: repoName, commitId, filePath")
  }

  downloader, err := fs.api.GetFileObject(fs.repoName, fs.branchName, fs.commitId, filePath)
  if err != nil {
    base.Log("[RepoFs.PullObject] Failed to retrieve file data: ", fs.repo.Name, fs.commit.Id, filePath)
    return 0, err  
  }

  defer downloader.Close()

  buf := utils.GetBuffer()
  defer utils.PutBuffer(buf)

  for {
    // bytes read 
    br, err := downloader.Read(buf)
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