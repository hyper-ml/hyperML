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
  InitRepo(name string) (*RepoAttrs, error)
  InitBranch(repoName, branchName, head string) (*BranchAttrs, error)

  InitCommit(repoName, branchName string) (*CommitAttrs, error)
  StartCommit(repoName, branchName string) (string, error)
  EndCommit(repoName string, branchName string, commitId string) (error)
  
  GetRepoAttrs(repoName string) (*RepoAttrs, error) 
  GetBranchAttrs(repoName string, branchName string) (*BranchAttrs, error)
  GetCommitAttrs(repoName string, commitId string) (*CommitAttrs, error) 
  GetCommitMap(repoName string, commitId string) (*FileMap, error) 
  
  GetFileAttrs(repoName string, commitId string, filePath string) (*FileAttrs, error)
  DownloadRepo(repoName, branchName string) (*RepoAttrs, *BranchAttrs, *CommitAttrs, *FileMap, error)

  PutFile(repoName string, branchName string, commitId string, fpath string, reader io.Reader) (*FileAttrs, int64, error)

  CreateDataset(name string) error 
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


func (a *apiServer) InitRepo(name string) (*RepoAttrs, error) {
  //TODO: auth, validate repo name
  master_branch:= "master"
  repo_name := name
  var err error 

  new_repo := &Repo {
      Name: repo_name,
  }

  new_branch := &Branch {
    Repo: new_repo, 
    Name: master_branch,
  }

  branch_attrs:= &BranchAttrs { Branch: new_branch }

  err = a.q.InsertBranchAttrs(repo_name, master_branch, branch_attrs)

  repo_attrs :=  &RepoAttrs {
    Repo: new_repo,
  }
  
  repo_attrs.Description = "New Repo"
  repo_attrs.Type = STANDARD_REPO
  repo_attrs.AddBranch(new_branch)

  err = a.q.InsertRepoAttrs(repo_name, repo_attrs)
  if err != nil {
    base.Log("[apiServer.InitRepo] Failed to create repo: ", err)
    _ = a.q.DeleteBranch(repo_name, master_branch)
    return nil, err
  } 

  return repo_attrs, nil
}


func ( a *apiServer) InitBranch(repoName, branchName, headCommitId string) (*BranchAttrs, error) {
  var head *Commit

  if headCommitId != "" {
    commit_attrs, err := a.q.GetCommitAttrsById(repoName, headCommitId)
    if err != nil {
      return nil, err 
    }
    head = commit_attrs.Commit
  }
  
  repo_attrs, err := a.q.GetRepoAttrs(repoName)
  if err != nil {
    base.Log("[apiServer.InitBranch] Failed to retrieve repo attributes for : ", repoName)
    return nil, err
  }

  new_branch := &Branch {Repo: repo_attrs.Repo, Name: branchName}
  new_branch_attrs:= &BranchAttrs {Branch: new_branch, Head: head}
  err = a.q.InsertBranchAttrs(repoName, branchName, new_branch_attrs)
  if err != nil {
    return nil, err
  }
  return new_branch_attrs, nil
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
 
func (a *apiServer) InitCommit(repoName, branchName string) (*CommitAttrs, error) {
  var commit_id string
  ct, err := NewCommitTxn(repoName, branchName, commit_id, a.db)
  if err != nil {
    return nil, err
  }
  return ct.Init() 
} 

func (a *apiServer) StartCommit(repoName, branchName string) (commitId string, err error) {
  ct, err := NewCommitTxn(repoName, branchName, "", a.db)
  if err != nil {
    return "", err
  }
  return ct.Start() 
}

func (a *apiServer) EndCommit(repoName string, branchName string, commitId string) (error) {
  ct, err := NewCommitTxn(repoName, branchName, commitId, a.db)
  if err != nil {
    return err
  }
  return ct.End()
}


func (a *apiServer) CloseOpenCommit(repoName, branchName string) error {
  var branch_attr *BranchAttrs
  var err error 
  
  if branch_attr, err = a.q.GetBranchAttrs(repoName, branchName); err != nil {
    base.Log("Failed to retrieve Branch Info")
    return err
  }

  if branch_attr.Head == nil {
    return fmt.Errorf("No commits in this repo")
  } 

  return a.EndCommit(repoName, branchName, branch_attr.Head.Id)
}

func (a *apiServer) IsOpenCommit(repoName string, branchName string, commitId string, fpath string) bool {
  ct, _ := NewCommitTxn(repoName, branchName, commitId, a.db)
  return ct.IsOpenCommit()
}
 
// TODO: putfile by hash


// put file by reader
func (a *apiServer) PutFile(repoName string, branchName string, commitId string, fpath string, reader io.Reader) (*FileAttrs, int64, error) {
   
  var size int64 
  var err error  

  ct, err := NewCommitTxn(repoName, branchName, commitId, a.db)
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
 

func (a *apiServer) AddFileToRepo(repoName string, branchName string, path string, reader io.Reader) (int64, error) {
  // find master branch and commit 
  // raise error if no open commit 
  var size int64 

  if !a.q.CheckRepoExists(repoName) {
    base.Log("Invalid Repo - %s", repoName)
    return 0, errInvalidRepoName(repoName)
  }
 
  // create object in store first 
  objName, cs, size, err :=  a.objWrapper.CreateObject(reader) 
  if err != nil {
    base.Log("Failed to write object on server", err)
    return 0, err
  }

  branch_attr, err := a.q.GetBranchAttrs(repoName, branchName)

  // add object metadat to DB
  ct, err := NewCommitTxn(repoName, branchName, branch_attr.Head.Id, a.db)
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



func (a *apiServer) DownloadRepo(repoName, branchName string) (*RepoAttrs, *BranchAttrs, *CommitAttrs, *FileMap, error) {
  
  repo_attrs, err := a.q.GetRepoAttrs(repoName)
  if err != nil {
    base.Log("Invalid Repo - %s", repoName)
    return nil, nil, nil, nil, err
  }

  branch_attr, err := a.q.GetBranchAttrs(repoName, branchName)
  commit_head :=branch_attr.Head.Id
  commit_attrs, err:= a.GetCommitAttrs(repoName, commit_head)

  fmap, err:= a.q.GetFileMap(repoName, commit_head)

  return repo_attrs, branch_attr, commit_attrs, fmap, nil

}










