package rest


import(
  "net/http"
  "hyperview.in/server/base" 
  "hyperview.in/server/base/structs"  
  
)


func (h *Handler) handleGetOrStartCommit() error {
  base.Debug("[Handler.handleGetOrStartCommit]")
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  repoName := h.getQuery("repoName")
  branchName := h.getQuery("branchName")

  //TODO: handle error
  commit_attrs, err := h.server.workspaceApi.InitCommit(repoName, branchName)
  
  if err != nil {
    response = structs.Map(commit_attrs)
    return err
  }

  base.Debug("[Handler.handlePostRepo] Repo created ", repoName)
  
  response = structs.Map(commit_attrs)
  h.writeJSON(response) 
  return nil
} 
