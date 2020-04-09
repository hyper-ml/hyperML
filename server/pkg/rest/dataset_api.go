package rest

import(
  "net/http"
  "github.com/hyper-ml/hyperml/server/pkg/base" 
  
)


func (h *Handler) handlePostDataSet() error {
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  repoName := h.getQuery("repoName")
  
  //TODO: handle error
  _, err := h.server.wsAPI.CreateDataset(repoName)
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