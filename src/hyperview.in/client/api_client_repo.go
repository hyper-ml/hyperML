package client

import (
  "fmt"
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

func (c *ApiClient) InitBranch(repoName string, branchName string, headCommit string) error {

  return nil
}
  
func (c *ApiClient) createDataset(repoName string) error {
  repo_req := c.dataSetInfo.Post()
  repo_req.Param("repoName", repoName)
  resp := repo_req.Do()
  _, err := resp.Raw()

  if err != nil {
    return fmt.Errorf("[CreateDataset] Failed while initializing repo: %s", err)
  }

  base.Log("[CreateDataset] Repo created: ", repoName)
  return nil
}