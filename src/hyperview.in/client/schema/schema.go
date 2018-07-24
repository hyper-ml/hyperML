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
 