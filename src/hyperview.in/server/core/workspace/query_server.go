package workspace

import (
  "fmt"
  "encoding/json"
  "hyperview.in/server/base"
  db_pkg "hyperview.in/server/core/db"
)

const (
  REPO_KEY_PREFIX = "repo:"
  COMMIT_KEY_PREFIX = "commit:"
  BRANCH_KEY_PREFIX = "branch:"
  FILE_KEY_PREFIX = "file:"
  UPDATE_OP = "update"
  DELETE_OP = "delete"
)

//TODO: all updates need SELECT FOR UPDATE to add locking 

type queryServer struct {
  db *db_pkg.DatabaseContext
}

func NewQueryServer(d *db_pkg.DatabaseContext) *queryServer {
  return &queryServer {
    db: d,
  }
}

/******************************/
/*****  Repo Operations ******/

func (q *queryServer) getRepoKey(repoName string) string {
  return REPO_KEY_PREFIX + repoName
}

func (q *queryServer) CheckRepoExists(repoName string) bool {
  repo_key:= q.getRepoKey(repoName)
  return q.db.KeyExists(repo_key)
}

func (q *queryServer) DeleteRepoInfo(repoName string) error {
  repo_key := q.getRepoKey(repoName)

  return q.db.Delete(repo_key)
}

func (q *queryServer) UpdateRepoInfo(repoName string, repoInfo *RepoInfo) error {
  repo_key := q.getRepoKey(repoName)

  return q.db.Update(repo_key, repoInfo)
}


func (q *queryServer) GetRepoInfo(name string) (*RepoInfo, error) {
  var err error
  
  data, err := q.db.Get(REPO_KEY_PREFIX + name)
  
  repoInfo :=  &RepoInfo{} 
  err = json.Unmarshal(data, &repoInfo)
  
  return repoInfo, err
}


/******************************/
/*****  Branch Operations ******/

func (q *queryServer) getBranchKey(repoName string, branchName string) string {
  return REPO_KEY_PREFIX + repoName + ":" + BRANCH_KEY_PREFIX + branchName
}

func (q *queryServer) InsertBranchInfo(repoName string, branchName string, branchInfo *BranchInfo) error{
  branch_key:=  q.getBranchKey(repoName, branchName)

  return q.db.Insert(branch_key, branchInfo)
}

func (q *queryServer) GetBranchInfo(repoName string, branchName string) (*BranchInfo, error) {
  var err error
  branch_key:= q.getBranchKey(repoName, branchName)
  data, err := q.db.Get(branch_key)

  branch_info :=  &BranchInfo{} 
  err = json.Unmarshal(data, &branch_info)
  return branch_info, err
}

func (q *queryServer) UpdateBranchInfo(repoName string, branchName string, branchInfo *BranchInfo) error {
  branch_key:=  q.getBranchKey(repoName, branchName)

  return q.db.Update(branch_key, branchInfo)  
}


/******************************/
/*****  Commit Operations ******/
//

func (q *queryServer) getCommitKey(repoName string, commitId string) string {
  return REPO_KEY_PREFIX + repoName + ":" + COMMIT_KEY_PREFIX + commitId
}

func (q *queryServer) GetCommitInfoById(repoName string, commitId string) (*CommitInfo, error) {
  var err error
  commit_key:= q.getCommitKey(repoName, commitId)
  data, err := q.db.Get(commit_key)
  
  commit_info :=  &CommitInfo{} 
  err = json.Unmarshal(data, &commit_info)
  return commit_info, err
}

func (q *queryServer) GetCommitInfoByBranch(repoName, branchName string) (*CommitInfo, error) {
  var err error
  
  branch_info, err := q.GetBranchInfo(repoName, branchName)  
  if err != nil {
    return nil, err
  }
  
  if branch_info.Head == nil {
    base.Log("Branch has no head. Panic.", repoName, branchName)
    return nil, fmt.Errorf("Branch has no head. %s %s", repoName, branchName)
  }

  commit_key:= q.getCommitKey(repoName, branch_info.Head.Id)
  data, err := q.db.Get(commit_key)
  
  commit_info :=  &CommitInfo{} 
  err = json.Unmarshal(data, &commit_info)
  return commit_info, err
}

func (q *queryServer) InsertCommitInfo(repoName string, commitId string, commitInfo *CommitInfo) error {
  commit_key := q.getCommitKey(repoName, commitId)
  return q.db.Insert(commit_key, commitInfo)
}


func (q *queryServer) UpdateCommitInfo(repoName string, commitId string, commitInfo *CommitInfo) error {
  commit_key := q.getCommitKey(repoName, commitId)
  return q.db.Update(commit_key, commitInfo)
}

func (q *queryServer) DeleteCommitInfo(repoName, commitId string) (error) {
  commit_key := q.getCommitKey(repoName, commitId)
  return q.db.SoftDelete(commit_key)
}

/****************************************/
/*****  Commit File Map Operations ******/ 

func (q *queryServer) getFileMapKey(repoName string, commitId string) string {
  return REPO_KEY_PREFIX + repoName + ":" + COMMIT_KEY_PREFIX + commitId +":file_map"
}

func (q *queryServer) GetFileMap(repoName string, commitId string) (*FileMap, error) {
  var err error
  map_key := q.getFileMapKey(repoName, commitId)
  data, err := q.db.Get(map_key)
  
  map_info :=  &FileMap{} 
  err = json.Unmarshal(data, &map_info)
  return map_info, err
}

func (q *queryServer) InsertFileMap(repoName string, commitId string, fmapInfo *FileMap) error {
  map_key := q.getFileMapKey(repoName, commitId)
  fmt.Println("map info", fmapInfo)
  return q.db.Insert(map_key, fmapInfo)
}

func (q *queryServer) AddFileToMap(repoName string, commitId string, file *File) error {
  // TODO: should i add map if missing?
  map_key := q.getFileMapKey(repoName, commitId)
  fmap, err := q.GetFileMap(repoName, commitId)

  if err != nil {
    base.Log("Count not find the map: %s %s",repoName, commitId)
    return err 
  }
  // to do: add or update file 
  fmap.Add(file)

  return q.db.Update(map_key, fmap)
}


/**********************************/
/*****  File Meta Operations ******/


func (q *queryServer) getFileKey(repoName string, commitId string, filePath string) string {
  return q.getRepoKey(repoName) + ":commit:" + commitId + ":path:" + filePath
}

func (q *queryServer) GetFileInfo(repoName string, commitId string, filePath string) (*FileInfo, error) {
  var err error
  file_key:= q.getFileKey(repoName, commitId, filePath)
  data, err := q.db.Get(file_key)
  
  if err != nil {
    base.Log("queryServer.GetFileInfo(): Error retrieving FileInfo:", commitId, filePath)
    if db_pkg.IsErrRecNotFound(err) {
      return &FileInfo{}, err
    }
    return nil, err
  }

  file_info :=  &FileInfo{} 
  err = json.Unmarshal(data, &file_info)
  return file_info, err
}

func (q *queryServer) UpsertFileInfo(repoName string, commitId string, filePath string, fileInfo *FileInfo) error {
  var err error
  //file_id := fileInfo.File.Path
  //fileInfoFromDB, err := q.GetFileInfo(repoName, commitId, file_id)
  //TODO: check size difference 

  file_key:= q.getFileKey(repoName, commitId, filePath)
  err = q.db.Upsert(file_key, fileInfo)

  if err != nil {
    return err
  }
  return nil 
}

/***** Other Utitily DB methods *****/

func (q *queryServer) AssignBranch(repoName string, branch *Branch) (error) {
  // lock DB record 

  // update repoinfo 
  repo_info, err := q.GetRepoInfo(repoName)
  if err != nil {
    return err
  }
  repo_info.Branch = branch

  return q.UpdateRepoInfo(repoName, repo_info)

}













