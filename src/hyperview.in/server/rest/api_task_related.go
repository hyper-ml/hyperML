package rest

import(
  "io/ioutil"
  "encoding/json"
  "net/http"
  "hyperview.in/server/base" 
  "hyperview.in/server/base/structs" 
  tsk "hyperview.in/server/core/tasks"
)


func (h *handler) CreateTask() error {

  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  var response map[string]interface{}
  
  if h.rq.Body == nil {
    h.writeJSON(response)
    return nil
  }


  config_data, err := ioutil.ReadAll(h.rq.Body)
  task_config := &tsk.TaskConfig{}
  err = json.Unmarshal(config_data, &task_config)

  task_attrs, err := h.server.tasker.CreateTask(task_config)

  if err == nil {
    response =  structs.Map(task_attrs)
  } else {
      return err
  }

  h.writeJSON(response)
  return nil

}


func (h *handler) GetTaskAttrs() error {

  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  
  task_id := h.getQuery("taskId")
  if task_id == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid task Id - task_id")
  }

  var response map[string]interface{}

  task_attrs, err := h.server.tasker.GetTaskAttrs(task_id)

  if err == nil {
    response = structs.Map(task_attrs)
  } else {
    return err
  }

  h.writeJSON(response)
  return nil

}