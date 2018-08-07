package client

// what: Client to access apis and process

import (
  "net/http" 
  "fmt"
  "net/url"
  "encoding/json"
  "hyperview.in/client/config"
  "hyperview.in/client/code_sync"
  "hyperview.in/client/rest_client"
  "hyperview.in/client/fs"
  "hyperview.in/client/schema"
 
)


type FsConfig struct {
  repoDir string
}

type ServerConfig struct {
  server_http string
  base_uri string
  repo_path string
}

type ApiClient struct {

  //Rest Client to fetch info from server   
  repoInfo rest_client.Interface

  //Rest Client to retrieve object metadata 
  codeSync *code_sync.Client
  
  //File system interface for local read/writers in the repo directory
  repoFS fs.FS

  //TODO: add stats 
}

func NewApiClient(repoDir string) (*ApiClient, error) {
  
  c, err := config.ReadFromFile()

  if err != nil {
    fmt.Println("Failed to read config file")
    c = config.Default()
  } 
  
  server_addr, err := url.Parse(c.ServerAddr) 
  if err != nil {
    return nil, err
  }

  repo_attrs, err := rest_client.NewRESTClient(server_addr, c.RepoUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  // TODO: find a use or remove this
  repo_fs := fs.NewFS(repoDir, c.Concurrency)

  code_sync := code_sync.NewClient(c.ServerAddr, repo_fs)

  return &ApiClient {
    repoInfo: repo_attrs,
    codeSync: code_sync,
    repoFS: repo_fs,
  }, nil

}

func (c *ApiClient) CloneRepo(repoName string) error {
  var err error
  repo_req := c.repoInfo.Verb("GET")
  repo_req.Param("repoName", repoName)
  resp := repo_req.Do()
  
  body, err := resp.Raw()

  if err != nil {
    return fmt.Errorf("Error occured in get_repo: %s", err)
  }

  getRepoResponse := schema.GetRepoResponse{}
  err = json.Unmarshal(body, &getRepoResponse)

  if err != nil {
    return fmt.Errorf("Failed to decode get_repo response %s", err)
  } 

  fmt.Println("Got Repo", getRepoResponse) 

  pull_req := &code_sync.PullRepoRequest {
    RepoName: repoName,
    CommitId: getRepoResponse.CommitId,
    FileMap: getRepoResponse.FileMap,
  }

  parallel_ops := 3
  s, err := c.codeSync.PullRemoteRepo(pull_req, parallel_ops)
  if err != nil {
    return err
  }
  fmt.Println("clone size:", s)
  return nil
}










