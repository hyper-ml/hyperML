package rest


import(
  "net/http"
  "hyperflow.in/server/pkg/base" 
  "hyperflow.in/server/pkg/base/structs"  
  ws "hyperflow.in/server/pkg/workspace"
  
)

func isEmpty(v string) bool {
  if v == "" {
    return true
  }
  return false
}

func (h *Handler) handleGetCommitAttrs() error {
  var commit_attrs *ws.CommitAttrs
  var err error 

  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  var response map[string]interface{}
  repo_name := h.getQuery("repoName") 
  commit_id := h.getQuery("commitId")
  branch_name := h.getQuery("branchName")


  if isEmpty(repo_name) {
    repo_name, _ = h.getMandatoryUrlParam("repoName")
  }

  if isEmpty(commit_id) {
    commit_id, _ = h.getMandatoryUrlParam("commitId")
  }

  if isEmpty(repo_name) || (isEmpty(commit_id) && isEmpty(branch_name)) { 
    return base.HTTPErrorf(http.StatusBadRequest, "one of these params is missing - repoName, branchName, commitId: ", repo_name, branch_name, commit_id)
  }

  switch {
  case !isEmpty(commit_id):
    commit_attrs, err = h.server.wsAPI.GetCommitAttrs(repo_name, commit_id)
  default: 
    commit_attrs, err = h.server.wsAPI.GetCommitAttrsByBranch(repo_name, branch_name)
  } 

  if err != nil {
    base.Error("[Handler.handleGetCommitAttrs] Error: ", repo_name, commit_id, err)
    return base.HTTPErrorf(http.StatusBadRequest, "Failed to retrieve commit attributes for repo, commit ID: %s %s %s", repo_name, commit_id, err.Error())
  }
  
  response = structs.Map(commit_attrs)
  h.writeJSON(response) 
  return nil
}


func (h *Handler) handleGetOrStartCommit() error {
  
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  repo_name := h.getQuery("repoName")
  branch_name := h.getQuery("branchName")
  commit_id := h.getQuery("commitId")

  //TODO: handle error
  commit_attrs, err := h.server.wsAPI.InitCommit(repo_name, branch_name, commit_id)
  
  if err != nil {
    return base.HTTPErrorf(http.StatusBadRequest, " Failed to initialize a commit: " + err.Error())
  }
  
  response = structs.Map(commit_attrs)
  h.writeJSON(response) 
  return nil
} 


func (h *Handler) handleCloseCommit() error {
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  repo_name := h.getQuery("repoName")
  branch_name := h.getQuery("branchName")
  commit_id := h.getQuery("commitId")

  err := h.server.wsAPI.CloseCommit(repo_name, branch_name, commit_id)
  
  if err != nil {
    return base.HTTPErrorf(http.StatusBadRequest, "Failed to close commit: %s", err)
  }

  response = map[string]interface{}{
    "status": "success",
  }

  h.writeJSON(response) 
  return nil
}

func (h *Handler) handleGetCommitSize() error {
  base.Info("[Handler.handleGetCommitSize]")
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  repo_name, _ := h.getMandatoryUrlParam("repoName")
  branch_name, _ := h.getMandatoryUrlParam("branchName")
  commit_id, _ := h.getMandatoryUrlParam("commitId")

  if repo_name == "" || branch_name== "" || commit_id == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "invalid_input: One of the params is missing: repo_name commit_id branch_name")
  }
  commit_size, err := h.server.wsAPI.GetCommitSize(repo_name, branch_name, commit_id)
   
  if err != nil {
    base.Warn("[Handler.handleGetCommitSize] Failed to get commit size: ", err)
    base.HTTPErrorf(http.StatusBadRequest, err.Error())
  }

  size_resp := &ws.CommitSizeResponse{}
  size_resp.Size = commit_size

  response = structs.Map(size_resp)
  h.writeJSON(response)    
  
  return nil
}
