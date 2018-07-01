package workspace

import (
  "io"
  "fmt"
  "time"
  "golang.org/x/net/context"
  "hyperview.in/server/base"
  "hyperview.in/server/core/utils"
  "hyperview.in/server/core/db"
  "hyperview.in/server/core/storage"

)



type apiServer struct {
	version string
  ctx context.Context
  db *db.DatabaseContext
  q *queryServer
  objApiWrapper *ObjApiWrapper
}

func NewApiServer(db *db.DatabaseContext, oapi storage.ObjectAPIServer) (*apiServer, error) {
	return &apiServer{
    version: "0.1", 
    ctx: context.Background(),
    db: db, 
    q: &queryServer{
      db: db,
    },
    objApiWrapper: &ObjApiWrapper{
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


func (a *apiServer) addMasterBranch(repo *Repo, commit *Commit) (*BranchInfo, error) {
  var err error 
  var repo_info *RepoInfo
  
  branch := &Branch {
    Name: "master",
    Repo: repo,
  }
  branch_info := &BranchInfo{
    Branch: branch,
    Head: commit,
  }

  err = a.q.InsertBranchInfo(repo.Name, "master", branch_info)

  if err != nil {
    return nil, err
  }

  // update repoinfo 
  repo_info, err = a.q.GetRepoInfo(repo.Name)
  repo_info.Branch = branch

  //TODO: send context of error
  err = a.q.UpdateRepoInfo(repo.Name, repo_info)

  if err != nil {
    return nil, err
  }

  return branch_info, err
}

//TODO: copy file tree to new commit 
func (a *apiServer) addCommit(repo *Repo, parent_commit_info *CommitInfo) (*Commit, error) {
  commit_id := utils.NewUUID()
  commit_key:= REPO_KEY_PREFIX + repo.Name + ":commit:"+ commit_id
  commit := &Commit {
    Id: commit_id,
    Repo: repo,
  }

  var initFileMap = make(map[string]*File)

  if parent_commit_info != nil {
    initFileMap = parent_commit_info.FileMap
  } 

  commitInfo := &CommitInfo {
    Commit: commit,
    Parent_commit: parent_commit_info.Commit,
    Started: time.Now(),
    FileMap: initFileMap,
  }

  return commit, a.db.Insert(commit_key, commitInfo)
}

func (a *apiServer) scoopHead(branchInfo *BranchInfo, commit *Commit) error {
  branch := branchInfo.Branch
  repo := branchInfo.Branch.Repo

  branchInfo.Head = commit

  err:= a.q.UpdateBranchInfo(repo.Name, branch.Name, branchInfo)

  return err
}

func (a *apiServer) InitCommit(repoName string) (string, error) {
  var err error 
  var branch_info *BranchInfo
  var repo_info *RepoInfo
  var commit *Commit
  var parent_commit_info *CommitInfo

  repo_info, err = a.q.GetRepoInfo(repoName)

  if (err !=nil) {
    base.Log("InitiateCommit: Could not fetch repo with given name %s", repoName)
    return "", err
  }

  // if this is first ever commit. create master branch
  if repo_info.Branch == nil {
    
    // add master branch
    branch_info, err = a.addMasterBranch(repo_info.Repo, nil)

  } else {
    
    branch_info, err = a.q.GetBranchInfo(repoName, "master")

    // check if there is pending commit 
    if branch_info.Head != nil {

      parent_commit_info, err = a.q.GetCommitInfoById(repoName, branch_info.Head.Id)
      
      if parent_commit_info.Finished.IsZero() {
        return "", fmt.Errorf("There is a pending commit against this repo")
      }
    }
  }

  if branch_info != nil {
    // add commit with current head as parent 
    commit, err = a.addCommit(repo_info.Repo, parent_commit_info)

    // update branch head with new commit  
    err = a.scoopHead(branch_info, commit)
  }

  if err != nil {
    return "", fmt.Errorf("Failed to create or retrieve master branch: %s", err)
  }
  
  if commit != nil {
    return commit.Id, nil
  }

  return "", err
}

func (a *apiServer) finishCommit(repoName string, commitId string) error {

  commit_info, err := a.q.GetCommitInfoById(repoName, commitId)
  if (err !=nil) {
    base.Log("finishCommit: Could not fetch commit for repo %s with commit %s", repoName, commitId)
    return err
  }

  if commit_info.Finished.IsZero() {
    commit_info.Finished = time.Now()
    err = a.q.UpdateCommitInfo(repoName, commitId, commit_info)
    return err  
  } else {
    base.Log("finishCommit: No open commit for this repo", repoName)
    return fmt.Errorf("No open commit for this repo: %s", repoName)
  }
  
}

func (a *apiServer) CloseCommit(repoName string) error {
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

  return a.finishCommit(repoName, branch_info.Head.Id)
}

func (a *apiServer) AddFileToRepo(repoName string, path string, reader io.Reader) (int, error) {
  // find master branch and commit 
  // raise error if no open commit 

  repo_info, err := a.q.GetRepoInfo(repoName)

  if err != nil {
    base.Log("Invalid Repo - %s", repoName)
    return 0, err
  }
  branch_info, err := a.q.GetBranchInfo(repoName, repo_info.Branch.Name)
  commit_head :=branch_info.Head.Id

  commit_info, err := a.q.GetCommitInfoById(repoName, commit_head) 

  if !commit_info.Finished.IsZero() {
    return 0, fmt.Errorf("This repo has no open commit. Please initialize commit before adding files.")
  } 

  file_info, err := a.objApiWrapper.PutObject(commit_info.Commit, path, reader, false)

  err = a.q.UpsertFileInfo(repoName, commit_head, file_info) 

  return 0, nil
}
