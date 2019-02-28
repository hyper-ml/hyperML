package workspace

import (
  "io"
  "fmt"
  "net/http"
  "golang.org/x/net/context" 
  "hyperflow.in/server/pkg/storage"
  db_pkg "hyperflow.in/server/pkg/db"

)

type ApiServer interface { 
  InitRepo(name string) (*RepoAttrs, error)
  InitTypedRepo(repoType RepoType, name string) (*RepoAttrs, error)
  RemoveRepo(name string) error

  CheckRepoExists(name string) bool
  InitBranch(repoName, branchName, head string) (*BranchAttrs, error)

  InitCommit(repoName, branchName, commitId string) (*CommitAttrs, error)
  StartCommit(repoName, branchName string) (*CommitAttrs, error)
  CloseCommit(repoName string, branchName string, commitId string) (error)

  GetModel(srcRepoName, srcBranchName, srcCommitId string) (*Repo, *Branch, *Commit, error)
  GetOrCreateModel(srcRepoName, srcBranchName, srcCommitId string) (*Repo, *Branch, *Commit, error)

  GetRepo(repoName string) (*Repo, error)
  GetRepoAttrs(repoName string) (*RepoAttrs, error) 
  GetBranchAttrs(repoName string, branchName string) (*BranchAttrs, error)
  GetCommitAttrs(repoName string, commitId string) (*CommitAttrs, error) 
  GetCommitAttrsByBranch(repoName, branchName string) (*CommitAttrs, error)
  GetCommitMap(repoName string, commitId string) (*FileMap, error) 
  GetFileAttrs(repoName string, commitId string, filePath string) (*FileAttrs, error)
  
  GetCommitSize(repoName, branchName, commitId string) (int64, error) 
  GetBranchSize(repoName, branchName string) (int64, error)

  ExplodeRepoAttrs(repoName, branchName, commitId string) (*RepoAttrs, *BranchAttrs, *CommitAttrs, *FileMap, error)
  ExplodeRepo(repoName, branchName string) (*Repo, *Branch, *Commit, error)

  PutFile(repoName string, branchName string, commitId string, fpath string, reader io.Reader) (*FileAttrs, int64, error)

  GetFileURL(repoId, commitId, filepath string) (string, error)
  PutFileURL(repoId, branchName, commitId, filepath string) (string, error)
  PutFilePartURL(partSeq int, repoId, branchName, commitId, filepath string) (string, error)

  CreateDataset(name string) (*RepoAttrs, error) 
  //CreateTask(config *task_pkg.TaskConfig) (*task_pkg.Task, error)

  FileCheckIn(repoId, branchName, commitId, filepath string, size int64, checksum string) (*FileAttrs, error)  
  FileMergeAndCheckIn(repoId, branchName, commitId, filepath string, parts []int, size int64) (*FileAttrs, error)

}


type apiServer struct { 
  ctx context.Context
  db db_pkg.DatabaseContext
  q *queryServer
  objAPI storage.ObjectAPIServer 
}

func NewApiServer(db db_pkg.DatabaseContext, objAPI storage.ObjectAPIServer) (*apiServer, error) {
  return &apiServer{ 
    ctx: context.Background(),
    db: db, 
    q: &queryServer{
      db: db,
    },
    objAPI: objAPI,
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
      return nil, fmt.Errorf("failed to retreive head commit passed, err: %v", err) 
    }
    
    if commit_attrs.IsOpen() {
      return nil, fmt.Errorf("failed - Please close the commit in current branch first.")
    }

    head = commit_attrs.Commit
  }
  
  repo_attrs, err := a.q.GetRepoAttrs(repoName)
  if err != nil {
    return nil, err
  }

  new_branch := &Branch{
    Repo: repo_attrs.Repo, 
    Name: branchName}

  new_branch_attrs:= &BranchAttrs{
    Branch: new_branch, 
    Head: head}
  
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
  
  ct, err := NewCommitTxn(repoName, branchName, commitId, a.q)
  if err != nil {
    return nil, err
  }

  cattrs, err := ct.OpenCommit()
  if err != nil{
    return nil, err
  }
 
  return cattrs, nil 
} 

func (a *apiServer) StartCommit(repoName, branchName string) (*CommitAttrs, error) {
  return a.InitCommit(repoName, branchName, "") 
}
 
func (a *apiServer) CloseCommit(repoName, branchName, commitId string) error {
  ct, err := NewCommitTxn(repoName, branchName, commitId, a.q)
  if err != nil {
    return err
  }

  return ct.CloseCommit()

}

func (a *apiServer) CheckCommit(repoName string, branchName string, commitId string, fpath string) bool {
  ct, _ := NewCommitTxn(repoName, branchName, commitId, a.q)
  return ct.IsOpenCommit()
}
  
// Method puts file on to storage first then adds to commit
// 
func (a *apiServer) PutFile(repoName string, branchName string, commitId string, fpath string, reader io.Reader) (*FileAttrs, int64, error) {
   
  var size int64 
  var err error  

  ct, err := NewCommitTxn(repoName, branchName, commitId, a.q)
  if err != nil {
    return nil, size, err
  }

  file_attrs, err := a.q.GetFileAttrs(repoName, commitId, fpath)  
  if err != nil {
    if !db_pkg.IsErrRecNotFound(err) {
      return nil, size, err
    }
  } 
  var new_hash string
  var cs string

  if file_attrs.Object != nil { 
    
    object_name := file_attrs.Object.Hash     
    new_hash, _, size, err =  a.objAPI.SaveObject(object_name, "", reader, false)  

    if err != nil {
      return nil, size, err
    } 
    
    file_attrs.SizeBytes = size
    return file_attrs, size, nil
  } 
 
  new_hash, cs, size, err =  a.objAPI.CreateObject(reader, false) 
  if err!= nil {
    return nil, size, err
  }

  f_info, err := ct.AddFile(fpath, new_hash, size, cs)
  if err != nil {
    return nil, size, err
  }

  return f_info, size, nil 
}

func (a *apiServer) GetFileURL(repoId, commitId, filepath string) (string, error){
  file_attrs, err := a.q.GetFileAttrs(repoId, commitId, filepath)  
  if err != nil {
    return "", err
  }

  hash := file_attrs.GetObjectHash()
  if hash == "" {
    return "", fmt.Errorf("Invalid File Object. No hash found")
  }

  return a.objAPI.ObjectURL(http.MethodGet, hash)
}

func (a *apiServer) PutFileURL(repoId, branchName, commitId, filepath string)(string, error) {
  return a.PutFilePartURL(0, repoId, branchName, commitId, filepath)
} 

func (a *apiServer) PutFilePartURL(partSeq int, repoId, branchName, commitId, filepath string) (string, error){
  var signed_url string 
  
  var file_postfix string 

  if partSeq > 0 {
    file_postfix =  "." + fmt.Sprint(partSeq)
  }


  ct, err := NewCommitTxn(repoId, branchName, commitId, a.q)
  if err != nil {
    return signed_url, err
  }

  file_attrs, err := a.q.GetFileAttrs(repoId, commitId, filepath)  
  if err != nil {
    if !db_pkg.IsErrRecNotFound(err) {
      return signed_url, fmt.Errorf("failed getting file attrs, err: %v", err)
    }
  } 
  
  if (file_attrs.GetObjectHash() != "" ) {
      obj_name := file_attrs.GetObjectHash() + file_postfix    
      return a.objAPI.ObjectURL(http.MethodPut, obj_name)
  } else {
      new_hash := a.objAPI.NewObjectHash()
      _, err = ct.AddFile(filepath, new_hash, 0, "")
      return a.objAPI.ObjectURL(http.MethodPut, new_hash)
  }

  return "", fmt.Errorf("unknown error generating signed URL")
}

// This method is used along with PutFileURL to updates object size 
// (queried from storage) in file attrs after the client uploads object
// directly into storage 

func (a *apiServer) FileCheckIn(repoId, branchName, commitId, filepath string, size int64, checksum string) (*FileAttrs, error) {  
  var dest_hash string

  file_attrs, err := a.GetFileAttrs(repoId, commitId, filepath)
  if err != nil {
    return nil, err
  }

  if file_attrs.Object == nil {
    return nil, fmt.Errorf("failed to locate storage object in the commit file. Unexpected err: %v", err)
  } else {
    dest_hash = file_attrs.Object.Hash
  }

  dest_size := a.objAPI.ObjectSize(dest_hash)
  
  // todo: add tolerence?
  if dest_size != size { 
    return nil, fmt.Errorf("file size of local object doesnt match with remote storage. File may be corrupted. dest size: %d Local: %d dest hash: %s", dest_size, size, dest_hash)
  }  
  
  return a.CommitFile(repoId, branchName, commitId, filepath, dest_hash, dest_size, checksum)
}

func (a *apiServer) CommitFile(repoId, branchName, commitId, path, objname string, size int64, checksum string) (*FileAttrs, error) {
  
  ct, err := NewCommitTxn(repoId, branchName, commitId, a.q)
  if err != nil {
    return nil, err
  }
  
  file_attrs, err := ct.AddFile(path, objname, size, checksum)
  return file_attrs, err
}

// This method merges file parts and then calls checkIn to update size 
// of parent object

func (a *apiServer) FileMergeAndCheckIn(repoId, branchName, commitId, filepath string, parts []int, size int64) (*FileAttrs, error) {
  
  file_attrs, err := a.q.GetFileAttrs(repoId, commitId, filepath)
  if err != nil {
    return nil, fmt.Errorf("failed to confirm object details, err: %v", err)
  } 
  
  dest_hash := file_attrs.Object.Hash
  var src_hashes []string

  for _, part_seq := range parts {
    src_hashes =  append(src_hashes, dest_hash + "." + fmt.Sprint(part_seq))
  }
  
  err = a.objAPI.MergeObjects(dest_hash, src_hashes)
  if err != nil {
    return nil, fmt.Errorf("failed while merging objects: %v", err)
  }

  n := a.objAPI.ObjectSize(dest_hash)
  // todo : check size matches 

  return a.CommitFile(repoId, branchName, commitId, filepath, dest_hash, n, "")
}

func (a *apiServer) ExplodeRepoAttrs(repoName, branchName, commitId string) (*RepoAttrs, *BranchAttrs, *CommitAttrs, *FileMap, error) {
  
  var commit_attrs *CommitAttrs
  var fmap *FileMap
  var commit_head string

  branch_name := branchName 

  repo_attrs, err := a.q.GetRepoAttrs(repoName)
  if err != nil {
    return nil, nil, nil, nil, err
  }
  
  if branch_name == "" {
    branch_name = DefaultBranch
  }

  branch_attr, err := a.q.GetBranchAttrs(repoName, branchName)
  
  if branch_attr.Head != nil {
    commit_head = branch_attr.Head.Id
    commit_attrs, err = a.GetCommitAttrs(repoName, commit_head)
    fmap, err = a.q.GetFileMap(repoName, commit_head)
  }
  
  return repo_attrs, branch_attr, commit_attrs, fmap, nil
}

func (a *apiServer) ExplodeRepo(repoName, branchName string) (*Repo, *Branch, *Commit, error) {
  
  repo, err := a.GetRepo(repoName)
  if err != nil {
    return nil, nil, nil, err
  }

  branch_attr, err := a.q.GetBranchAttrs(repoName, branchName)
  if err != nil {
    return nil, nil, nil, err
  }

  return repo, branch_attr.Branch, branch_attr.Head, nil
}


func getModelName(srcRepoName, srcCommitID string) string {
  return "repo-" + srcRepoName + "-commit-" + srcCommitID + "-model"
}

func (a *apiServer) CreateModel(srcRepoName, srcBranchName, srcCommitId string) (*Repo, *Branch, *Commit, error) {
  repo_name   := getModelName(srcRepoName, srcCommitId)
  branch_name :="master"

  repo_attrs, err:= a.InitTypedRepo(MODEL_REPO, repo_name)
  if err != nil {
    return nil, nil, nil, err
  }

  branch := &Branch{Name: branch_name }
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
      retErr = err
      continue
    }
    size = size + f_attrs.Size()
  }

  return size, retErr
}


func (a *apiServer) GetCommitAttrsByBranch(repoName, branchName string) (*CommitAttrs, error) {
  return a.q.GetCommitAttrsByBranch(repoName, branchName)
}

 

