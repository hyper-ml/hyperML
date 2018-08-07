package workspace

import (
  "io"
  "fmt"
  "golang.org/x/net/context"
  "hyperview.in/server/base"
  //task_pkg "hyperview.in/server/core/tasks"
  db_pkg "hyperview.in/server/core/db"
  "hyperview.in/server/core/storage"

)

type ApiServer interface { 
  GetRepoAttrs(repoName string) (*RepoAttrs, error) 
  GetBranchAttrs(repoName string, branchName string) (*BranchAttrs, error)
  GetCommitAttrs(repoName string, commitId string) (*CommitAttrs, error) 
  GetCommitMap(repoName string, commitId string) (*FileMap, error) 
  
  GetFileAttrs(repoName string, commitId string, filePath string) (*FileAttrs, error)
  DownloadRepo(repoName string) (*RepoAttrs, *BranchAttrs, *CommitAttrs, *FileMap, error)

  PutFile(repoName string, commitId string, fpath string, reader io.Reader) (*FileAttrs, int64, error)

  //CreateTask(config *task_pkg.TaskConfig) (*task_pkg.Task, error)
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

  newRepoAttrs :=  &RepoAttrs {
    Repo: newRepo,
  }
  
  newRepoAttrs.Description = "New Repo"
  

  err := a.db.Insert(REPO_KEY_PREFIX + name, newRepoAttrs)

  return err 
}

// TODO: handle errors and add validations
func (a *apiServer) RemoveRepo(name string) error {
  err := a.q.DeleteRepoAttrs(name)
  // TODO: delete branches, commits etc
  return err
}

func (a *apiServer) GetRepo(name string) (*RepoAttrs, error) {
  return a.q.GetRepoAttrs(name)
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
  var branch_attr *BranchAttrs
  var err error 

  if !a.q.CheckRepoExists(repoName) {
    base.Log("Failed to retrieve Repo Info: %s", repoName)
    return fmt.Errorf("Failed to retrieve Repo Info: %s", repoName)
  }

  if branch_attr, err = a.q.GetBranchAttrs(repoName, "master"); err != nil {
    base.Log("Failed to retrieve Branch Info")
    return err
  }

  if branch_attr.Head == nil {
    return fmt.Errorf("No commits in this repo")
  } 

  return a.EndCommit(repoName, branch_attr.Head.Id)
}

func (a *apiServer) IsOpenCommit(repoName string, commitId string, fpath string) bool {
  ct, _ := NewCommitTxn(repoName, commitId, a.db)
  return ct.IsOpenCommit()
}
 
// TODO: putfile by hash


// put file by reader
func (a *apiServer) PutFile(repoName string, commitId string, fpath string, reader io.Reader) (*FileAttrs, int64, error) {
   
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

  file_attrs, err := a.q.GetFileAttrs(repoName, commitId, fpath)
  
  if err != nil {
    if !db_pkg.IsErrRecNotFound(err) {
      base.Log("[apiServer.PutFile] Failed to retrieve File Info", repoName, commitId, fpath, err)
      return nil, 0, err
    }
  } 
  
  if file_attrs.Object != nil { 
    
    base.Debug("[apiServer.PutFile] Found file_attrs. Updating object. ", file_attrs.Object.Hash)
    object_hash := file_attrs.Object.Hash 
    
    new_hash, chksum, size, err :=  a.objWrapper.AppendObject(object_hash, reader) 
    base.Debug("[apiServer.PutFile] Object updated: ", new_hash, chksum)

    if err != nil {
      base.Log("[apiServer.PutFile] Failed to write object on server", err)
      return nil, 0, err
    } 
    
    // TODO: append should add to size or pull size from storage 
    file_attrs.SizeBytes = size

    //TODO: update file_attrs 
    return file_attrs, size, nil
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
  repo_attrs, err := a.q.GetRepoAttrs(repoName)

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

  branch_attr, err := a.q.GetBranchAttrs(repoName, repo_attrs.Branch.Name)

  // add object metadat to DB
  ct, err := NewCommitTxn(repoName, branch_attr.Head.Id, a.db)
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



func (a *apiServer) DownloadRepo(repoName string) (*RepoAttrs, *BranchAttrs, *CommitAttrs, *FileMap, error) {
  
  repo_attrs, err := a.q.GetRepoAttrs(repoName)
  if err != nil {
    base.Log("Invalid Repo - %s", repoName)
    return nil, nil, nil, nil, err
  }

  branch_attr, err := a.q.GetBranchAttrs(repoName, repo_attrs.Branch.Name)
  commit_head :=branch_attr.Head.Id
  commit_attrs, err:= a.GetCommitAttrs(repoName, commit_head)

  fmap, err:= a.q.GetFileMap(repoName, commit_head)

  return repo_attrs, branch_attr, commit_attrs, fmap, nil

}










