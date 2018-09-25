package client

import (
  "net/url"
  "io/ioutil"
  "bytes"
  "hyperview.in/client/api_client"
  "hyperview.in/client/config"
  "hyperview.in/client/fs"
  flow_pkg "hyperview.in/server/core/flow"
  ws "hyperview.in/server/core/workspace"
)

type Client interface{
  InitRepo(repoName string) error
  InitBranch(repoName string, branchName string, headCommit string) error
  InitDataRepo(dir string, repoName string) error
  RunTask(repoName string, branchName string, commitId string, cmdStr string) (flow_id string, openCommitId string, task_status string, fnError error)
  RequestLog(flowId string) ([]byte, error)
  PullResults(flowId string) error
  PullSavedModels(flowId string) error
  
  CloneRepo() error
  PushRepo(repoName, branchName, commitId string) (*ws.Commit, error)
}


func New(repoPath string) Client {

  c, err := config.ReadFromFile()
  server_string := c.DefaultServerAddr

  if err != nil {
    fmt.Println("Failed to read config file")
    c = config.Default()
  } 
  
  server_addr, err := url.Parse(server_string) 
  if err != nil {
    return nil, err
  }
  api:= api_client.NewApiClient(addr, c.Urlmap)

  return &client {
    repoPath: repoPath,
    serverAddr: server_addr,
    config: c,
  }
}

type client struct {

  // actual os path to repo 
  repoPath string

  //File system interface for local read/writers in the repo directory
  repoFS fs.RepoFs

  // rest API client
  api *api_client.ApiClient

  serverAddr *url.URL
  config *config.Config
}

func (c *client) AttachRepoFS(repoName string, branchName, commitId string) error {
  rfs := NewRepoFs(c.repoPath, 0, repoName, branchName, commitId, c.api)
  c.repoFs = rfs
  return
}

func (c *client) InitRepo(repoName string) error {
  return c.api.InitRepo(repoName)
}

func (c *client) InitBranch(repoName string, branchName string, headId string) error {
  // TODO 
  return 
}

func (c *client) InitDataRepo(dir string, repoName string) error {
  return c.api.InitDataRepo(dir, repoName)
}


func (c *client) RunTask(repoName string, branchName string, commitId string, cmdStr string) (flow_id string, openCommitId string, task_status string, fnError error) {
  return c.api.RunTask(repoName, branchName, commitId, cmdStr)
}

func (c *client) RequestLog(flowId string) ([]byte, error) {
  return c.api.RequestLog(flowId)
}


func (c *client) PullResults(flowId string) error {
  // to do :
  fmt.Println("[client.PullResults] TODO")
  return 
}

func (c *client) PullSavedModels(flowId string) error {
  // to do :
  fmt.Println("[client.PullSavedModels] TODO")
  return 
}
  
func (c *ApiClient) CloneRepo(rname, bname, cid string) error {
  
  if err := c.AttachRepoFS(rname, bname, cid); err != nil {
    return err
  }

  if err := c.repoFs.Clone(); err != nil {
    return err
  }
  return nil
}

func (c *client) PushRepo(repoName, branchName, commitId string) (*ws.Commit, error) {
  if err := c.AttachRepoFS(rname, bname, cid); err != nil {
    return err
  }
  return c.api.PushRepo()
}



// push code updates and then call run 
func (c *ApiClient) RunTask(repoName string, branchName string, commitId string, cmdStr string) (flowId string, finalCommitId string, flowStatus string, fnError error) {
  var err error 
  var commit *ws.Commit 
  
  commit, err = c.PushRepo(repoName, branchName, commitId) 
  
  if err != nil {
    base.Log("[ApiClient.RunTask] Failed to push code updates to server: ", err)
    return "", "", "", err
  }

  flow_msg := flow_pkg.FlowMessage {
    CmdStr: cmdStr,
    Repos: []{
      *RepoMessage{
        Repo: {
          Name: repoName,
        },
        Branch: {
          Name: branchName,
        },
        Commit: {
          Id: commitId,
        },
      },
    },
  }


  client, _ := rest_client.New(c.serverAddr, c.config.FlowUriPath)
  req := client.Verb("POST") 

  json_msg, _ := json.Marshal(&flow_msg) 
  _ = api_req.SetBodyReader(ioutil.NopCloser(bytes.NewReader(json_msg)))

  resp := req.Do()
  json_response, err := resp.Raw()  

  flow_resp :=  flow_pkg.FlowMessage{}
  err = json.Unmarshal(json_response, &flow_resp)
  
  base.Log("[RunTask] Flow Id: ", flow_resp.Flow.Id)
  if flow_resp.Flow != nil {
    flowId = flow_resp.Flow.Id
  }
  
  if flow_resp.Commit != nil {
    finalCommitId = flow_resp.Commit.Id
  }
  
  flowStatus = flow_resp.FlowStatusStr
  return
}


