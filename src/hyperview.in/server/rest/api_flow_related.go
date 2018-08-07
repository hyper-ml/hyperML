package rest


import(
//  "io/ioutil"
//  "encoding/json"
  "net/http"
  "hyperview.in/server/base" 
  "hyperview.in/server/base/structs"  
)

func (h *handler) GetFlowAttrs() error {

  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  flow_id := h.getQuery("flowId")
  if flow_id == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid flow Id - flow_id")
  }

  var response map[string]interface{}

  flow_attrs, err := h.server.flowServer.GetFlowAttr(flow_id)

  if err == nil {
    response = structs.Map(flow_attrs)
  } else {
    return err
  }

  h.writeJSON(response)
  return nil

}


func (h *handler) RegisterWorker() error {

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


func (h *handler) UnregisterWorker() error {
  
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  worker_id := h.getQuery("workerId")

  return h.server.flowServer.UnregisterWorker(worker_id)
}



