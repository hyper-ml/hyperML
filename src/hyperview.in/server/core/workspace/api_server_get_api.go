package workspace

import(
  "fmt"
  "hyperview.in/server/base"
)

func (a *apiServer) GetRepoAttrs(repoName string) (*RepoAttrs, error) {
  repo_attrs, err := a.q.GetRepoAttrs(repoName)
  if err != nil {
    base.Log("Invalid Repo - %s", repoName)
    return nil, err
  }

  return repo_attrs, nil
}

func (a *apiServer) GetBranchAttrs(repoName string, branchName string) (*BranchAttrs, error) {
  branch_attr, err := a.q.GetBranchAttrs(repoName, branchName)
  if err != nil {
    return nil, err
  }
  
  return branch_attr, nil
}


func (a *apiServer) GetCommitAttrs(repoName string, commitId string) (*CommitAttrs, error) {

  commit_attrs, err := a.q.GetCommitAttrsById(repoName, commitId) 
  if err !=  nil {
    return nil, err
  }

  if commit_attrs.Finished.IsZero() {
    return nil, fmt.Errorf("This repo has open commit. Please finish commit before downloading files.")
  } 

  return commit_attrs, nil
}

func (a *apiServer) GetCommitMap(repoName string, commitId string) (*FileMap, error) {
  commit_map, err := a.q.GetFileMap(repoName, commitId) 
  if err !=  nil {
    return nil, err
  } 

  return commit_map, nil
}


func (a *apiServer) GetFileAttrs(repoName string, commitId string, filePath string) (*FileAttrs, error) {
  file_attrs, err := a.q.GetFileAttrs(repoName, commitId, filePath)

  if err != nil {
    base.Log("apiServer.GetFileAttrs:", repoName, commitId, filePath)
    base.Log("error:", err)
    return nil, err
  }
  return file_attrs, nil
}







