package api_client

import (
  "fmt"
  "hyperflow.in/client/rest_client" 
  "hyperflow.in/server/pkg/base"
)

func (c *ApiClient) InitDataRepo(dir string, repoName string) error {  
  return c.createTypedRepo(repoName, c.config.DatasetUriPath)
}

func (c *ApiClient) InitModelRepo(dir, repoName string) error {  
  return c.createTypedRepo(repoName, c.config.ModelUriPath)
}
 

func (c *ApiClient) InitOutRepo(dir, repoName string) error {  
  return c.createTypedRepo(repoName, c.config.OutUriPath)
}

func (c *ApiClient) InitBranch(repoName string, branchName string, headId string) error {
  return fmt.Errorf("unimplemented feature")
}


func (c *ApiClient) createTypedRepo(repoName, uriPath string) error {
  fmt.Println("uri Path: ", uriPath)
  client, _   := rest_client.New(c.serverAddr, uriPath)
  repo_req := client.Verb("POST", c.jwt)
  repo_req.Param("repoName", repoName)
  resp := repo_req.Do()
  _, err := resp.Raw()

  if err != nil {
    return fmt.Errorf("[createTypedRepo] Failed while initializing repo: %s", err)
  }

  base.Log("[createTypedRepo] Repo created: ", repoName)
  return nil
}



 
/*
func (c *ApiClient) CloneRepo(repoName, branchName string) (commitId string, err error) {
  var commit_id string

  client   := rest_client.New(c.serverAddr, c.config.RepoAttrsUriPath)

  sub_path:= "/" + repoName + "/explode"
  repo_req := client.VerbSp("GET", sub_path) 
  repo_req.Param("branchName", branchName)

  resp := repo_req.Do()
  
  body, err := resp.Raw()

  if err != nil {
    return commit_id, pullRepoError(err)
  }
  repo_msg := ws.RepoAttrsMessage{}

  err = json.Unmarshal(body, &repo_msg)

  if err != nil {
    return commit_id, pullRepoError(err)
  } 

  fmt.Println("Got Repo", repo_msg) 

  pull_req := &code_sync.PullRepoRequest {
    RepoName: repo_msg.Repo.Name,
    BranchName: repo_msg.Branch.Name,
    CommitId: repo_msg.Commit.Id,
    FileMap:  repo_msg.FileMap,
  }

  parallel_ops := 3
  s, err := c.codeSync.PullRemoteRepo(pull_req, parallel_ops)
  if err != nil {
    return commit_id, err
  }
  fmt.Println("clone size:", s)
  commitId = getRepoResponse.CommitId
  return
}*/