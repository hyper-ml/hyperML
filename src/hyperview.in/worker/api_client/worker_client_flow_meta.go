package api_client

import( 
  "bytes"
  "encoding/json"
  "hyperview.in/server/base"
  "io/ioutil"
  flw "hyperview.in/server/core/flow" 
  
) 


func (w *WorkerClient) UpdateTaskStatus(workerId string, taskId string, tsr *flw.TaskStatusChangeRequest) (*flw.TaskStatusChangeResponse, error) {
  base.Log("[WorkerClient.UpdateTaskStatus] WorkerId: ", workerId)
  req := w.WorkerAttrs.VerbSp("PATCH", workerId + "/" + "task_status")
  
  base.Log("[WorkerClient.UpdateTaskStatus] Url: ", req.URL())

  tsr_json, _ := json.Marshal(&tsr) 
  _ = req.SetBodyReader(ioutil.NopCloser(bytes.NewReader(tsr_json)))

  response := req.Do()

  base.Debug("[WorkerClient.UpdateTaskStatus] Updating Status to: ", tsr.TaskStatus)  
  
  tsr_resp_raw, err := response.Raw()

  if err != nil {
    base.Log("[WorkerClient.UpdateTaskStatus] http request Failed: ", err)
    return nil, err
  }
 
  tsr_resp :=  flw.TaskStatusChangeResponse{}
  err = json.Unmarshal(tsr_resp_raw, &tsr_resp)
 
  if err != nil {
    base.Log("[WorkerClient.UpdateTaskStatus] Failed to Unmarshal json response. Expected TaskAttr object. ", err)
    return nil, err
  }

  return &tsr_resp, nil

}

