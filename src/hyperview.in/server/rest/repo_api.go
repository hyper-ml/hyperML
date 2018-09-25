package rest


import(
  "net/http"
  "hyperview.in/server/base" 
  "hyperview.in/server/base/structs"  

  ws "hyperview.in/server/core/workspace"
)



func (h *Handler) handleGetRepo() error {
  var response map[string]interface{}
  repoName, _ := h.getMandatoryUrlParam("repoName")
  
  if repoName == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "[GetRepo] Repo name is mandatory")   
  }

  repo, err := h.server.workspaceApi.GetRepo(repoName)
  if err != nil {
    return base.HTTPErrorf(http.StatusBadRequest, "Failed to get repo: " + err.Error())
  }
  response = structs.Map(repo)
  h.writeJSON(response)

  return nil
} 

func (h *Handler) handlePostRepo() error {
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  repoName := h.getQuery("repoName")

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


func (h *Handler) handleExplodeRepo() error {
  var response map[string]interface{}
  repoName, _ := h.getMandatoryUrlParam("repoName")
  branchName := h.getQuery("branchName")
  
  if repoName == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "[GetRepo] Repo name is mandatory")   
  }

  repo, branch, commit, err := h.server.workspaceApi.ExplodeRepo(repoName, branchName)

  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }

  repo_msg := &ws.RepoMessage {
    Repo: repo,
    Branch: branch,
    Commit: commit,
  } 
  
  response = structs.Map(repo_msg)
  h.writeJSON(response)

  return nil

}


func (h *Handler) handleExplodeRepoAttrs() error {
  var response map[string]interface{}
  repoName, _ := h.getMandatoryUrlParam("repoName")
  branchName := h.getQuery("branchName")
  commitId := h.getQuery("commitId")
  
  if repoName == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "[GetRepo] Repo name is mandatory")   
  }

  repo_attrs, branch_attrs, commit_attrs, file_map, err := h.server.workspaceApi.ExplodeRepoAttrs(repoName, branchName, commitId)

  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }

  return_msg := &ws.RepoAttrsMessage {
    RepoAttrs: repo_attrs,
    BranchAttrs: branch_attrs,
    CommitAttrs: commit_attrs,
    FileMap: file_map,
  } 
  
  response = structs.Map(return_msg)
  h.writeJSON(response)

  return nil

}


func (h *Handler) handleGetModel() error {
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "[Handler.handleGetModel] Invalid method %s", h.rq.Method)
  }
  
  repo_name, _ := h.getMandatoryUrlParam("repoName")
  branch_name, _ := h.getMandatoryUrlParam("branchName")
  commit_id, _ := h.getMandatoryUrlParam("commitId")
  if repo_name == "" || commit_id =="" {
    return base.HTTPErrorf(http.StatusBadRequest, "missing_id_param: repo_name or commit id params is missing %s", repo_name, commit_id)   
  } 

  // create or get output repo for the flow
  repo, branch, commit, err := h.server.workspaceApi.GetModel(repo_name, branch_name, commit_id)

  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }

  model_response := &ws.RepoMessage {
    Repo: repo,
    Branch: branch,
    Commit: commit,
  } 
  
  response := structs.Map(model_response)
  h.writeJSON(response)

  return nil
}

func (h *Handler) handleGetOrCreateModel() error {
   
  if (h.rq.Method != "POST" && h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "[Handler.handleGetOrCreateModel] Invalid method %s", h.rq.Method)
  }
  
  repo_name, _ := h.getMandatoryUrlParam("repoName")
  branch_name, _ := h.getMandatoryUrlParam("branchName")
  commit_id, _ := h.getMandatoryUrlParam("commitId")

  if repo_name == "" || commit_id =="" {
    return base.HTTPErrorf(http.StatusBadRequest, "missing_id_param: repo_name or commit id params is missing %s", repo_name, commit_id)   
  } 

  // create or get output repo for the flow
  repo, branch, commit, err := h.server.workspaceApi.GetOrCreateModel(repo_name, branch_name, commit_id)

  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }

  model_response := &ws.RepoMessage {
    Repo: repo,
    Branch: branch,
    Commit: commit,
  } 
  
  response := structs.Map(model_response)
  h.writeJSON(response)

  return nil
}  

