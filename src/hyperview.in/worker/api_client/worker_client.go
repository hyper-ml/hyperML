package api_client

// what: Client to access apis and process

import (
  "net/http" 
  "fmt"
  "net/url" 
  "encoding/json"
  //"strconv"
  "io"
  //"bytes" 
  //"io/ioutil"

  "hyperview.in/server/base" 

  "hyperview.in/worker/config" 
  "hyperview.in/worker/rest_client" 
  //tsk "hyperview.in/server/core/tasks"
  flw "hyperview.in/server/core/flow"
  ws "hyperview.in/server/core/workspace"

  //local_schema "hyperview.in/worker/schema"
)


type FsConfig struct {
  repoDir string
}

type ServerConfig struct {
  server_http string
  base_uri string
  repo_path string
}

type WorkerClient struct {

  // Rest Client to fetch info from server   
  RepoAttrs rest_client.Interface

  // Rest client for commit info
  CommitAttrs rest_client.Interface

  // Rest client for commit map
  CommitMap rest_client.Interface

  // Rest client for branch info
  BranchAttrs rest_client.Interface

  // File Info client
  FileAttrs rest_client.Interface

  // object client
  ContentIo rest_client.Interface

  // flow attributes
  FlowAttrs rest_client.Interface

  TaskAttrs rest_client.Interface

  TaskStatus rest_client.Interface

  //worker client

  WorkerAttrs rest_client.Interface


  // Virtual FS Client 
  vfs rest_client.Interface

  BaseUrl *url.URL
  Config *config.Config
  //TODO: add stats 
}

func NewWorkerClient(serverAddr string) (*WorkerClient, error) {
  var server_addr *url.URL

  c, err := config.ReadFromFile()

  if err != nil {
    fmt.Println("Failed to read config file")
    c = config.Default()
  } 

  if serverAddr != "" {
    server_addr, err = url.Parse(serverAddr)  
    
    if err != nil {
      base.Log("[NewWorkerClient] Invalid Server address string: ", serverAddr)
      return nil, err
    }

  } else {
    server_addr, err = url.Parse(c.DefaultServerAddr)   
    
     if err != nil {
      base.Log("[NewWorkerClient] Invalid Default server address string: ", c.DefaultServerAddr)
      return nil, err
    }
  }

  repo_attrs, err := rest_client.NewRESTClient(server_addr, c.RepoAttrsUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  commit_attrs, err := rest_client.NewRESTClient(server_addr, c.CommitAttrsUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  commit_map, err := rest_client.NewRESTClient(server_addr, c.CommitMapUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  branch_attr, err := rest_client.NewRESTClient(server_addr, c.BranchAttrsUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  file_attrs, err := rest_client.NewRESTClient(server_addr, c.FileAttrsUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }
 
  flow_attrs, err := rest_client.NewRESTClient(server_addr, c.FlowAttrsUriPath, http.DefaultClient)
  task_attrs, err := rest_client.NewRESTClient(server_addr, c.TaskAttrsUriPath, http.DefaultClient)
  //task_status, err := rest_client.NewRESTClient(server_addr, c.TaskStatusUriPath, http.DefaultClient)

  vfs, err := rest_client.NewRESTClient(server_addr, c.VfsUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  contents, err := rest_client.NewRESTClient(server_addr, c.ObjectUriPath, http.DefaultClient)

  worker_attrs, err := rest_client.NewRESTClient(server_addr, c.WorkerUriPath, http.DefaultClient)

  return &WorkerClient {
    RepoAttrs: repo_attrs,
    BranchAttrs: branch_attr,
    CommitAttrs: commit_attrs,
    FileAttrs: file_attrs,
    vfs: vfs,
    ContentIo: contents,
    FlowAttrs: flow_attrs,
    TaskAttrs: task_attrs,
    WorkerAttrs: worker_attrs,
    CommitMap: commit_map,
    BaseUrl: server_addr,
    Config: c,
    //TaskStatus: task_status,
  }, nil

} 
 

func (wc *WorkerClient) FetchBranchAttrs(repoName, branchName string) (*ws.BranchAttrs, error) {
  var branch_attr ws.BranchAttrs

  brq := wc.BranchAttrs.Get() 
  brq.Param("repoName", repoName)
  brq.Param("branchName", branchName)

  resp := brq.Do()
  body, err := resp.Raw()
  if err != nil {
    return nil, err
  } 

  branch_attr =  ws.BranchAttrs{}
  err = json.Unmarshal(body, &branch_attr) 

  return &branch_attr, nil
}

func (wc *WorkerClient) FetchCommitAttrs(repoName, commitId string) (*ws.CommitAttrs, error) {
  var commit_attrs ws.CommitAttrs

  crq := wc.CommitAttrs.Get() 
  crq.Param("repoName", repoName)
  crq.Param("commitId", commitId)
  resp := crq.Do()
  body, err := resp.Raw()
  if err != nil {
    return nil, err
  }
  err = json.Unmarshal(body, &commit_attrs)
  return &commit_attrs, nil
}

func (wc *WorkerClient) FetchCommitMap(repoName, commitId string) (*ws.FileMap, error) {
  var file_map ws.FileMap

  rq := wc.CommitMap.Get()
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
 
func (wc *WorkerClient) FetchFileAttrs(repoName string, commitId string, fpath string) (*ws.FileAttrs, error){
  var file_attrs *ws.FileAttrs 

  crq := wc.FileAttrs.Verb("GET") 
  crq.Param("repoName", repoName)
  crq.Param("commitId", commitId)
  crq.Param("path", fpath)

  resp := crq.Do()
  body, err := resp.Raw()
  if err != nil {
    return nil, err
  }
  
  err = json.Unmarshal(body, &file_attrs)
  return file_attrs, nil

}

func (wc *WorkerClient) FetchFlowAttrs(flowId string) (*flw.FlowAttrs, error){
  var flow_attrs *flw.FlowAttrs 
  flow_rq := wc.FlowAttrs.VerbSp("GET", flowId) 
  //flow_rq.Param("flowId", flowId) 

  resp := flow_rq.Do()
  flow_service_resp, err := resp.Raw()

  if err != nil {
    base.Debug("[WorkerClient.FetchFlowAttrs] Fetch Error: ", err)
    return nil, err
  }
  
  err = json.Unmarshal(flow_service_resp, &flow_attrs)
  return flow_attrs, nil
}




func (wc *WorkerClient) RegisterWorker(flowId string, taskId string, ip string) (*flw.WorkerAttrs, error) {
  req := wc.WorkerAttrs.VerbSp("POST", "register")
  req.Param("flowId", flowId)
  req.Param("taskId", taskId)
  req.Param("ip", ip)

  response, err := req.Do().Raw()

  if err != nil {
    base.Log("[WorkerClient.RegisterWorker] Failed to register worker for flow: ", flowId, taskId, err)
    return nil, err
  }
  
  worker_attrs:= flw.WorkerAttrs{}
  err = json.Unmarshal(response, &worker_attrs)

  return &worker_attrs, err
}



func (wc *WorkerClient) DetachWorker(flowId string, taskId string, workerId string) (error) {
  req := wc.WorkerAttrs.VerbSp("POST", "detach")
  req.Param("flowId", flowId)
  req.Param("taskId", taskId)
  req.Param("workerId", workerId)

  fmt.Println("req url: ", req.URL())
  _, err := req.Do().Raw()

  if err != nil {
    base.Log("[WorkerClient.UnRegisterWorker] Failed to detach worker from the flow: ", flowId, taskId, workerId)
    base.Log("[WorkerClient.UnRegisterWorker] Error: ", err)
    return err
  }
  
  return nil
}

func (wc *WorkerClient) GetModelRepo(srcRepoName, srcBranch, srcCommitId string) (*ws.Repo, *ws.Branch, *ws.Commit, error) {
  path:= srcRepoName + "/branch/" + srcBranch + "/commit/"+ srcCommitId +"/model_repo"
  req := wc.RepoAttrs.VerbSp("GET", path)

  base.Info("[WorkerClient.GetModelRepo] Model Repo creation url: ", req.URL())
  resp := req.Do()
  body, err := resp.Raw()
  if err != nil {
    return nil, nil, nil, err
  }

  model_response := &ws.ModelRepoResponse{}
  err = json.Unmarshal(body, model_response)
  
  base.Debug("[WorkerClient.GetModelRepo] Model Repo for Repo: ", model_response.Repo.Name)  
  return model_response.Repo, model_response.Branch, model_response.Commit, nil

}


func (wc *WorkerClient) PostLogWriter(taskId string) (io.WriteCloser, error) { 
  client, _ := rest_client.NewRESTClient(wc.BaseUrl, wc.Config.TasksUriPath, http.DefaultClient)

  r := client.VerbSp("POST", "/" + taskId + "/log")
  base.Info("[WorkerClient.PostLogWriter] log url:",  r.URL())

  hw := &httpObjectWriter {
    r: r,
  }  

  return hw, nil

}



