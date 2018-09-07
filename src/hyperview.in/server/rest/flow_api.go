package rest


import( 
  "io"
  "fmt"
  "time"
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

  flow_req := flow_pkg.NewFlowLaunchRequest{} 

  if err := json.Unmarshal(raw_input, &flow_req); err != nil {
    base.Log("[rest.flow.launchFlow] Invalid JSON for NewFlowLaunchRequest: ", err)
    return err
  }

  if flow_req.Repo.Name == "" || flow_req.CmdString == "" {
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid method params")
  } 

  if flow_req.Commit.Id == "" {
    commit_id, err := h.server.workspaceApi.StartCommit(flow_req.Repo.Name, "master")
    if err != nil {
      base.Log("[Handler.handleLaunchFlow] Failed to start commit: ", err)
      h.writeJSON(response)   
      return nil
    }
    flow_req.Commit  = ws.Commit { Id: commit_id }
  } 

  flow_resp, err := h.server.flowServer.LaunchFlow(flow_req)
  
  base.Log("flow_resp in handler: ", flow_resp)
  // wait for a second or when flow is ready ?
  
  // keep track if possible 
  response =  structs.Map(flow_resp)

  h.writeJSON(response)

  return nil
}




func (h *Handler) handleGetFlowLog() error {
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  flow_id, err := h.getMandatoryUrlParam("flowId")
  if err != nil  {
    return base.HTTPErrorf(http.StatusInternalServerError, "One of the params is missing: flowId")
  }

  log_path :=  h.server.flowServer.GetFlowLogPath(flow_id)
  rs, err:= h.server.objectAPI.ReadSeeker(log_path, 0, 0)

  if err != nil {
    if  err != io.EOF {
      base.Log("Failed to fetch log object: ", log_path, err)
      return base.HTTPErrorf(http.StatusBadRequest, "Error occurred while fetching log object.")
    }
    base.Log("[Handler.handleGetFlowLog]: Reached EOF. Nothing more to read. ", log_path)
    //return nil
  }

  h.setHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", log_path))
  http.ServeContent(h.response, h.rq, log_path, time.Time{}, rs)
  
  return nil
}


