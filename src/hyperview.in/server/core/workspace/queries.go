package workspace

import (
  "fmt"
  "encoding/json"
  "hyperview.in/server/base"

  "hyperview.in/server/core/db"
)

const (
  REPO_KEY_PREFIX = "repo:"
  COMMIT_KEY_PREFIX = "commit:"
  BRANCH_KEY_PREFIX = "branch:"
  FILE_KEY_PREFIX = "file:"
)

type queryServer struct {
  db *db.DatabaseContext
}

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

func (q *queryServer) getBranchKey(repoName string, branchName string) string {
  return REPO_KEY_PREFIX + repoName + ":" + BRANCH_KEY_PREFIX + branchName
}

func (q *queryServer) InsertBranchInfo(repoName string, branchName string, branchInfo *BranchInfo) error{
  branch_key:=  q.getBranchKey(repoName, branchName)

  return q.db.Insert(branch_key, branchInfo)
}

func (q *queryServer) GetBranchInfo(repoName string, branchName string) (*BranchInfo, error) {
  var err error
  
  data, err := q.db.Get(REPO_KEY_PREFIX + repoName + ":" + BRANCH_KEY_PREFIX + branchName)
  
  branch_info :=  &BranchInfo{} 
  err = json.Unmarshal(data, &branch_info)
  return branch_info, err
}

func (q *queryServer) UpdateBranchInfo(repoName string, branchName string, branchInfo *BranchInfo) error {
  branch_key:=  q.getBranchKey(repoName, branchName)

  return q.db.Update(branch_key, branchInfo)  
}


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

func (q *queryServer) UpdateCommitInfo(repoName string, commitId string, commitInfo *CommitInfo) error {
  commit_key := q.getCommitKey(repoName, commitId)
  return q.db.Update(commit_key, commitInfo)
}

func (q *queryServer) AppendFileToCommit(commitInfo *CommitInfo, fileInfo *FileInfo) error {

  commitInfo.treeLock.Lock()
  defer commitInfo.treeLock.Unlock()

  if commitInfo.Commit == nil {
    base.Log("Commit Info missing Commit Record. Its bad!")
    return fmt.Errorf("Incorrect Commit Info record. Missing Commit Id")
  }

  commitInfo, err := commitInfo.AddFileToMap(fileInfo.File)

  base.Log("commit Info after add: %s", commitInfo.Commit.Id, commitInfo.FileMap, err)

  if err != nil {
    return err
  }  
  commit_key:= q.getCommitKey(commitInfo.Commit.Repo.Name, commitInfo.Commit.Id)
  return q.db.Update(commit_key, commitInfo)
}

func (q *queryServer) getFileKey(repoName string, commitId string, filePath string) string {
  return q.getRepoKey(repoName) + ":commit" + commitId + ":file:" + filePath
}

func (q *queryServer) GetFileInfo(repoName string, commitId string, filePath string) (*FileInfo, error) {
  var err error
  file_key:= q.getFileKey(repoName, commitId, filePath)
  data, err := q.db.Get(file_key)
  
  file_info :=  &FileInfo{} 
  err = json.Unmarshal(data, &file_info)
  return file_info, err
}

func (q *queryServer) UpsertFileInfo(repoName string, commitId string, fileInfo *FileInfo) error {
  var err error
  file_id := fileInfo.File.Path

  //TODO: add index 
  fileInfoFromDB, err := q.GetFileInfo(repoName, commitId, file_id)

  if fileInfoFromDB != nil {
    //file already exists. Compare check sum
    if (fileInfoFromDB.CheckSum == fileInfo.CheckSum) {
      //do nothing
      base.Log("file already exists: %s %s %s", file_id, fileInfoFromDB.CheckSum, fileInfo.CheckSum)
      return nil
    }
  }

  file_key:= q.getFileKey(repoName, commitId, file_id)
  err = q.db.Upsert(file_key, fileInfo)

  if err != nil {
    return err
  }

  commit_info, err := q.GetCommitInfoById(repoName, commitId)

  if err != nil {
    return err
  }
  // TODO: delete fileinfo if commit info update fails

  return q.AppendFileToCommit(commit_info, fileInfo)

}
















