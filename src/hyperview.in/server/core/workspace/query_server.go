package workspace

import (
  "fmt"
  "encoding/json"
  "hyperview.in/server/base"
  db_pkg "hyperview.in/server/core/db"
)

const (
  REPO_KEY_PREFIX = "repo:"
  DATA_REPO_KEY_PREFIX = "data:repo:"
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

func (q *queryServer) InsertRepoAttrs(repoName string, attrs *RepoAttrs) error {
  repo_key := q.getRepoKey(repoName)

  return q.db.Insert(repo_key, attrs)
}

func (q *queryServer) DeleteRepoAttrs(repoName string) error {
  repo_key := q.getRepoKey(repoName)

  return q.db.Delete(repo_key)
}

func (q *queryServer) UpdateRepoAttrs(repoName string, repoInfo *RepoAttrs) error {
  repo_key := q.getRepoKey(repoName)

  return q.db.Update(repo_key, repoInfo)
}


func (q *queryServer) GetRepoAttrs(name string) (*RepoAttrs, error) {
  var err error
  
  data, err := q.db.Get(REPO_KEY_PREFIX + name)
  
  repoInfo :=  &RepoAttrs{} 
  err = json.Unmarshal(data, &repoInfo)
  
  return repoInfo, err
}


/******************************/
/*****  Branch Operations ******/

func (q *queryServer) getBranchKey(repoName string, branchName string) string {
  return REPO_KEY_PREFIX + repoName + ":" + BRANCH_KEY_PREFIX + branchName
}

func (q *queryServer) InsertBranchAttrs(repoName string, branchName string, branchInfo *BranchAttrs) error{
  branch_key:=  q.getBranchKey(repoName, branchName)

  return q.db.Insert(branch_key, branchInfo)
}

func (q *queryServer) GetBranchAttrs(repoName string, branchName string) (*BranchAttrs, error) {
  var err error
  branch_key:= q.getBranchKey(repoName, branchName)
  data, err := q.db.Get(branch_key)

  branch_attr :=  &BranchAttrs{} 
  err = json.Unmarshal(data, &branch_attr)
  return branch_attr, err
}

func (q *queryServer) UpdateBranchAttrs(repoName string, branchName string, branchInfo *BranchAttrs) error {
  branch_key:=  q.getBranchKey(repoName, branchName)

  return q.db.Update(branch_key, branchInfo)  
}


func (q *queryServer) DeleteBranchAttrs(repoName, branchName string) error {
  branch_key:=  q.getBranchKey(repoName, branchName)
  return q.db.Delete(branch_key)  
}


/******************************/
/*****  Commit Operations ******/
//

func (q *queryServer) getCommitKey(repoName string, commitId string) string {
  return REPO_KEY_PREFIX + repoName + ":" + COMMIT_KEY_PREFIX + commitId
}

func (q *queryServer) GetBranchCommitById(repoName string, branchName string, commitId string) (*CommitAttrs, error) {
  var err error
  commit_key:= q.getCommitKey(repoName, commitId)
  data, err := q.db.Get(commit_key)
  
  commit_attrs :=  &CommitAttrs{} 
  err = json.Unmarshal(data, &commit_attrs)
  return commit_attrs, err
}

func (q *queryServer) GetCommitAttrsById(repoName string, commitId string) (*CommitAttrs, error) {
  var err error
  commit_key:= q.getCommitKey(repoName, commitId)
  data, err := q.db.Get(commit_key)
  
  commit_attrs :=  &CommitAttrs{} 
  err = json.Unmarshal(data, &commit_attrs)
  return commit_attrs, err
}

func (q *queryServer) GetCommitAttrsByBranch(repoName, branchName string) (*CommitAttrs, error) {
  var err error
  
  branch_attr, err := q.GetBranchAttrs(repoName, branchName)  
  if err != nil {
    return nil, err
  }
  
  if branch_attr.Head == nil {
    base.Log("Branch has no head. Panic.", repoName, branchName)
    return nil, fmt.Errorf("Branch has no head. %s %s", repoName, branchName)
  }

  commit_key:= q.getCommitKey(repoName, branch_attr.Head.Id)
  data, err := q.db.Get(commit_key)
  
  commit_attrs :=  &CommitAttrs{} 
  err = json.Unmarshal(data, &commit_attrs)
  return commit_attrs, err
}

func (q *queryServer) InsertCommitAttrs(repoName string, commitId string, commitInfo *CommitAttrs) error {
  commit_key := q.getCommitKey(repoName, commitId)
  return q.db.Insert(commit_key, commitInfo)
}


func (q *queryServer) UpdateCommitAttrs(repoName string, commitId string, commitInfo *CommitAttrs) error {
  commit_key := q.getCommitKey(repoName, commitId)
  return q.db.Update(commit_key, commitInfo)
}

func (q *queryServer) DeleteCommitAttrs(repoName, commitId string) (error) {
  commit_key := q.getCommitKey(repoName, commitId)
  return q.db.SoftDelete(commit_key)
}

func (q *queryServer) IsBranchHead(repoName string, branchName string, commitId string) (bool, error) {
  branch_attrs, err := q.GetBranchAttrs(repoName, branchName)
  if err != nil {
    base.Log("[commitTxn.isBranchHead] Something wrong. branch_attrs missing for repo: ", repoName, branchName)
    return false, err
  }
  if branch_attrs == nil {
    return false, errBranchMissing(repoName + ":" + branchName)
  }

  if branch_attrs.Head == nil {
    return false, nil
  } else if branch_attrs.Head.Id == commitId {
    return true, nil
  }

  return false, nil

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

func (q *queryServer) GetFileAttrs(repoName string, commitId string, filePath string) (*FileAttrs, error) {
  var err error
  file_key:= q.getFileKey(repoName, commitId, filePath)
  data, err := q.db.Get(file_key)
  
  if err != nil {
    base.Log("queryServer.GetFileAttrs(): Error retrieving FileAttrs:", commitId, filePath)
    if db_pkg.IsErrRecNotFound(err) {
      return &FileAttrs{}, err
    }
    return nil, err
  }

  file_attrs :=  &FileAttrs{} 
  err = json.Unmarshal(data, &file_attrs)
  return file_attrs, err
}

func (q *queryServer) UpsertFileAttrs(repoName string, commitId string, filePath string, fileAttr *FileAttrs) error {
  var err error
  //file_id := fileAttr.File.Path
  //fileInfoFromDB, err := q.GetFileAttrs(repoName, commitId, file_id)
  //TODO: check size difference 

  file_key:= q.getFileKey(repoName, commitId, filePath)
  err = q.db.Upsert(file_key, fileAttr)

  if err != nil {
    return err
  }
  return nil 
}

/***** Other Utitily DB methods *****/

func (q *queryServer) AssignBranch(repoName string, branch *Branch) (error) {
  // lock DB record 

  // update repoinfo 
  repo_attrs, err := q.GetRepoAttrs(repoName)
  if err != nil {
    return err
  }
  repo_attrs.AddBranch(branch) 

  return q.UpdateRepoAttrs(repoName, repo_attrs)

}













