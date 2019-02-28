package schema

import(
  ws "hyperflow.in/server/pkg/workspace"
)


type GetRepoResponse struct {
  Repo ws.Repo `json:"repo"`
  Branch ws.Branch `json:"branch"` 
  CommitId string `json:"commit_id"`
  FileMap map[string]ws.File `json:"file_map"`
}
 
 

type GetRepoAttrsResponse struct {
  RepoAttrs ws.RepoAttrs `json:"repo_attrs"`
  BranchAttrs ws.BranchAttrs `json:"branch_attr"` 
  CommitAttrs ws.CommitAttrs `json:"commit_attrs"`
  FileMap map[string]ws.File `json:"file_map"`
}
 

type PutFileResponse struct {
  FileAttrs *ws.FileAttrs `json:"file_attrs"`
  Written int64  `json:"written"`
  Error string `json:"error"`
}