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
  CheckRepoExists(name string) bool
  InitBranch(repoName, branchName, head string) (*BranchAttrs, error)

  InitCommit(repoName, branchName, commitId string) (*CommitAttrs, error)
  StartCommit(repoName, branchName string) (string, error)
  EndCommit(repoName string, branchName string, commitId string) (error)

  GetModel(srcRepoName, srcBranchName, srcCommitId string) (*Repo, *Branch, *Commit, error)
  GetOrCreateModel(srcRepoName, srcBranchName, srcCommitId string) (*Repo, *Branch, *Commit, error)

  GetRepo(repoName string) (*Repo, error)
  GetRepoAttrs(repoName string) (*RepoAttrs, error) 
  GetBranchAttrs(repoName string, branchName string) (*BranchAttrs, error)
  GetCommitAttrs(repoName string, commitId string) (*CommitAttrs, error) 
  GetCommitMap(repoName string, commitId string) (*FileMap, error) 
  GetFileAttrs(repoName string, commitId string, filePath string) (*FileAttrs, error)
  
  GetCommitSize(repoName, branchName, commitId string) (int64, error) 
  GetBranchSize(repoName, branchName string) (int64, error)

  ExplodeRepoAttrs(repoName, branchName, commitId string) (*RepoAttrs, *BranchAttrs, *CommitAttrs, *FileMap, error)
  ExplodeRepo(repoName, branchName string) (*Repo, *Branch, *Commit, error)

  PutFile(repoName string, branchName string, commitId string, fpath string, reader io.Reader) (*FileAttrs, int64, error)

  CreateDataset(name string) (*RepoAttrs, error) 
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

func (a *apiServer) CheckRepoExists(name string) bool {
  return a.q.CheckRepoExists(name)
}

func (a *apiServer) InitRepo(name string) (*RepoAttrs, error) {
  return a.InitTypedRepo(STANDARD_REPO, name)
}

func (a *apiServer) InitTypedRepo(repoType RepoType, name string) (*RepoAttrs, error) {
  //TODO: auth, validate repo name
  master_branch:= "master"
  repo_name := name
  repo_type := repoType

  var err error 

  if a.q.CheckRepoExists(repo_name) {
    base.Debug("[apiServer.InitRepo] Repo already exists: ", repo_name)
    return nil, errRepoNameExists(repo_name)
  }

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
  repo_attrs.Type = repo_type

  repo_attrs.AddBranch(new_branch)

  err = a.q.InsertRepoAttrs(repo_name, repo_attrs)
  if err != nil {
    base.Log("[apiServer.InitRepo] Failed to create repo: ", err)
    _ = a.q.DeleteBranchAttrs(repo_name, master_branch)
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
  // loop through repo branches and delete 
  repo_attrs, err := a.GetRepoAttrs(name)
  if err != nil {
    return err 
  }
  if repo_attrs.Repo.Name != "" {
    if len(repo_attrs.Branches) > 0 {
      for branch_name, _ := range repo_attrs.Branches {
        err := a.q.DeleteBranchAttrs(repo_attrs.Repo.Name, branch_name)
        if err != nil {
          return err
        }
      }
    }
    err := a.q.DeleteRepoAttrs(name)
    if err != nil {
      return err
    }
  }

  return nil
}


 
func (a *apiServer) InitCommit(repoName, branchName, commitId string) (*CommitAttrs, error) {
  var commit_id string = commitId
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


//TODO: add code for commitId 
func (a *apiServer) ExplodeRepoAttrs(repoName, branchName, commitId string) (*RepoAttrs, *BranchAttrs, *CommitAttrs, *FileMap, error) {
  branch_name := branchName 

  repo_attrs, err := a.q.GetRepoAttrs(repoName)
  if err != nil {
    base.Log("Invalid Repo - %s", repoName)
    return nil, nil, nil, nil, err
  }
  
  if branch_name == "" {
    branch_name = DefaultBranch
  }

  branch_attr, err := a.q.GetBranchAttrs(repoName, branchName)
  commit_head :=branch_attr.Head.Id
  commit_attrs, err:= a.GetCommitAttrs(repoName, commit_head)

  fmap, err:= a.q.GetFileMap(repoName, commit_head)

  return repo_attrs, branch_attr, commit_attrs, fmap, nil
}

func (a *apiServer) ExplodeRepo(repoName, branchName string) (*Repo, *Branch, *Commit, error) {
  
  repo, err := a.GetRepo(repoName)
  if err != nil {
    base.Debug("[apiServer.ExplodeRepo] Invalid Repo - %s", repoName)
    return nil, nil, nil, err
  }

  branch_attr, err := a.q.GetBranchAttrs(repoName, branchName)
  if err != nil {
    base.Debug("[apiServer.ExplodeRepo] Invalid branch - %s", repoName)
    return nil, nil, nil, err
  }

  return repo, branch_attr.Branch, branch_attr.Head, nil
}


func getModelName(srcRepoName, srcCommitID string) string {
  return "repo-" + srcRepoName + "-commit-" + srcCommitID + "-model"
}

func (a *apiServer) CreateModel(srcRepoName, srcBranchName, srcCommitId string) (*Repo, *Branch, *Commit, error) {
  repo_name := getModelName(srcRepoName, srcCommitId)
  branch_name:="master"

  repo_attrs, err:= a.InitTypedRepo(MODEL, repo_name)
  if err != nil {
    base.Log("[apiServer.CreateModel] Failed to create model repo: ", err)
    return nil, nil, nil, err
  }

  branch := &Branch{ Name: branch_name }
  commit_attrs, err:= a.InitCommit(repo_name, branch_name, "")

  return repo_attrs.Repo, branch, commit_attrs.Commit, nil
}

func (a *apiServer) GetModel(srcRepoName, srcBranchName, srcCommitId string) (*Repo, *Branch, *Commit, error) {
  repo_name := getModelName(srcRepoName, srcCommitId)
  branch_name := "master"

  repo_attrs, _ := a.q.GetRepoAttrs(repo_name)
  branch_attrs, _ := a.q.GetBranchAttrs(repo_name, branch_name)

  return repo_attrs.Repo, branch_attrs.Branch, branch_attrs.Head, nil
}

 
func (a *apiServer) GetOrCreateModel(srcRepoName, srcBranchName, srcCommitId string) (*Repo, *Branch, *Commit, error) {
  if !a.CheckRepoExists(getModelName(srcRepoName, srcCommitId)) {
    return a.CreateModel(srcRepoName, srcBranchName, srcCommitId)
  } 
  
  return a.GetModel(srcRepoName, srcBranchName, srcCommitId)
}

// returns size of head commit in branch
func (a *apiServer) GetBranchSize(repoName, branchName string) (int64, error) {
  var size int64
  branch_attrs, _ := a.q.GetBranchAttrs(repoName, branchName)
  if branch_attrs.Head.Id == "" {
    return size, nil
  } 

  return a.GetCommitSize(repoName, branchName, branch_attrs.Head.Id)
}

func (a *apiServer) GetCommitSize(repoName, branchName, commitId string) (int64, error) {
  var size int64
  var retErr error
  file_map, _ := a.q.GetFileMap(repoName, commitId)

  if len(file_map.Entries) == 0 {
    return size, nil
  }

  for name, _ := range file_map.Entries {
    f_attrs, err := a.q.GetFileAttrs(repoName, commitId, name)
    if err != nil {
      base.Debug("[apiServer.GetCommitSize] Failed to find size of file: ", repoName, commitId, name)
      retErr = err
      continue
    }
    size = size + f_attrs.Size()
  }

  base.Info("[apiServer.GetSize] Size of Repo: ", size, repoName, branchName, commitId)
  return size, retErr
}








