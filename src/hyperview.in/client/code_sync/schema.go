package code_sync

import(
  ws "hyperview.in/server/core/workspace"
)


type PullRepoRequest struct {
  RepoName string
  CommitId string
  FileMap map[string]ws.File
}

 