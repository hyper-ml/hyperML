package workspace

import(
  "fmt"
  "hyperview.in/server/base"
)

func (a *apiServer) GetRepoInfo(repoName string) (*RepoInfo, error) {
  repo_info, err := a.q.GetRepoInfo(repoName)
  if err != nil {
    base.Log("Invalid Repo - %s", repoName)
    return nil, err
  }

  return repo_info, nil
}

func (a *apiServer) GetBranchInfo(repoName string, branchName string) (*BranchInfo, error) {
  branch_info, err := a.q.GetBranchInfo(repoName, branchName)
  if err != nil {
    return nil, err
  }
  
  return branch_info, nil
}


func (a *apiServer) GetCommitInfo(repoName string, commitId string) (*CommitInfo, error) {

  commit_info, err := a.q.GetCommitInfoById(repoName, commitId) 
  if err !=  nil {
    return nil, err
  }

  if commit_info.Finished.IsZero() {
    return nil, fmt.Errorf("This repo has open commit. Please finish commit before downloading files.")
  } 

  return commit_info, nil
}


func (a *apiServer) GetFileInfo(repoName string, commitId string, filePath string) (*FileInfo, error) {
  file_info, err := a.q.GetFileInfo(repoName, commitId, filePath)

  if err != nil {
    base.Log("apiServer.GetFileInfo:", repoName, commitId, filePath)
    base.Log("error:", err)
    return nil, err
  }
  return file_info, nil
}







