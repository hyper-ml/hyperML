package rest

import ( 

  "net/http"
  "io/ioutil"
  "encoding/json"

  flow_pkg "hyperview.in/server/core/flow"
  "hyperview.in/server/base" 
  "hyperview.in/server/base/structs"   

)


//These are worker specific API functions

func (h *Handler) handleUpdateTaskStatus() error {
  if (h.rq.Method != "PATCH") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  // declare in/out variables 
  var response map[string]interface{}
  var raw_input []byte
  var err error

  worker_id, _ := h.getMandatoryUrlParam("workerId")
  base.Debug("[Handler.handleUpdateTaskStatus] worker Id: ", worker_id)  

  if worker_id == "" {
    return base.HTTPErrorf(http.StatusBadRequest, "No worker Id")
  }

  if h.rq.Body == nil {
    return base.HTTPErrorf(http.StatusBadRequest, "Empty status change request")
  }

  // read parameters
  raw_input, err = ioutil.ReadAll(h.rq.Body)
  
  // create a status change request from raw json body
  change_req := flow_pkg.TaskStatusChangeRequest{} 

  if err := json.Unmarshal(raw_input, &change_req); err != nil {
    base.Log("[rest.flow.UpdateTaskStatus] Invalid JSON for TaskStatusChangeRequest: ", err)
    return err
  }
  
  base.Log("[Handler.handleUpdateTaskStatus] New Status: ", change_req.TaskStatus)

  // call internal API
  worker := flow_pkg.Worker {Id: worker_id} 
  change_resp, err := h.server.flowServer.UpdateWorkerTaskStatus(worker, &change_req)

  if err != nil {
    base.Log("[Handler.handleUpdateTaskStatus] Failed task update: ", change_req)
    base.Log("[Handler.handleUpdateTaskStatus] ", err)
    return err 
  }

  response =  structs.Map(change_resp)
  h.writeJSON(response)

  return nil
}

func (h *Handler) handleRegisterWorker() error {

  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  flow_id := h.getQuery("flowId")
  task_id := h.getQuery("taskId")
  ip      := h.getQuery("ip")

  var response map[string]interface{}

  worker_attr, err := h.server.flowServer.RegisterWorker(flow_id, task_id, ip)
  if err != nil {
    return base.HTTPErrorf(http.StatusInternalServerError, err.Error() )
  }
  
  response =  structs.Map(worker_attr)
  
  h.writeJSON(response)
  return nil

}


func (h *Handler) handleDetachTaskWorker() error {
  
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  flow_id := h.getQuery("flowId")
  task_id := h.getQuery("taskId")
  worker_id := h.getQuery("workerId")

  if flow_id == "" || task_id == "" || worker_id == "" {
    return base.HTTPErrorf(http.StatusInternalServerError, "One of the params is missing: flowId, taskId, workerId")
  }

  return h.server.flowServer.DetachTaskWorker(worker_id, flow_id, task_id)
}