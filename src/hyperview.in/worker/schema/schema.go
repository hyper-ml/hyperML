package schema

import(
  ws "hyperview.in/server/core/workspace"
)

type GetRepoResponse struct {
  Repo ws.Repo `json:"repo"`
  Branch ws.Branch `json:"branch"`  
  CommitId string `json:"commit_id"`
  FileMap map[string]ws.File `json:"file_map"`
}
 

type GetRepoInfoResponse struct {
  RepoInfo ws.RepoInfo `json:"repo_info"`
  BranchInfo ws.BranchInfo `json:"branch_info"` 
  CommitInfo ws.CommitInfo `json:"commit_info"`
  FileMap map[string]ws.File `json:"file_map"`
}
 

type PutFileResponse struct {
  FileInfo *ws.FileInfo `json:"file_info"`
  Written int64  `json:"written"`
  Error string `json:"error"`
}