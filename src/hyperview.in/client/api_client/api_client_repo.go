package api_client

import (
  "fmt"
  "hyperview.in/client/rest_client" 
  "hyperview.in/server/base"
)

func (c *ApiClient) InitDataRepo(dir string, repoName string) error {
  err := c.createDataset(repoName)
  if err != nil {
    base.Log("[InitDataRepo] Failed to create dataset record: ", err)
    return err
  }
 
  return nil
}

func (c *ApiClient) InitBranch(repoName string, branchName string, headId string) error {

  return nil
}
  
func (c *ApiClient) createDataset(repoName string) error {
  
  client, _   := rest_client.New(c.serverAddr, c.config.DatasetUriPath)

  repo_req := client.Verb("POST")
  repo_req.Param("repoName", repoName)
  resp := repo_req.Do()
  _, err := resp.Raw()

  if err != nil {
    return fmt.Errorf("[CreateDataset] Failed while initializing repo: %s", err)
  }

  base.Log("[CreateDataset] Repo created: ", repoName)
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