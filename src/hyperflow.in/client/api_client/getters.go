package api_client

import (
  "fmt"
  "encoding/json"
  "hyperflow.in/client/rest_client" 
  flow_pkg "hyperflow.in/server/pkg/flow"
  ws "hyperflow.in/server/pkg/workspace"

)


func (c *ApiClient) GetCommitMap(repoName, commitId string) (*ws.FileMap, error) {
  var file_map ws.FileMap
  
  client, _ := rest_client.New(c.serverAddr, c.config.CommitMapUriPath)
  rq := client.Verb("GET", c.jwt)
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

func (c *ApiClient) GetFlowAttrs(flowId string) (*flow_pkg.FlowAttrs, error) {
  
  client, _   := rest_client.New(c.serverAddr, c.config.FlowUriPath)
  sub_path := "/" + flowId 
  http_req := client.VerbSp("GET", sub_path, c.jwt)

  http_resp := http_req.Do()
  flow_json, err := http_resp.Raw()
  if err != nil {
    return nil, err
  }

  flow_attrs := flow_pkg.FlowAttrs{}
  err = json.Unmarshal(flow_json, &flow_attrs)
  if err != nil {
    return nil, err
  }

  return &flow_attrs, nil
}

func (c *ApiClient) GetFlowStatus(flowId string) (string, error) {
  
  client, _   := rest_client.New(c.serverAddr, c.config.FlowUriPath)
  sub_path := "/" + flowId  +"/status"
  http_req := client.VerbSp("GET", sub_path, c.jwt)
  
  http_resp := http_req.Do()
  flow_json, err := http_resp.Raw()
  if err != nil {
    return "", err
  }

  flow_msg := flow_pkg.FlowMessage{}
  err = json.Unmarshal(flow_json, &flow_msg)
  if err != nil {
    return "", err
  }

  return flow_msg.FlowStatusStr, nil
}

func (c *ApiClient) RequestLog(flowId string) ([]byte, error) {
  if flowId == "" {
    return nil, fmt.Errorf("Please submit a task Id with --task option")
  }

  client, _   := rest_client.New(c.serverAddr, c.config.FlowUriPath)
  sub_path := "/" + flowId + "/cmd_log"
  log_req  := client.VerbSp("GET", sub_path, c.jwt)
  
  log_resp := log_req.Do()
  log_bytes, err:= log_resp.Raw() 
  fmt.Println("Execution log: \n", string(log_bytes), err)  
 
  if err != nil {
    return nil, err
  }
  
  return nil, nil
} 