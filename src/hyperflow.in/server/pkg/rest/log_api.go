package rest


import( 
  "io"
  "fmt"
  "time" 
  "strings"
  "net/http" 
  "github.com/gorilla/mux"  
  "github.com/gorilla/websocket"


  "hyperflow.in/server/pkg/base"    
  "hyperflow.in/server/pkg/base/structs"  

  ws "hyperflow.in/server/pkg/workspace"


)

func serveLogStream(sc *ServerContext, w http.ResponseWriter, r *http.Request) error{

  var err error 
  req_params := mux.Vars(r)
  task_id := req_params["taskId"]
  
  c, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    base.Error("websocket connection upgrade error:", err)
    return base.HTTPErrorf(http.StatusBadRequest, "Could not upgrade HTTP connection to websocket")
  }

  defer c.Close()
  reader, err := sc.flowAPI.LogStream(task_id)
  if err != nil {
    base.Error("failed to read log stream: ", err) 
    return base.HTTPErrorf(http.StatusBadRequest, "Failed to read log stream: " + err.Error())
  }

  msg_buf := make([]byte, 128)
  //todo: utils.GetBuffer()
  //defer utils.PutBuffer()
 
  for {
    n, err := reader.Read(msg_buf)
    if err != nil{
      if err == io.EOF {

        //should handle any remaining bytes.
        if err = c.WriteMessage(websocket.TextMessage, msg_buf[:n]); err != nil {
          break
        }

        break
      }

      base.Error(err.Error()) 
      
    }

    if err = c.WriteMessage(websocket.TextMessage, msg_buf[:n]); err != nil {
      return base.HTTPErrorf(http.StatusBadRequest, "Failed to write to websocket: " + err.Error())
    }
  
  }
  
  return nil
}

 
func (h *Handler) handleGetTaskLog() error {
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  task_id, _ := h.getMandatoryUrlParam("taskId") 
  log_path := h.server.logAPI.GetObjectPath(h.server.flowAPI.GetTaskLogPath(task_id))

  rs, err := h.server.logAPI.ReadSeeker(log_path, 0, 0)
  if err != nil {
    if  err != io.EOF {
      fmt.Println("Failed to fetch log from :", log_path)
      return base.HTTPErrorf(http.StatusBadRequest, "Error occurred while fetching log object.")
    }
  } 
  
  h.setHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", log_path))
  http.ServeContent(h.response, h.rq, log_path, time.Time{}, rs)
  
  return nil

}

func errLogObjectMissing(err error) bool {
  if err == nil {
    return false
  }

  err_string := err.Error()
  if strings.Contains(err_string, "object doesn't exist") {
    return true
  }
  return false
}

func (h *Handler) handleGetCommandLog() error {

  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  task_id, _ := h.getMandatoryUrlParam("taskId")  
  log_path := h.server.flowAPI.GetCommandLogPath(task_id)
  log_path  = h.server.logAPI.GetObjectPath(log_path)
  log_path  = log_path + ".log"

  rs, err := h.server.logAPI.ReadSeeker(log_path, 0, 0)
  
  if errLogObjectMissing(err) {
    log_path := h.server.logAPI.GetObjectPath(h.server.flowAPI.GetTaskLogPath(task_id))
    rs, err = h.server.logAPI.ReadSeeker(log_path, 0, 0)
  }

  if err != nil {
    if errLogObjectMissing(err) {
      return nil
    }

    if  err != io.EOF {
      return base.HTTPErrorf(http.StatusBadRequest, "Error occurred while fetching log object.")
    }
  } 
  
  h.setHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", log_path))
  http.ServeContent(h.response, h.rq, log_path, time.Time{}, rs)
  
  return nil
}


func (h *Handler) handlePostCommandLog() error {
  
  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  var response map[string]interface{} 
  task_id, err := h.getMandatoryUrlParam("taskId")
  
  if err != nil  {
    return base.HTTPErrorf(http.StatusBadRequest, "One of the params is missing: taskId")
  }
   
  if h.rq.Body != nil { 
    log_path := h.server.flowAPI.GetCommandLogPath(task_id)
    log_path = h.server.logAPI.GetObjectPath(log_path)
    base.Info("[Handler.handlePostTaskLog] log_path: ", log_path)
    
    obj_path, chksm, size, err := h.server.logAPI.SaveObject(log_path, "", h.rq.Body, false)
    fmt.Println("size: ", size)

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

    response = structs.Map(obj) 
    h.writeJSON(response)
    return nil

  } else {

    base.Debug("[Handler.handlePostFlowLog] Empty request body was sent for task: ", task_id)
    return base.HTTPErrorf(http.StatusBadRequest, "Request body is empty")
  } 
  return nil
}



