package file_system

import (
  "io"
  "os"
  "fmt"
  "path"
  "sync"
  "time"
  "io/ioutil"
  "sync/atomic"


  filepath_pkg "path/filepath"
  "golang.org/x/sync/errgroup"

  "hyperflow.in/server/pkg/base"
  ws "hyperflow.in/server/pkg/workspace"
  
  "hyperflow.in/worker/utils"
  "hyperflow.in/worker/api_client"
)

const (
  RepoSyncTicker = 900 // secs 
  DefaultFileOpLimit = 5
  DefaultRepoDirPerm os.FileMode = 0775

  DefaultModelPath = "/saved_models"
  DefaultModelPerm os.FileMode = 0775

  DefaultOutPath = "/out"
  DefaultOutPerm os.FileMode = 0775

  UserPath = "/wh_data"
)

type FileSyncMode string
const (
  SyncModeD FileSyncMode = "DIRECT"
  SyncModeTS FileSyncMode = "THROUGH_SERVER"
)
 
type RepoFs struct {
  repo *ws.Repo
  branch *ws.Branch
  commit *ws.Commit

  basePath string
  concurrency int
  lastSyncLock sync.Mutex
  lastSync time.Time

  commitMap *ws.FileMap
  wc *api_client.WorkerClient
  pushFileMode FileSyncMode
  pullFileMode FileSyncMode
}


func NewRepoFs(basePath string, concurrency int, repoName string, branchName string, commitId string, wc *api_client.WorkerClient) *RepoFs {
  
  var conc_limit int = DefaultFileOpLimit

  if concurrency != 0 {
    conc_limit = concurrency
  }

  return &RepoFs {
    wc: wc, 
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
    concurrency: conc_limit,
    pushFileMode: SyncModeD,
    pullFileMode: SyncModeD,
  } 
}  

// Method to get base directory of current repo object

func (fs *RepoFs) GetWorkingDir() string {
  return fs.basePath
}

// Method takes a subpath and returns a full repo path

func (fs *RepoFs) getAbsolutePath(subpath string) string {
  return filepath_pkg.Join(fs.basePath, subpath)
}

// Method to get current repo Id

func (fs *RepoFs) GetRepoId() string {
  return fs.repo.Name
}

// Method to get current commit Id

func (fs *RepoFs) GetCommitId() string {
  return fs.commit.Id
}

// Generate local file path. Useful when pulling or pushing 
// the object file

func (fs *RepoFs) GetLocalFilePath(fileName string) string{
  return path.Join(fs.GetWorkingDir(), fileName)
}

// Print list of files in current repo directory

func (fs *RepoFs) PrintFileList(subpath string) error {
  
  path := filepath_pkg.Join(fs.basePath + subpath)
  if utils.PathExists(path) {
    return utils.ListFiles(path)
  }
  return nil
}

// Mount files from remote to local file system
// 
func (fs *RepoFs) Mount() error {

  if err := fs.makeRepoPathIfNotExists(); err != nil {
    base.Error("[RepoFs.Mount] Failed to create workspae dir: ", err)
    return err
  }   
  
  // fetch commit file map
  if err := fs.getCommitMap(); err != nil {
      base.Error("[RepoFs.Mount] Failed to pull commit map for repo commit:", fs.repo.Name, fs.commit.Id)
      base.Error(err.Error())
      return err
    }

    
  if repo_size, err := fs.mountRepo(); err != nil {
      base.Error("[RepoFs.Mount] Failed to mount repo on local filesystem:", fs.repo.Name, fs.commit.Id)
      base.Error(err.Error())
      return err
    } else {
      base.Println("Repo Size (in bytes): ", repo_size)
    }

  /*if err := fs.createOutDir(); err != nil {
    base.Log("[RepoFs.Mount] Failed to create out dir: ", err)
    return err
  }   

  if err := fs.createSavedModelsDir(); err != nil {
    base.Log("[RepoFs.Mount] Failed to create saved_model dir: ", err)
    return err
  } */

  return nil 
}

func (fs *RepoFs) PushModelDir() (size int64, fnError error) {
  return fs.PushDir("")
}

// Description: Upload files from out directory to commit map
func (fs *RepoFs) PushOutputDir() (size int64, fnError error) {
  return fs.PushDir("")
}

func (fs *RepoFs) PushDir(subpath string) (size int64, fnError error) {
  return fs.pushDirUpdates(subpath)
}

// private method that mounts repo on to local system

func (fs *RepoFs) mountRepo() (repoSize int64, fnerror error) {

  op_limiter := NewOpLimiter(fs.concurrency)
  var repo_size int64
  var eg errgroup.Group

  if fs.commitMap == nil {
    return 0, fmt.Errorf("[RepoFs.mountRepo] The local repo Fs has no knowledge of commit map. Either it is not pulled or doesn't exist on server")
  }

  for file_path, _ := range fs.commitMap.Entries {  
    
    op_limiter.Ask()
    file_path := file_path
    eg.Go(func() (retError error){
        defer op_limiter.Release()  
        downl_bytes, err := fs.pullObject(file_path)

        atomic.AddInt64(&repo_size, downl_bytes) 
        return err
    })

  }
  
  fnerror = eg.Wait()
  repoSize = repo_size

  return
}

// Method pulls object from remote repo to local FS
// returns bytes written and error, if any
// TODO: compare read with written 

func (fs *RepoFs) pullObject(filePath string) (int64, error) {
  
  // written bytes
  var w_size int64

  f_request := fs.wc.ContentIo.Get()
  f_request.Param("commitId", fs.commit.Id)
  f_request.Param("repoName", fs.repo.Name)
  f_request.Param("filePath", filePath)

  file_io, err := f_request.ReadResponse()
  defer file_io.Close()

  if err != nil {
    return 0, fmt.Errorf("failed to file %s: %v", filePath, err)  
  }

  buf := utils.GetBuffer()
  defer utils.PutBuffer(buf)

  for {

    w_size = w_size
    n, err := file_io.Read(buf)
    
    if n == 0 && err != nil {
        if err == io.EOF {
          return w_size, nil
        }
        return w_size, err
      }

    // bytes written in this lap
    bw, err := fs.touch(filePath, func(w io.Writer) error {
      _, err:= w.Write(buf[:n])
      return err})
    
    if err != nil {
        return w_size, err
      } else {
        w_size = w_size + bw 
      }
  }

  return w_size, nil
}

func (fs *RepoFs) getCommitMap() error {
  commit_map, err := fs.wc.FetchCommitMap(fs.repo.Name, fs.commit.Id)
  if err != nil {
    return err
  }
  fs.commitMap = commit_map
  return nil
} 


// Method creates a file at given location and 
// returns a writer
func (fs *RepoFs) touch(fpath string, f func(io.Writer) error) (int64, error) {
  file_path := fs.GetLocalFilePath(fpath)

  if err := os.MkdirAll(filepath_pkg.Dir(file_path), 0700); err != nil {
    return 0, err
  }

  file, err := os.Create(file_path)
  if err != nil {
    return 0, err
  }

  defer func() {
    if err = file.Close(); err != nil{
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

func (fs *RepoFs) makeRepoPathIfNotExists() error { 
  if err := os.MkdirAll(fs.basePath, DefaultRepoDirPerm); err != nil {
    base.Error("[RepoFs.createRepoPathIfNotExists] Failed to create repo path %s: %v", fs.basePath, err)
    return err
  }
  return nil
}


// Method pushes updated files to server. also expects 
// an open commit to work with.
// TODO: infinite backoff for file xfer

func (fs *RepoFs) pushDirUpdates(subpath string) (push_bytes int64, fnError error) {

  repo_path := fs.basePath
  push_path := fs.basePath
  
  if subpath != "" {
    push_path = filepath_pkg.Join(fs.basePath, subpath)
  }
  
  if err := fs.CheckCommit(); err != nil {
    base.Error("[RepoFs.PushDir] Commit Error : ", fs.commit.Id, err)
    return 0, err
  }

  var eg errgroup.Group
  push_map := make(map[string]time.Time)
  size_map := make(map[string]int64)

  if err := filepath_pkg.Walk(push_path, func(current_path string, file_osinfo os.FileInfo, err error) error {
      
      // skip root
      if (current_path == push_path) {
        return nil
      }
      
      //skip sub-directories    
      if file_osinfo.IsDir() {
        return nil
      }

      eg.Go(func() (error){
          
          // derive path relative to workspace directory 
          // as this will be updated in commit map
          relative_path, err := filepath_pkg.Rel(repo_path, current_path)
          if err != nil {
            base.Error("[RepoFs.PushDir] Failed to find relative path of current path to repo :", current_path, repo_path)
            base.Error("[RepoFs.PushDir] ", err)
            return err 
          }

          if (file_osinfo.Mode() & file_osinfo.Mode() > os.ModeNamedPipe) {
            return nil
          }

          if (file_osinfo.Mode() & file_osinfo.Mode() > os.ModeSymlink) {
            return nil
          }
          
          if (file_osinfo.ModTime().Before(fs.lastSync) || file_osinfo.ModTime().Equal(fs.lastSync)) {
            return nil // skip
          }  

          base.Println("Checking in " + relative_path, file_osinfo.Size())
          var file_len int64

          switch fs.pushFileMode {
            case SyncModeTS: 
              file_len, err = fs.pushObject(relative_path)
              
            case SyncModeD:
              file_len, err = fs.pushObjectByURL(relative_path) 

          }
                    
          push_map[relative_path] = file_osinfo.ModTime()
          size_map[relative_path] = file_len

          if err != nil {
            //todo
            return err
          }

          return nil

        })  

      return nil
    }); err != nil {
    return 0, fmt.Errorf("failed during pushDir(), err: %v", err) 
  }

  // wait for push to complete
  fnError = eg.Wait()
  
  if fnError != nil {
    // todo : discard partially sent files
    return 0, fmt.Errorf ("failed during push directory, err: %v", fnError)
  }
  
  // update last sync as recent modified file time 
  var max_modtime time.Time
  for _, mod_time := range push_map {
    if mod_time.After(max_modtime) {
      max_modtime = mod_time
    } 
  }

  if max_modtime.After(fs.lastSync) {
    fs.RecordSyncTime(max_modtime)
  }

  // update upload size by adding up size_map
  for _, s := range size_map {
    push_bytes = push_bytes
    push_bytes = push_bytes + s 
  }

  return 
}
 
func (fs *RepoFs) pushObject(rel_path string) (int64, error) {
  
  // bytes sent 
  var sent int64

  file_path := filepath_pkg.Join(fs.basePath, rel_path)
  file_io, err := os.Open(file_path)
  if err != nil {
    return sent, err
  } 

  buf := utils.GetBuffer()
  defer utils.PutBuffer(buf)

  w, err := fs.wc.PutFileWriter(fs.repo.Name, fs.commit.Id, file_path)
  defer w.Close()

  for {
    n, err := file_io.Read(buf)

    if n == 0 && err != nil {
      if err == io.EOF {
        return sent, nil
      }
      return sent, err
    }
    
    _, err = w.Write(buf[:n])
    sent := sent + int64(n)

    if err != nil {
      return sent, err
    }

  }

  return sent, nil
}

func (fs *RepoFs) CloseCommit() error {
  repo_name := fs.repo.Name
  branch_name := fs.branch.Name
  commit_id := fs.commit.Id

  if err := fs.wc.CloseCommit(repo_name, branch_name, commit_id);  err != nil {
    return err
  } 

  return nil
}

func (fs *RepoFs) CheckCommit() error {
  if fs.repo.Name == "" {
    return errInvalidRepo()
  }

  commit_attrs, err := fs.wc.FetchCommitAttrs(
                        fs.repo.Name, 
                        fs.commit.Id)
  if err != nil {
    return err
  }

  if !commit_attrs.IsOpen() {
    return errCommitStatus("Closed")
  }

  return nil
}

 
func (fs *RepoFs) Reader(path string, offset uint64, size uint64) (io.ReadCloser, error) {
  return nil, fmt.Errorf("unimplemeted feature")
}

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
    return nil, err
  }

  if printFlag && len(files) > 0 {
    for _, f := range files {
      base.Println(" " + f.Name())
    }
  }

  return files, nil
}

// updates sync time to most recent pushed file timestamp
//

func (fs *RepoFs) RecordSyncTime(last_mod time.Time) {
  fs.lastSyncLock.Lock()
  fs.lastSync = last_mod  
  fs.lastSyncLock.Unlock()
}

func SyncRepo(fs *RepoFs, done chan int) error {
  ticker := time.NewTicker(RepoSyncTicker * time.Second)
  go func() {
    for {
      select {
      case <- ticker.C:
                
        // Check if any file exists
        // If so then send a new commit 
        if !utils.IsPathEmpty(fs.GetWorkingDir()) {
          
          push_size, err := fs.PushDir("")
          if err != nil {
            base.Error("Failed to push directory: ", fs.GetWorkingDir(), err)
            return 
          }

          base.Println("Sent (bytes): ", push_size)
        } else {
          break
        }

      case val := <-done:
        if val == 1 {

          ticker.Stop()
          base.Println("sending directory: ", fs.GetWorkingDir())
          push_size, err := fs.PushDir("")
          if err != nil {
            base.Error("Failed to push directory: ", fs.GetWorkingDir(), err)
            return
          }
          base.Println("Push size (bytes): ", push_size)
          return
        }

      }
    }
  }()

  return nil
}