package rest


import( 
  "io"
  "fmt"
  "time" 
  "net/http" 
  "hyperview.in/server/base"    
  "hyperview.in/server/base/structs"  

  ws "hyperview.in/server/core/workspace"
)

func (h *Handler) handleGetTaskLog() error {
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  task_id, err := h.getMandatoryUrlParam("taskId")
  if err != nil  {
    return base.HTTPErrorf(http.StatusInternalServerError, "One of the params is missing: taskId")
  }

  base.Info("[Handler.handleGetTaskLog] task_id: ", task_id)
  
  log_path := h.server.flowServer.GetTaskLogPath(task_id)
  log_path = h.server.logApi.GetObjectPath(log_path)

  base.Info("[Handler.handleGetTaskLog] log_path: ", log_path)
  rs, err:= h.server.logApi.ReadSeeker(log_path, 0, 0)

  if err != nil {
    if  err != io.EOF {
      base.Error("[Handler.handleGetTaskLog]Failed to fetch log object: ", log_path, err)
      return base.HTTPErrorf(http.StatusBadRequest, "Error occurred while fetching log object.")
    }
    base.Debug("[Handler.handleGetTaskLog]: Reached EOF. Nothing more to read. ", log_path)
    //return nil
  }

  h.setHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", log_path))
  http.ServeContent(h.response, h.rq, log_path, time.Time{}, rs)
  
  return nil
}


func (h *Handler) handlePostTaskLog() error {
  
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  var response map[string]interface{} 
  task_id, err := h.getMandatoryUrlParam("taskId")
  
  if err != nil  {
    return base.HTTPErrorf(http.StatusBadRequest, "One of the params is missing: taskId")
  }
   
  if h.rq.Body != nil { 
    log_path := h.server.flowServer.GetTaskLogPath(task_id)

    base.Info("[Handler.handlePostTaskLog] log_path: ", log_path)
    obj_path, chksm, size, err := h.server.logApi.SaveObject(log_path, "", h.rq.Body, false)
   
    if err != nil {
      base.Debug("[Handler.handlePostFlowLog] Error occurred writing log on server ", task_id, err)
      return base.HTTPErrorf(http.StatusBadRequest, "Failed to save log object: ", err)
    }
    
    if size == 0 {
      base.Debug("[Handler.handlePostFlowLog] The input file is empty for flow log request: ", task_id)
      return base.HTTPErrorf(http.StatusBadRequest, "Input file is empty")
    }

    obj := &ws.Object{
      Path: obj_path, 
      CheckSum: chksm,
      Size: int(size) }
    base.Debug("[Handler.handlePostFlowLog] obj: ", obj_path)
    response = structs.Map(obj) 
    h.writeJSON(response)
    return nil

  } else {

    base.Debug("[Handler.handlePostFlowLog] Empty request body was sent for task: ", task_id)
    return base.HTTPErrorf(http.StatusBadRequest, "Request body is empty")
  } 
  return nil
}
