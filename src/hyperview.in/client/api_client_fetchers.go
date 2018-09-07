package client

import (
  "fmt"
  "encoding/json"
  ws "hyperview.in/server/core/workspace"

)


func (c *ApiClient) fetchCommitMap(repoName, commitId string) (*ws.FileMap, error) {
  var file_map ws.FileMap
  
  rq := c.CommitMap.Get()
  rq.Param("repoName", repoName)
  rq.Param("commitId", commitId)
  resp := rq.Do()

  body, err := resp.Raw()
  if err != nil {
    return nil, err
  }

  err = json.Unmarshal(body, &file_map)
  return &file_map, nil
}



func (c *ApiClient) RequestLog(flowId string) ([]byte, error) {
  log_req := c.flowIo.VerbSp("GET", "/" + flowId + "/log")
  log_resp := log_req.Do()
  log_bytes, err:= log_resp.Raw() 
  fmt.Println("[ApiClient.RequestLog] log_bytes:", string(log_bytes), err)  
 
  if err != nil {
    return nil, err
  }
  
  return nil, nil
} 