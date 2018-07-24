package workspace

import (
  "io"
  "fmt"
  "golang.org/x/net/context"
  "hyperview.in/server/base"
  task_pkg "hyperview.in/server/core/tasks"
  db_pkg "hyperview.in/server/core/db"
  "hyperview.in/server/core/storage"

)

type ApiServer interface { 
  GetRepoInfo(repoName string) (*RepoInfo, error) 
  GetBranchInfo(repoName string, branchName string) (*BranchInfo, error)
  GetCommitInfo(repoName string, commitId string) (*CommitInfo, error) 
  GetFileInfo(repoName string, commitId string, filePath string) (*FileInfo, error)
  DownloadRepo(repoName string) (*RepoInfo, *BranchInfo, *CommitInfo, *FileMap, error)

  PutFile(repoName string, commitId string, fpath string, reader io.Reader) (*FileInfo, int64, error)

  CreateTask(config *task_pkg.TaskConfig) (*Task, error)
}

type apiServer struct {
	version string
  ctx context.Context
  db *db_pkg.DatabaseContext
  q *queryServer
  objWrapper *ObjWrapper 
}

func NewApiServer(db *db_pkg.DatabaseContext, oapi storage.ObjectAPIServer) (*apiServer, error) {
  return &apiServer{
    version: "0.1", 
    ctx: context.Background(),
    db: db, 
    q: &queryServer{
      db: db,
    },
    objWrapper: &ObjWrapper{
      api: oapi,
    },
  }, nil
}

func (a *apiServer) InitRepo(name string) error {
  //TODO: auth, validate repo name
    
  newRepo := &Repo {
      Name: name,
  }

  //newBranch := &Branch {Repo: newRepo, Name: "master"}

  newRepoInfo :=  &RepoInfo {
    Repo: newRepo,
  }
  
  newRepoInfo.Description = "New Repo"
  

  err := a.db.Insert(REPO_KEY_PREFIX + name, newRepoInfo)

  return err 
}

// TODO: handle errors and add validations
func (a *apiServer) RemoveRepo(name string) error {
  err := a.q.DeleteRepoInfo(name)
  // TODO: delete branches, commits etc
  return err
}

func (a *apiServer) GetRepo(name string) (*RepoInfo, error) {
  return a.q.GetRepoInfo(name)
}
 

func (a *apiServer) StartCommit(repoName string) (commitId string, err error) {
  ct, err := NewCommitTxn(repoName, "", a.db)
  if err != nil {
    return "", err
  }
  return ct.Start() 
}

func (a *apiServer) EndCommit(repoName string, commitId string) (error) {
  ct, err := NewCommitTxn(repoName, commitId, a.db)
  if err != nil {
    return err
  }
  return ct.End()
}


func (a *apiServer) CloseOpenCommit(repoName string) error {
  var branch_info *BranchInfo
  var err error 

  if !a.q.CheckRepoExists(repoName) {
    base.Log("Failed to retrieve Repo Info: %s", repoName)
    return fmt.Errorf("Failed to retrieve Repo Info: %s", repoName)
  }

  if branch_info, err = a.q.GetBranchInfo(repoName, "master"); err != nil {
    base.Log("Failed to retrieve Branch Info")
    return err
  }

  if branch_info.Head == nil {
    return fmt.Errorf("No commits in this repo")
  } 

  return a.EndCommit(repoName, branch_info.Head.Id)
}

func (a *apiServer) IsOpenCommit(repoName string, commitId string, fpath string) bool {
  ct, _ := NewCommitTxn(repoName, commitId, a.db)
  return ct.IsOpenCommit()
}
 
// TODO: putfile by hash


// put file by reader
func (a *apiServer) PutFile(repoName string, commitId string, fpath string, reader io.Reader) (*FileInfo, int64, error) {
   
  var size int64 
  var err error  

  ct, err := NewCommitTxn(repoName, commitId, a.db)
  if err != nil {
    base.Log("[apiServer.PutFile] Failed to create commit txn", err)
    return nil, 0, err
  }

  // check if commit is open 
  if ct.IsOpenCommit() == false {
    return nil, 0, fmt.Errorf("[apiServer.PutFile] PutFile requires an open commit. Params: %s %s %s", repoName, commitId, fpath)
  }

  file_info, err := a.q.GetFileInfo(repoName, commitId, fpath)
  
  if err != nil {
    if !db_pkg.IsErrRecNotFound(err) {
      base.Log("[apiServer.PutFile] Failed to retrieve File Info", repoName, commitId, fpath, err)
      return nil, 0, err
    }
  } 
  
  if file_info.Object != nil { 
    
    base.Debug("[apiServer.PutFile] Found file_info. Updating object. ", file_info.Object.Hash)
    object_hash := file_info.Object.Hash 
    
    new_hash, chksum, size, err :=  a.objWrapper.AppendObject(object_hash, reader) 
    base.Debug("[apiServer.PutFile] Object updated: ", new_hash, chksum)

    if err != nil {
      base.Log("[apiServer.PutFile] Failed to write object on server", err)
      return nil, 0, err
    } 
    
    // TODO: append should add to size or pull size from storage 
    file_info.SizeBytes = size

    //TODO: update file_info 
    return file_info, size, nil
  } 
 
  // in case object doesnt exist then add
  base.Debug("[apiServer.PutFile] This is a new file for commit. Creating storage object: ")
  new_hash, cs, size, err :=  a.objWrapper.CreateObject(reader) 
  
  if err!= nil {
    base.Log("[apiServer.PutFile] New Object creation failed: ", new_hash, cs, size, err)
    return nil, 0, err
  }

  base.Debug("[apiServer.PutFile] Storage object created. Adding file to commit map - ", new_hash)
  f_info, err := ct.AddFile(fpath, new_hash, size, cs)
  if err != nil {
    base.Log("[apiServer.PutFile] Failed to add file to commit log", fpath, new_hash, size)
    return nil, 0, err
  }

  return f_info, size, nil 
}
 

func (a *apiServer) AddFileToRepo(repoName string, path string, reader io.Reader) (int64, error) {
  // find master branch and commit 
  // raise error if no open commit 
  var size int64
  repo_info, err := a.q.GetRepoInfo(repoName)

  if err != nil {
    base.Log("Invalid Repo - %s", repoName)
    return 0, err
  }

 
  // create object in store first 
  objName, cs, size, err :=  a.objWrapper.CreateObject(reader) 
  if err != nil {
    base.Log("Failed to write object on server", err)
    return 0, err
  }

  branch_info, err := a.q.GetBranchInfo(repoName, repo_info.Branch.Name)

  // add object metadat to DB
  ct, err := NewCommitTxn(repoName, branch_info.Head.Id, a.db)
  if err != nil {
    base.Log("Failed to create commit txn", err)
    return 0, err
  }

  f_info, err := ct.AddFile(path, objName, size, cs)
  
  if err != nil {
    base.Log("Failed to create commit record", err)
    return 0, err
  }
  base.Debug("File Added to commit:", &f_info)

  return size, nil
}



func (a *apiServer) DownloadRepo(repoName string) (*RepoInfo, *BranchInfo, *CommitInfo, *FileMap, error) {
  
  repo_info, err := a.q.GetRepoInfo(repoName)
  if err != nil {
    base.Log("Invalid Repo - %s", repoName)
    return nil, nil, nil, nil, err
  }

  branch_info, err := a.q.GetBranchInfo(repoName, repo_info.Branch.Name)
  commit_head :=branch_info.Head.Id
  commit_info, err:= a.GetCommitInfo(repoName, commit_head)

  fmap, err:= a.q.GetFileMap(repoName, commit_head)

  return repo_info, branch_info, commit_info, fmap, nil

}










