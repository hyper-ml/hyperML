package rest


import( 
  "io/ioutil"
  "net/http"
  "encoding/json"
  flow_pkg "hyperview.in/server/core/flow"
  "hyperview.in/server/base" 
  "hyperview.in/server/base/structs"  
  ws "hyperview.in/server/core/workspace"

)

func (h *Handler) handleGetFlowAttrs() error {
  var err error 
  var response map[string]interface{}

  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  flow_id, _ := h.getMandatoryUrlParam("flowId")
  flow_attrs, err := h.server.flowServer.GetFlowAttr(flow_id)

  if err != nil {
    base.Log("[rest.flow.GetFlowAttrs] Failed to retrieve flow attributes. Please check params. ", flow_id, err)
    return err
  } 

  response = structs.Map(flow_attrs) 
  h.writeJSON(response)

  return nil 
}

// launches a flow for given repo -commit 
// async submission

func (h *Handler) handleLaunchFlow() error {

  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }  

  var response map[string]interface{}
  var raw_input []byte
  var err error 
  
  raw_input, err = ioutil.ReadAll(h.rq.Body)
  if err != nil {
    base.Log("[rest.flow.launchFlow] Invalid input for NewFlowLaunchRequest: ", err)
    return err
  }

  flow_msg := flow_pkg.FlowMessage{} 

  if err := json.Unmarshal(raw_input, &flow_msg); err != nil {
    base.Log("[rest.flow.launchFlow] Invalid JSON for NewFlowLaunchRequest: ", err)
    return err
  }
  repo_msg := flow_msg.Repos[0]
  repo_name := repo_msg.Repo.Name
  branch_name := repo_msg.Branch.Name
  commit_id := repo_msg.Commit.Id

  if repo_name == "" || commit_id == "" || flow_msg.CmdStr == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid method params")
  } 

  if branch_name != "" { 
    branch_name = "master"
  }

  if commit_id == "" {
    commit_id, err = h.server.workspaceApi.StartCommit(
      repo_name, branch_name)

    if err != nil {
      base.Log("[Handler.handleLaunchFlow] Failed to start commit: ", err)
      //h.writeJSON(response)   
      return base.HTTPErrorf(http.StatusBadRequest, err.Error())
    }
  } 

  flow_attrs, err := h.server.flowServer.LaunchFlow(repo_name, branch_name, commit_id, flow_msg.CmdStr)
 
  if err != nil {
    base.Log("[Handler.handleLaunchFlow] Failed to launch task: ", err)
    return base.HTTPErrorf(http.StatusBadRequest, err.Error())
  }

  flow_msg.Flow = &flow_attrs.Flow
  flow_msg.FlowStatusStr = flow_pkg.FlowStatusToString(flow_attrs.Status)
  flow_msg.Repos[0].Commit = &ws.Commit {Id: commit_id}
  
  response =  structs.Map(flow_msg)

  h.writeJSON(response)

  return nil
}
 
func (h *Handler) handleGetFlowOutput() error {
  
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  flow_id, _ := h.getMandatoryUrlParam("flowId")
  if flow_id == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flow_id)   
  } 

  f := flow_pkg.Flow { Id: flow_id }
  repo, branch, commit, err := h.server.flowServer.GetOutput(f)
  
  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }

  if repo == nil {
    return base.HTTPErrorf(http.StatusBadRequest, "Request failed ")    
  }
  
  out_repo := &ws.RepoMessage {
    Repo: repo,
    Branch: branch,
    Commit: commit,
  } 

  response := structs.Map(out_repo)
  h.writeJSON(response)
  return nil
}

func (h *Handler) handleGetOrCreateFlowOutput() error {
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  flow_id, _ := h.getMandatoryUrlParam("flowId")
  if flow_id == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flow_id)   
  } 

  f := flow_pkg.FlowRef(flow_id)
  repo, branch, commit, err := h.server.flowServer.GetOrCreateOutput(f)
  
  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }

  if repo == nil {
    return base.HTTPErrorf(http.StatusBadRequest, "Request failed ")    
  }

  out_repo := &ws.RepoMessage{
    Repo: repo,
    Branch: branch,
    Commit: commit,
  }

  response := structs.Map(out_repo)
  h.writeJSON(response)

  return nil
}


func (h *Handler) handleGetOrCreateFlowModel() error {
   if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  flow_id, _ := h.getMandatoryUrlParam("flowId")
  if flow_id == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flow_id)   
  } 

  f := flow_pkg.FlowRef(flow_id)
  repo, branch, commit, err := h.server.flowServer.GetOrCreateModel(f)
  
  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }

  if repo == nil {
    return base.HTTPErrorf(http.StatusBadRequest, "Request failed ")    
  }

  out_repo := &ws.RepoMessage{
    Repo: repo,
    Branch: branch,
    Commit: commit,
  }

  response := structs.Map(out_repo)
  h.writeJSON(response)

  return nil
}