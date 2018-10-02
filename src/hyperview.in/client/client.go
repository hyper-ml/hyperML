package client

import ( 
  "fmt"
  "strings"
  "net/url"
  "path/filepath"
  "hyperview.in/client/api_client"
  "hyperview.in/client/config"
  "hyperview.in/client/fs"
  "hyperview.in/server/base"
  flow_pkg "hyperview.in/server/core/flow"
  ws "hyperview.in/server/core/workspace"
)

const (
  OutDirName = "/out"
  SavedModelsDirName ="/saved_models"
)

type Client interface{
  InitRepo(repoName string) error
  InitBranch(repoName string, branchName string, headCommit string) error
  InitDataRepo(dir string, repoName string) error
  RunTask(repoName string, branchName string, commitId string, cmdStr string) (flow *flow_pkg.Flow, newCommit *ws.Commit, fnError error)   
  RequestLog(flowId string) ([]byte, error)
  PullResults(flowId string) (string, *ws.Repo, *ws.Branch, *ws.Commit, error)
  PullSavedModels(flowId string) (string, *ws.Repo, *ws.Branch, *ws.Commit, error)
  
  CloneRepo(rname string) (commitId string, fnError error)
  CloneBranch(rname, bname string) (commitId string, fnError error)
  CloneCommit(rname, bname, cid string) (commitId string, fnError error)

  PushRepo(rname, bname, cid string) (*ws.Commit, error)
}


func New(repoPath string) (Client, error) {

  c, err := config.ReadFromFile()
  server_string := c.DefaultServerAddr 

  if err != nil {
    base.Warn("Failed to read config file")
    c = config.Default()
  } 
  
  server_addr, err := url.Parse(server_string) 
  if err != nil {
    base.Error("[client.New] Failed to parse server URL: ", err)
    return nil, err
  }

  api, err:= api_client.NewApiClient(server_addr, c.UrlMap)
  if err != nil {
    base.Error("[client.New] Failed to create api client: ", err)
    return nil, err
  }
  return &client {
    repoPath: repoPath,
    serverAddr: server_addr,
    config: c,
    api: api,
  }, nil
}

type client struct {

  // actual os path to repo 
  repoPath string

  //File system interface for local read/writers in the repo directory
  repoFs *fs.RepoFs

  // rest API client
  api *api_client.ApiClient

  serverAddr *url.URL
  config *config.Config
}

func (c *client) AttachRepoFS(repoName string, branchName, commitId string) error {
  rfs := fs.NewRepoFs(c.repoPath, 0, repoName, branchName, commitId, c.api)
  c.repoFs = rfs
  return nil
}

func (c *client) InitRepo(repoName string) error {
  return c.api.InitRepo(repoName)
}

func (c *client) InitBranch(repoName string, branchName string, headId string) error {
  // TODO 
  return nil
}

func (c *client) InitDataRepo(dir string, repoName string) error {
  return c.api.InitDataRepo(dir, repoName)
}


func (c *client) RequestLog(flowId string) ([]byte, error) {
  return c.api.RequestLog(flowId)
}

func (c *client) getOutDir(flowId string) string {
  parent_dir := filepath.Join(c.repoPath, OutDirName)
  return filepath.Join(parent_dir, flowId)
}

// returns output repo details for local param storage 
//
func (c *client) PullResults(flowId string) (string, *ws.Repo, *ws.Branch, *ws.Commit, error) {
  
  out_dir :=  c.getOutDir(flowId)
  base.Info("[client.PullResults] out_dir: ", out_dir)

  // get out repo for the flow 
  out_repo, out_branch, out_commit, err := c.api.GetOutputRepo(flowId)
  
  if err != nil {
    
    if strings.Contains(err.Error(), "unexpected end of JSON input") { 
      base.Error("[client.PullResults] empty json response")
      return "", nil, nil, nil, fmt.Errorf("No results against this task.")
    }

    base.Error("[client.PullResults] Failed to retrieve out repo for given task: ", err)
    return "", nil, nil, nil, err
  }
  base.Debug("[client.PullResults] Out Repo, Branch and Commit: ", out_repo, out_branch, out_commit)
  // clone out in results folder repo 
  // hope all pans out 

  out_fs := fs.NewRepoFs(out_dir, 0, out_repo.Name, out_branch.Name, out_commit.Id, c.api)
  commit, err := out_fs.Clone()
  if err != nil {
    return "", nil, nil, nil, err
  }

  return out_dir, out_repo, out_branch, commit, nil
}
 

func (c *client) getModelDir(flowId string) string {
  parent_dir := filepath.Join(c.repoPath, SavedModelsDirName)
  return filepath.Join(parent_dir, flowId)
}

func (c *client) PullSavedModels(flowId string) (modelDir string, modelRepo *ws.Repo, modelBranch *ws.Branch, modelCommit *ws.Commit, fnError error) {

  model_dir := c.getModelDir(flowId)
  base.Info("[client.PullSavedModels] model_dir: ", model_dir)

  model_repo, model_branch, model_commit, err := c.api.GetModelByFlowId(flowId)
  
  switch {
  case err != nil:
    return model_dir, nil, nil, nil, err
  case model_repo.Name == "" && model_commit.Id == "":
    return model_dir, nil, nil, nil, base.ErrNullModelRepo() 
  }

  base.Info("[client.PullSavedModels] model Repo, Branch and Commit: ", model_repo, model_branch, model_commit)

  model_fs := fs.NewRepoFs(model_dir, 0, model_repo.Name, model_branch.Name, model_commit.Id, c.api)
  commit, err := model_fs.Clone()
  
  if err != nil {
    base.Error("[client.PullSavedModels] Model Repo clone failed: ", err)
    return model_dir, nil, nil, nil, err
  }

  return model_dir, model_repo, model_branch, commit, nil
}

func (c *client) CloneRepo(rname string) (commitId string, fnError error) {
  return c.CloneBranch(rname, "master")
}

func (c *client) CloneBranch(rname, bname string) (commitId string, fnError error) {
  return c.CloneCommit(rname, bname, "")
}

// todo: clone should retrieve commit id 
// and update local params  
func (c *client) CloneCommit(rname, bname, cid string) (commitId string, fnError error) {
  var commit *ws.Commit

  if fnError = c.AttachRepoFS(rname, bname, cid); fnError != nil {
    return 
  }

  if commit, fnError = c.repoFs.Clone(); fnError != nil {
    return 
  } 

  if commit != nil {
    return commit.Id, nil
  }
  return 
}

func (c *client) PushRepo(rname, bname, cid string) (*ws.Commit, error) {
  if err := c.AttachRepoFS(rname, bname, cid); err != nil {
    return nil, err
  }
  return c.repoFs.PushRepo()
}



// push code updates and then call run 
func (c *client) RunTask(repoName string, branchName string, commitId string, cmdStr string) (flow *flow_pkg.Flow, newCommit *ws.Commit, fnError error) {
  
  if _, err := c.PushRepo(repoName, branchName, commitId);  err != nil {
    base.Log("[client.RunTask] Failed to push code updates to server: ", err)
    fnError = err
    return  
  }

  return c.api.RunTask(repoName, branchName, commitId, cmdStr)
}


