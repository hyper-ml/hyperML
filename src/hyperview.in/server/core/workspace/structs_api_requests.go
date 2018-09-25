package workspace 

type RepoMessage struct {
  Repo *Repo `json:"repo"`
  Branch *Branch `json:"branch"` 
  Commit *Commit `json:"commit"`
  FileMap *FileMap `json:"file_map"`
}

type StandardRepoMessage struct {
  RepoMessage
  Output *RepoMessage `json:"output"`
  Model *RepoMessage `json:"model"`
}

type RepoAttrsMessage struct {
  RepoAttrs *RepoAttrs `json:"repo_attrs"`
  BranchAttrs *BranchAttrs `json:"branch_attrs"` 
  CommitAttrs *CommitAttrs `json:"commit_attrs"`
  FileMap *FileMap `json:"file_map"`
}
 
type StdRepoAttrsMessage struct {
  RepoAttrsMessage
  OutputAttrs *RepoAttrsMessage `json:"output_attrs"`
  ModelAttrs *RepoAttrsMessage `json:"model_attrs"`
}
 

type ModelRepoRequest struct {
}

type ModelRepoResponse struct {
  Repo *Repo
  Branch *Branch
  Commit *Commit
}

type GetRepoRequest struct {

}


type GetRepoResponse struct {
  Repo Repo `json:"repo"`
  Branch Branch `json:"branch"` 
  CommitId string `json:"commit_id"`
  FileMap map[string]File `json:"file_map"`
}


type PutFileResponse struct {
  FileAttrs *FileAttrs `json:"file_attrs"`
  Written int64  `json:"written"`
  Error string `json:"error"`
}


type CommitSizeRequest struct {
  Repo Repo `json:"repo"`
  Branch Branch `json:"branch"` 
  CommitId string `json:"commit_id"`
}

type CommitSizeResponse struct {
  Size int64 `json:"size"`
}