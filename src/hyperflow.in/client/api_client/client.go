package api_client

// what: Client to access apis and process

import ( 
  "io"
  "fmt"
  "bytes"
  "net/url"
  "io/ioutil"
  "encoding/json"

  "hyperflow.in/client/config"
  "hyperflow.in/client/rest_client"   
  "hyperflow.in/server/pkg/base"  
  
  flow_pkg "hyperflow.in/server/pkg/flow"
  ws "hyperflow.in/server/pkg/workspace" 
)
 
const (
  RestCallLimit int = 4
  outUrlPath = "/output"
  modelUrlPath = "/model"
  AuthUriPath = "auth"
  RepoUriPath = "repo" 
  ModelUriPath = "model" 
  OutUriPath = "out"
  RepoAttrsUriPath = "repo_attrs"
  DatasetUriPath = "dataset"
  BranchAttrsUriPath = "branch_attr"
  CommitUriPath = "commit"
  CommitAttrsUriPath = "commit_attrs"
  CommitMapUriPath = "commit_map"
  FileUriPath = "file"
  FileAttrsUriPath = "file_attrs"
  ObjectUriPath = "object"
  VfsUriPath = "vfs"
  WorkerUriPath = "worker"
  FlowUriPath = "flow"
  FlowAttrsUriPath = "flow_attrs"
  TaskAttrsUriPath = "tasks"
  TaskStatusUriPath = "task_status"
)    

type ApiClient struct { 
  serverAddr *url.URL
  config *config.UrlMap
  concurrency int
  jwt string
  //TODO: add stats 
}

func NewApiClient(addr *url.URL, c *config.UrlMap, jwt string) (*ApiClient, error) {
  
  return &ApiClient {
    serverAddr: addr,
    config: c,
    concurrency: RestCallLimit,
    jwt: jwt,
  }, nil

}


func (c *ApiClient) InitRepo(repoName string) error {
  client, _ := rest_client.New(c.serverAddr, RepoUriPath)
  repo_req := client.Verb("POST", c.jwt) 

  repo_req.Param("repoName", repoName)
  resp := repo_req.Do()
  _, err := resp.Raw()

  if err != nil {
    return fmt.Errorf("Failed while initializing repo: %s", err.Error())
  }

  base.Log("[InitRepo] Repo created: ", repoName)
  return nil

}

func (c *ApiClient) GetOutputRepo(flowId string) (*ws.Repo, *ws.Branch, *ws.Commit, error) {
  client, _ := rest_client.New(c.serverAddr, FlowUriPath)
  subpath := "/" + flowId + outUrlPath
  req := client.VerbSp("GET", subpath, c.jwt) 

  base.Info("[ApiClient.GetOutputRepo] Calling Url: ", req.URL())

  resp := req.Do()
  json_resp, err := resp.Raw()

  base.Info("[ApiClient.GetOutputRepo] json_resp: ", string(json_resp))
  repo_msg := ws.RepoMessage{}
  err = json.Unmarshal(json_resp, &repo_msg)

  if err != nil {
    return nil, nil, nil , err
  }

  base.Info("[ApiClient.GetModelByFlowId] repo_msg from server: ", string(json_resp))
  return repo_msg.Repo, repo_msg.Branch, repo_msg.Commit, nil
}

func (c *ApiClient) GetModelByFlowId(flowId string) (*ws.Repo, *ws.Branch, *ws.Commit, error) {

  client, _ := rest_client.New(c.serverAddr, FlowUriPath)
  subpath := "/" + flowId + modelUrlPath 
  req := client.VerbSp("GET", subpath, c.jwt)
  
  resp := req.Do()
  json_resp, err := resp.Raw()

  repo_msg := ws.RepoMessage{}
  err = json.Unmarshal(json_resp, &repo_msg)

  if err != nil {
    return nil, nil, nil , err
  }

  base.Info("[ApiClient.GetModelByFlowId] repo_msg from server: ", string(json_resp))
  return repo_msg.Repo, repo_msg.Branch, repo_msg.Commit, nil
}

func (c *ApiClient) GetFileObject(repoName, branchName, commitId, filePath string) (string, io.ReadCloser, error) {
  
  client, _ := rest_client.New(c.serverAddr, ObjectUriPath)
  f_request := client.Verb("GET", c.jwt)
  f_request.Param("repoName", repoName)
  f_request.Param("branchName", branchName)
  f_request.Param("commitId", commitId)
  f_request.Param("filePath", filePath) 

  return f_request.ResponseReader()
} 


func (c *ApiClient) GetCommit(repoName, branchName, commitId string) (*ws.Commit, error) {
  client, err := rest_client.New(c.serverAddr, CommitAttrsUriPath)
  req := client.Verb("GET", c.jwt)
  req.Param("repoName", repoName)
  req.Param("branchName", branchName)
  req.Param("commitId", commitId)

  resp := req.Do()
  json_body, err := resp.Raw()

  if err != nil {
    base.Log("[ApiClient.GetOrCreateCommit] Failed to retrieve an open repo commit: ", err)
    return nil, err
  }

  commit_attrs := &ws.CommitAttrs{}
  err = json.Unmarshal(json_body, &commit_attrs)

  return commit_attrs.Commit, nil
}

func (c *ApiClient) GetOrCreateCommit(repoName, branchName, commitId string) (*ws.Commit, error) {
  client, err := rest_client.New(c.serverAddr, CommitUriPath)
  req := client.Verb("GET", c.jwt)
  req.Param("repoName", repoName)
  req.Param("branchName", branchName)
  req.Param("commitId", commitId)

  resp := req.Do()
  json_body, err := resp.Raw()

  if err != nil {
    base.Log("[ApiClient.GetOrCreateCommit] Failed to retrieve an open repo commit: ", err)
    return nil, err
  }
  commit_attrs := &ws.CommitAttrs{}
  err = json.Unmarshal(json_body, &commit_attrs)

  return commit_attrs.Commit, nil
}


func (c *ApiClient) PutObjectWriter(repoName string, branchName string, commitId string, fpath string) (io.WriteCloser, error) {
  client, _ := rest_client.New(c.serverAddr, VfsUriPath)
  r := client.VerbSp("PUT", "put_file", c.jwt)
  
  r.Param("repoName", repoName)
  r.Param("branchName", branchName)

  r.Param("commitId", commitId)
  r.Param("path", fpath)
  
  hw := &httpWriter {
    r: r,
  }  

  return hw, nil
}


func (c *ApiClient) RunTask(rname, bname, headCommitId, cmdStr string, envVars map[string]string) (newFlow *flow_pkg.Flow, newCommit *ws.Commit, fnError error) {
  
  client, _ := rest_client.New(c.serverAddr, FlowUriPath)
  req := client.Verb("POST", c.jwt) 
  
  flow_msg := flow_pkg.FlowMessage {
    CmdStr: cmdStr,
    EnvVars: envVars,
    Repos: []*ws.RepoMessage{
      {
        Repo: &ws.Repo{
          Name: rname,
        },
        Branch: &ws.Branch{
          Name: bname,
        },
        Commit: &ws.Commit{
          Id: headCommitId,
        },
      },
    },
  }
 
  json_msg, _ := json.Marshal(&flow_msg) 
  _ = req.SetBodyReader(ioutil.NopCloser(bytes.NewReader(json_msg)))

  resp := req.Do()
  json_response, err := resp.Raw()  
  
  if err != nil {
    base.Error("[ApiClient.RunTask] Failed while calling launch flow end point: ", err)
    fnError = err
    return 
  }

  flow_resp :=  flow_pkg.FlowMessage{}
  err = json.Unmarshal(json_response, &flow_resp)
  if len(flow_resp.Repos) > 0 {
    // todo: need a better way to get master repo commit 
    newCommit = flow_resp.Repos[0].Commit
  }
  return flow_resp.Flow, newCommit, nil
} 


