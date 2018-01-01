package workspace 



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