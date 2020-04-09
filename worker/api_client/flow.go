package api_client

import( 
  "bytes"
  "encoding/json"
  "github.com/hyper-ml/hyperml/server/pkg/base"
  "io/ioutil"
  flw "github.com/hyper-ml/hyperml/server/pkg/flow" 
  ws "github.com/hyper-ml/hyperml/server/pkg/workspace"

) 


func (w *WorkerClient) UpdateTaskStatus(workerId string, taskId string, tsr *flw.TaskStatusChangeRequest) (*flw.TaskStatusChangeResponse, error) {

  req := w.WorkerAttrs.VerbSp("PATCH", workerId + "/" + "task_status") 
  tsr_json, _ := json.Marshal(&tsr) 
  _ = req.SetBodyReader(ioutil.NopCloser(bytes.NewReader(tsr_json)))
  response := req.Do()

  tsr_resp_raw, err := response.Raw()
  if err != nil {
    base.Error("[WorkerClient.UpdateTaskStatus] http request Failed: ", err)
    return nil, err
  }
 
  tsr_resp :=  flw.TaskStatusChangeResponse{}
  err = json.Unmarshal(tsr_resp_raw, &tsr_resp)
 
  if err != nil {
    base.Error("[WorkerClient.UpdateTaskStatus] Failed to Unmarshal json response. Expected TaskAttr object. ", err)
    return nil, err
  }

  return &tsr_resp, nil
}
 
func (wc *WorkerClient) GetOrCreateFlowOutRepo(flowId string) (*ws.Repo, *ws.Branch, *ws.Commit, error) {
  req := wc.FlowAttrs.VerbSp("POST", flowId + "/" + "output")

  resp := req.Do()
  body, err := resp.Raw()
  if err != nil {
    return nil, nil, nil, err
  }

  out_response := &flw.FlowOutRepoResponse{}
  err = json.Unmarshal(body, out_response)
  
  return out_response.Repo, out_response.Branch, out_response.Commit, nil
}
