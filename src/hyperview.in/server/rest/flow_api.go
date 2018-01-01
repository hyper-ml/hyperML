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
  var branch_name string
  
  raw_input, err = ioutil.ReadAll(h.rq.Body)
  if err != nil {
    base.Log("[rest.flow.launchFlow] Invalid input for NewFlowLaunchRequest: ", err)
    return err
  }

  flow_req := flow_pkg.NewFlowLaunchRequest{} 

  if err := json.Unmarshal(raw_input, &flow_req); err != nil {
    base.Log("[rest.flow.launchFlow] Invalid JSON for NewFlowLaunchRequest: ", err)
    return err
  }

  if flow_req.Repo.Name == "" || flow_req.CmdString == "" {
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid method params")
  } 
  if flow_req.Branch.Name != "" {
    branch_name = flow_req.Branch.Name
  } else {
    branch_name = "master"
  }

  if flow_req.Commit.Id == "" {
    commit_id, err := h.server.workspaceApi.StartCommit(
      flow_req.Repo.Name, branch_name)

    if err != nil {
      base.Log("[Handler.handleLaunchFlow] Failed to start commit: ", err)
      //h.writeJSON(response)   
      return base.HTTPErrorf(http.StatusBadRequest, err.Error())
    }
    flow_req.Commit  = ws.Commit { Id: commit_id }
  } 

  flow_resp, err := h.server.flowServer.LaunchFlow(flow_req)
  if err != nil {
    base.Log("[Handler.handleLaunchFlow] Failed to launch task: ", err)
    return base.HTTPErrorf(http.StatusBadRequest, err.Error())
  }

  base.Log("flow_resp in handler: ", flow_resp)
  // wait for a second or when flow is ready ?
  
  // keep track if possible 
  response =  structs.Map(flow_resp)

  h.writeJSON(response)

  return nil
}





func (h *Handler) handleGetOrCreateOutRepo() error {
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  flow_id, _ := h.getMandatoryUrlParam("flowId")
  if flow_id == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid flowId params %s", flow_id)   
  } 

  // create or get output repo for the flow
  repo_attrs, err := h.server.flowServer.GetOrCreateOutRepo(flow_pkg.Flow {Id: flow_id})
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

  out_response := &flow_pkg.FlowOutRepoResponse {
    Repo: repo_attrs.Repo,
    Branch: branch,
    Commit: commit_attrs.Commit,
  } 
  
  response := structs.Map(out_response)
  h.writeJSON(response)

  return nil
}  
