package rest


import(
  "net/http"
  "hyperview.in/server/base" 
  "hyperview.in/server/base/structs"  
  ws "hyperview.in/server/core/workspace"
  
)
func (h *Handler) handleGetCommitAttrs() error {
  base.Info("[Handler.handleGetCommitAttrs]")
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  var response map[string]interface{}
  repo_name := h.getQuery("repoName") 
  commit_id := h.getQuery("commitId")
  
  if repo_name == "" {
    repo_name, _ = h.getMandatoryUrlParam("repoName")
  }

  if commit_id == "" {
    commit_id, _ = h.getMandatoryUrlParam("commitId")
  }

  if repo_name == "" { 
    return base.HTTPErrorf(http.StatusBadRequest, "missing param - repoName")
  }

  if commit_id == "" { 
    return base.HTTPErrorf(http.StatusBadRequest, "missing param - commitId")
  }

  commit_attrs, err := h.server.workspaceApi.GetCommitAttrs(repo_name, commit_id)
  
  if err != nil {
    base.Error("[Handler.handleGetCommitAttrs] Error: ", repo_name, commit_id, err)
    return base.HTTPErrorf(http.StatusBadRequest, "Failed to retrieve commit attributes for repo, commit ID: %s %s %s", repo_name, commit_id, err.Error())
  }

  response = structs.Map(commit_attrs)
  h.writeJSON(response) 
  return nil
}


func (h *Handler) handleGetOrStartCommit() error {
  base.Debug("[Handler.handleGetOrStartCommit]")
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  repo_name := h.getQuery("repoName")
  branch_name := h.getQuery("branchName")
  commit_id := h.getQuery("commitId")

  //TODO: handle error
  commit_attrs, err := h.server.workspaceApi.InitCommit(repo_name, branch_name, commit_id)
  
  if err != nil {
    return base.HTTPErrorf(http.StatusBadRequest, "Failed to initialize a commit: " + err.Error())
  }

  base.Debug("[Handler.handlePostRepo] Repo created ", repo_name)
  
  response = structs.Map(commit_attrs)
  h.writeJSON(response) 
  return nil
} 


func (h *Handler) handleCloseCommit() error {
  base.Debug("[Handler.handleCloseCommit]")
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  repo_name := h.getQuery("repoName")
  branch_name := h.getQuery("branchName")
  commit_id := h.getQuery("commitId")

  err := h.server.workspaceApi.EndCommit(repo_name, branch_name, commit_id)
  
  if err != nil {
    return base.HTTPErrorf(http.StatusBadRequest, "Failed to close commit: %s", err)
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
  commit_size, err := h.server.workspaceApi.GetCommitSize(repo_name, branch_name, commit_id)
   
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
