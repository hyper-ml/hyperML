package api_client

// what: Client to access apis and process

import ( 
  "fmt"
  "net/url"
  "io"
  "encoding/json"


  "hyperview.in/client/config"
  "hyperview.in/client/rest_client"  

  "hyperview.in/server/base"

  ws "hyperview.in/server/core/workspace"
   

 
)
    
type ApiClient struct { 
  serverAddr *url.URL
  config *config.UrlMap
  concurrency int
  //TODO: add stats 
}

func NewApiClient(addr *url.URL, c *config.UrlMap, parallel int) (*ApiClient, error) {

  return &ApiClient {
    serverAddr: addr,
    config: c,
    concurrency: parallel,
  }, nil

}

func (c *ApiClient) InitRepo(repoName string) error {
  client, _ := rest_client.New(c.serverAddr, c.config.RepoUriPath)
  repo_req := client.Post()
  repo_req.Param("repoName", repoName)
  resp := repo_req.Do()
  _, err := resp.Raw()

  if err != nil {
    return fmt.Errorf("Failed while initializing repo: %s", err.Error())
  }

  base.Log("[InitRepo] Repo created: ", repoName)
  return nil

}
 

func (c *ApiClient) GetFileObject(repoName, branchName, commitId, filePath string) (io.ReadCloser, error) {
  client, _ := rest_client.New(c.serverAddr, c.config.ObjectUriPath)
  f_request := client.Verb("GET")
  f_request.Param("repoName", repoName)
  f_request.Param("branchName", branchName)
  f_request.Param("commitId", commitId)
  f_request.Param("filePath", filePath)

  return f_request.ReadResponse()
} 

func (c *ApiClient) GetOrCreateCommit(repoName, branchName, commitId string) (*ws.Commit, error) {
  client, err := rest_client.New(c.serverAddr, c.config.CommitUriPath)
  req := client.Verb("GET")
  req.Param("repoName", repoName)
  req.Param("branchName", branchName)
  req.Param("commitId", commitId)

  resp := req.Do()
  json_body, err := resp.Raw()

  if err != nil {
    base.Log("[ApiClient.PushRepo] Failed to retrieve an open repo commit: ", err)
    return nil, err
  }
  commit_attrs := &ws.CommitAttrs{}
  err = json.Unmarshal(json_body, &commit_attrs)

  return commit_attrs.Commit, nil
}


func (c *ApiClient) PutObjectWriter(repoName string, branchName string, commitId string, fpath string) (io.WriteCloser, error) {
  client, _ := rest_client.New(c.serverAddr, c.config.VfsUriPath)
  r := client.VerbSp("PUT", "put_file")
  
  r.Param("repoName", repoName)
  r.Param("branchName", branchName)

  r.Param("commitId", commitId)
  r.Param("path", fpath)
  
  hw := &httpWriter {
    r: r,
  }  

  return hw, nil
}
 
