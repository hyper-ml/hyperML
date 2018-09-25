package api_client

import (
  "fmt"
  "encoding/json"
  "hyperview.in/client/rest_client" 
  ws "hyperview.in/server/core/workspace"

)


func (c *ApiClient) GetCommitMap(repoName, commitId string) (*ws.FileMap, error) {
  var file_map ws.FileMap
  
  client, _ := rest_client.New(c.serverAddr, c.config.CommitMapUriPath)
  rq := client.Verb("GET")
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
  
  client, _   := rest_client.New(c.serverAddr, c.config.FlowAttrsUriPath)
  sub_path := "/" + flowId + "/log"
  log_req  := client.VerbSp("GET", sub_path)

  log_resp := log_req.Do()
  log_bytes, err:= log_resp.Raw() 
  fmt.Println("[ApiClient.RequestLog] log_bytes:", string(log_bytes), err)  
 
  if err != nil {
    return nil, err
  }
  
  return nil, nil
} 