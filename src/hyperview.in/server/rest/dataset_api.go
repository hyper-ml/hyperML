package rest

import(
  "net/http"
  "hyperview.in/server/base" 
  
)


func (h *Handler) handlePostDataSet() error {
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  repoName := h.getQuery("repoName")
  
  //TODO: handle error
  err := h.server.workspaceApi.CreateDataset(repoName)
  if err != nil {
    response = map[string]interface{}{
        "status": "Failed to create dataset" + err.Error(),
    } 
  }

  base.Debug("[Handler.handlePostDataSet] Dataset created ", repoName)
  
  response = map[string]interface{}{
        "name": repoName,
    }

  h.writeJSON(response) 
  return nil
} 