package fs

import (
  "io"
  "os"
  "path"
  "strings" 
  "fmt"
  "sync/atomic"
  filepath_pkg "path/filepath"
  "golang.org/x/sync/errgroup"
  "hyperflow.in/client/api_client"
  "hyperflow.in/client/utils" 
   "hyperflow.in/server/pkg/base"
  ws "hyperflow.in/server/pkg/workspace"
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
  ignoreMap map[string]bool
}

func NewRepoFs(basePath string, parallel int, repoName string, branchName string, commitId string, c *api_client.ApiClient, ignoreList []string) *RepoFs {
  
  var limit int = DefaultFileOpLimit
  ignore_map := make(map[string]bool)

  if parallel != 0 {
    limit = parallel
  }
  
  if len(ignoreList) >0 {
    for f := 0; f < len(ignoreList); f +=1 {
      ignore_map[string(ignoreList[f])] = true
    }
  }  

  return &RepoFs {
    api: c, 
    ignoreMap: ignore_map,
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

func (fs *RepoFs) RepoName() (repoName string) {
  if fs.repo != nil {
    return fs.repo.Name
  }
  return 
}

func (fs *RepoFs) BranchName() (branchName string) {
  if fs.branch != nil {
    return fs.branch.Name
  }
  return 
}

func (fs *RepoFs) CommitId() (commitId string) {
  if fs.commit != nil {
    return fs.commit.Id
  }
  return 
}

func (fs *RepoFs) GetRepoParams() (repoName, branchName, commitId string) {
  return fs.RepoName(), fs.BranchName(), fs.CommitId()
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

func (fs *RepoFs) queryCommit() error {repo_name, branch_name, commit_id := fs.GetRepoParams()
  
  commit, err := fs.api.GetCommit(repo_name, branch_name, commit_id)
  if err != nil {
    base.Debug("[RepoFs.queryCommit] Get Commit Failure: ", err)
    return err
  }

  fs.commit = commit
  return nil
}

func (fs *RepoFs) setCommit() error {
  repo_name, branch_name, commit_id := fs.GetRepoParams()
  
  commit, err := fs.api.GetOrCreateCommit(repo_name, branch_name, commit_id)
  if err != nil {
    base.Debug("[RepoFs.setCommit] Get or Create Commit Failure: ", err)
    return err
  }

  fs.commit = commit
  return nil
}

func (fs *RepoFs) Clone() (*ws.Commit, error) {

  op_limiter := NewOpLimiter(fs.concurrency)
  var repo_size int64
  var eg errgroup.Group

  if err := fs.CreateRepoDir(); err != nil {
    base.Error("[RepoFs.Clone] Failed to create repo directory. ", err)
    return nil, err
  }

  if err := fs.queryCommit(); err != nil {
    base.Error("[RepoFs.Clone] Failed to query commit from server", err)
    return nil, err
  } 

  if fs.fileMap == nil {
    if err:= fs.syncFileMap(); err != nil {
      return nil, err
    }
  }

  if len(fs.fileMap.Entries) == 0 {
    base.Debug("[RepoFs.Clone] No files in map")
    return nil, nil
  }

  for file_path, f_entry := range fs.fileMap.Entries { 
    base.Info("Downloading File:", file_path)
    op_limiter.Ask()
    
    eg.Go(func() (retError error){
        defer op_limiter.Release() 
        obj_repo_name := fs.RepoName()
        obj_commit_id := f_entry.Commit.Id
        downl_bytes, err := fs.PullObject(obj_repo_name, obj_commit_id, file_path)

        if err != nil {
          base.Error("[RepoFs.Clone] Failed to pull object: ", file_path, err)
        }

        atomic.AddInt64(&repo_size, downl_bytes) 
        return err
    })

  }
  err := eg.Wait()
  
  if err != nil {
    base.Error("[RepoFs.Clone] Failures while cloning the Repo: ", err)
    return fs.commit, err
  }

  base.Log("[RepoFs.Clone] Cloned Repo size: ", repo_size)
   
  return fs.commit, err 
}


func (fs *RepoFs) pushCode(fullFilePath string) (int64, error) {
  
  var upld_size int64
  repo_name, branch_name, commit_id := fs.GetRepoParams()

  base.Info("[RepoFs.PushObject] Pushing File: ", fullFilePath)
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
  var eg errgroup.Group

  repo_name, _, commit_id := fs.GetRepoParams()

  repo_path := fs.GetWorkingDir()

  commit_map, err := fs.api.GetCommitMap(repo_name, commit_id) 
  if err != nil {
    base.Log("[RepoFs.pushCodeUpdates] Error :", err)
    return err
  }

  if err:= filepath_pkg.Walk(repo_path, func(current_path string, file_osinfo os.FileInfo, err error) error {
    
    if len(fs.ignoreMap) > 0 { 
      sub_path := strings.Replace(current_path, repo_path , "", -1)
      sub_path_list := strings.Split(sub_path, "/")
      if len(sub_path_list) > 1 {
        if fs.ignoreMap[sub_path_list[1]] { 
          return nil
        }  
      } 
    } 

    if (file_osinfo.Mode() & file_osinfo.Mode() > os.ModeNamedPipe) {
      return nil
    }

    if (file_osinfo.Mode() & file_osinfo.Mode() > os.ModeSymlink) {
      return nil
    }

    //current_dir := filepath_pkg.Base(current_path)
    //TODO: ignore swap file  
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
  fmt.Println("Upload size (in bytes): ", upload_size)
  return fnError
}

func (fs *RepoFs) PushRepo() (commit *ws.Commit, fnError error) {

  if err := fs.setCommit(); err != nil {
    return nil, err
  }

  commit = fs.commit

  fnError =  fs.pushCodeUpdates()
  if fnError != nil {
    return
  }

  return
}

func (fs *RepoFs) PullObject(repoName, commitId, filePath string) (int64, error) {

  var bytes_wrtn int64
  repo_name := repoName
  commit_id := commitId
  file_path := filePath

  if repo_name == "" || commit_id == "" || file_path == "" {
    return 0, missingParamsError("repoName, commitId, filePath")
  }

  _, dl_stream, err := fs.api.GetFileObject(repo_name, "", commit_id, file_path)
  
  if err != nil {
    base.Log("[RepoFs.PullObject] Failed to retrieve file data: ", repo_name, commit_id, file_path)
    return 0, err  
  }

  defer dl_stream.Close()

  buf := utils.GetBuffer()
  defer utils.PutBuffer(buf)

  for {
    // bytes read 
    br, err := dl_stream.Read(buf)
    if br == 0 && err != nil {
      if err == io.EOF {
        return bytes_wrtn, nil
      }
      return bytes_wrtn, err
    }

    // bytes written in this lap
    bw, err := fs.MakeFile(file_path, func(w io.Writer) error {
      _, err:= w.Write(buf[:br])
      return err})
    
    bytes_wrtn = bytes_wrtn + bw 

    if err != nil {
      return bytes_wrtn, err
    }
  }

  base.Debug("[RepoFs.PullObject] File Name: ", file_path, " bytes written: ", bytes_wrtn)
  return bytes_wrtn, nil
}