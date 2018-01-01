package rest


import(
  "net/http"
  "hyperview.in/server/base" 
  "hyperview.in/server/base/structs"  

  ws "hyperview.in/server/core/workspace"
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
    return base.HTTPErrorf(http.StatusBadRequest, "Failed to create repo: " + err.Error())
  }

  response = map[string]interface{}{
        "name": repoName,
    }
  h.writeJSON(response) 
  return nil
} 

func (h *Handler) handleGetOrCreateModelRepo() error {
   
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  repo_name, _ := h.getMandatoryUrlParam("repoName")
  branch_name, _ := h.getMandatoryUrlParam("branchName")
  commit_id, _ := h.getMandatoryUrlParam("commitId")

  if repo_name == "" || commit_id =="" {
    return base.HTTPErrorf(http.StatusBadRequest, "missing_id_param: repo_name or commit id params is missing %s", repo_name, commit_id)   
  } 

  // create or get output repo for the flow
  repo_attrs, err := h.server.workspaceApi.GetOrCreateModelRepo(repo_name, branch_name, commit_id)

  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }

  //start a fresh commit in master branch 
  commit_attrs, err:= h.server.workspaceApi.InitCommit(repo_attrs.Repo.Name, "master", "")

  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }
  var branch *ws.Branch
  if repo_attrs.Branches["master"] != nil {
    branch = repo_attrs.Branches["master"] 
  }

  m_repo_response := &ws.ModelRepoResponse {
    Repo: repo_attrs.Repo,
    Branch: branch,
    Commit: commit_attrs.Commit,
  } 
  
  response := structs.Map(m_repo_response)
  h.writeJSON(response)

  return nil
}  

