package rest


import(
  "net/http"
  "hyperview.in/server/base" 
  
)


func (h *Handler) handlePostRepo() error {
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  repoName := h.getQuery("repoName")
  
  //TODO: handle error
  _, err := h.server.workspaceApi.InitRepo(repoName)
  if err != nil {
    response = map[string]interface{}{
        "status": "Failed to create repo" + err.Error(),
    } 
  }

  base.Debug("[Handler.handlePostRepo] Repo created ", repoName)
  
  response = map[string]interface{}{
        "name": repoName,
    }
  h.writeJSON(response) 
  return nil
} 
