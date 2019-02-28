package rest


import( 
  "io/ioutil"
  "net/http"
  "encoding/json"
  flow_pkg "hyperflow.in/server/pkg/flow"
  "hyperflow.in/server/pkg/base" 
  "hyperflow.in/server/pkg/base/structs"  
  ws "hyperflow.in/server/pkg/workspace"

)

func (h *Handler) handleGetFlowAttrs() error {
  var err error 
  var response map[string]interface{}

  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  flow_id, _ := h.getMandatoryUrlParam("flowId")
  flow_attrs, err := h.server.flowAPI.GetFlowAttr(flow_id)

  if err != nil {
    base.Log("[rest.flow.GetFlowAttrs] Failed to retrieve flow attributes. Please check params. ", flow_id, err)
    return base.HTTPErrorf(http.StatusBadRequest, err.Error())
  } 

  response = structs.Map(flow_attrs) 
  h.writeJSON(response)

  return nil 
}

// launches a flow for given repo -commit 
// async submission

func (h *Handler) handleLaunchFlow() error { 

  var response map[string]interface{}
  var raw_input []byte
  var err error 

  if err := ValidateMethod(h.rq.Method, "POST"); err != nil {
     return err  
  }  

  raw_input, err = ioutil.ReadAll(h.rq.Body)
  if err != nil {
    return err
  }

  flow_msg := flow_pkg.FlowMessage{} 
  if err := json.Unmarshal(raw_input, &flow_msg); err != nil {
    return err
  }

  if len(flow_msg.Repos) == 0 {
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid method params: Empty repos")
  }

  repo_msg := flow_msg.Repos[0]
  repo_name := repo_msg.Repo.Name
  branch_name := repo_msg.Branch.Name
  commit_id := repo_msg.Commit.Id
  command_str := flow_msg.CmdStr
  env_vars := flow_msg.EnvVars

  switch {
  case repo_name == "":
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid method params: repo_name missing")
  case flow_msg.CmdStr == "": 
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid method params: command string missing")
  
  }

  if branch_name != "" { 
    branch_name = "master"
  }

  if commit_id == "" {
    cattrs, err := h.server.wsAPI.StartCommit(repo_name, branch_name)
    if err != nil {
      return base.HTTPErrorf(http.StatusBadRequest, err.Error())
    } else {
      commit_id = cattrs.Id()
    }
  } 

  flow_attrs, err := h.server.flowAPI.LaunchFlow(repo_name, branch_name, commit_id, command_str, env_vars)
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
 
func (h *Handler) handleGetOutputByFlow() error {
  
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  flow_id, _ := h.getMandatoryUrlParam("flowId")
  base.Info("[Handler.handleGetOutputByFlow] FlowId: ", flow_id)

  if flow_id == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flow_id)   
  } 

  f := flow_pkg.Flow { Id: flow_id }
  repo, branch, commit, err := h.server.flowAPI.GetOutput(f)
  
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

func (h *Handler) handleGetOrCreateOutputByFlow() error {
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  flow_id, _ := h.getMandatoryUrlParam("flowId")
  if flow_id == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flow_id)   
  } 

  f := flow_pkg.FlowRef(flow_id)
  repo, branch, commit, err := h.server.flowAPI.GetOrCreateOutput(f)
  
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


func (h *Handler) handleGetModelByFlow() error {
  
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  flow_id, _ := h.getMandatoryUrlParam("flowId")
  if flow_id == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flow_id)   
  } 

  f := flow_pkg.Flow { Id: flow_id }
  repo, branch, commit, err := h.server.flowAPI.GetModel(f)
  
  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }

  if repo == nil {
    return base.HTTPErrorf(http.StatusBadRequest, "Request failed ")    
  }
  
  model_repo := &ws.RepoMessage {
    Repo: repo,
    Branch: branch,
    Commit: commit,
  } 

  response := structs.Map(model_repo)
  h.writeJSON(response)
  return nil
}


func (h *Handler) handleGetOrCreateModelByFlow() error {
   if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  flow_id, _ := h.getMandatoryUrlParam("flowId")
  if flow_id == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flow_id)   
  } 

  f := flow_pkg.FlowRef(flow_id)
  repo, branch, commit, err := h.server.flowAPI.GetOrCreateModel(f)
  
  if err != nil {
   return base.HTTPErrorf(http.StatusBadRequest, "Request failed %s", err.Error())    
  }

  if repo == nil {
    return base.HTTPErrorf(http.StatusBadRequest, "Request failed ")    
  }

  model_repo := &ws.RepoMessage{
    Repo: repo,
    Branch: branch,
    Commit: commit,
  }

  response := structs.Map(model_repo)
  h.writeJSON(response)

  return nil
}


func (h *Handler) handleGetFlowStatus() error {
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  flow_id, _ := h.getMandatoryUrlParam("flowId")
  flow_attrs, err := h.server.flowAPI.GetFlowAttr(flow_id)

  if err != nil {
    base.Log("[rest.flow.GetFlowAttrs] Failed to retrieve flow attributes. Please check params. ", flow_id, err)
    return base.HTTPErrorf(http.StatusBadRequest, err.Error())
  } 

  flow_msg := &flow_pkg.FlowMessage {
    FlowStatusStr: flow_attrs.FlowStatus(),
  }

  response := structs.Map(flow_msg)
  h.writeJSON(response)
  
  return nil
}