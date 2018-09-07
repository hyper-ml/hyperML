package client

// what: Client to access apis and process

import (
  "os"
  "net/http" 
  "fmt"
  "net/url"
  "io"
  "io/ioutil"
  "bytes"
  "encoding/json"
  "sync/atomic"
  "golang.org/x/sync/errgroup"
  filepath_pkg "path/filepath"


  "hyperview.in/client/config"
  "hyperview.in/client/code_sync"
  "hyperview.in/client/rest_client"
  "hyperview.in/client/fs"
  "hyperview.in/client/schema"

  "hyperview.in/server/base"

  flow_pkg "hyperview.in/server/core/flow"
  ws "hyperview.in/server/core/workspace"
  
  "hyperview.in/client/utils"  

 
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

  // rest client 
  dataSetInfo rest_client.Interface

  // commit rest client
  commitAttrs rest_client.Interface

  //flow end point client
  flowIo rest_client.Interface

  // Rest client for commit map
  CommitMap rest_client.Interface

  // object client
  ContentIo rest_client.Interface

  vfs rest_client.Interface

  //Rest Client to retrieve object metadata 
  codeSync *code_sync.Client
  
  //File system interface for local read/writers in the repo directory
  repoFS fs.FS

  ServerAddr *url.URL
  config *config.Config
  //TODO: add stats 
}

func NewApiClient(repoDir string) (*ApiClient, error) {
  
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

  repo_attrs, err := rest_client.NewRESTClient(server_addr, c.RepoUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  base.Debug("[NewApiClient] Server Address:", server_addr)

  flow_ep, err := rest_client.NewRESTClient(server_addr, c.FlowAttrsUriPath, http.DefaultClient)


  // TODO: find a use or remove this
  repo_fs := fs.NewFS(repoDir, c.Concurrency)

  code_sync := code_sync.NewClient(server_string, repo_fs)

  commit_map, err := rest_client.NewRESTClient(server_addr, c.CommitMapUriPath, http.DefaultClient)
  contents, err := rest_client.NewRESTClient(server_addr, c.ObjectUriPath, http.DefaultClient)
  vfs, err := rest_client.NewRESTClient(server_addr, c.VfsUriPath, http.DefaultClient)
  datasetInfo, err := rest_client.NewRESTClient(server_addr, c.DataSetUriPath, http.DefaultClient)

  return &ApiClient {
    repoInfo: repo_attrs,
    codeSync: code_sync,
    repoFS: repo_fs,
    flowIo: flow_ep,
    CommitMap: commit_map,
    ContentIo: contents,
    dataSetInfo: datasetInfo,
    vfs: vfs,
    ServerAddr: server_addr,
    config: c,
  }, nil

}

func (c *ApiClient) InitRepo(repoName string) error {

  repo_req := c.repoInfo.Post()
  repo_req.Param("repoName", repoName)
  resp := repo_req.Do()
  _, err := resp.Raw()

  //fmt.Println("resp: ", string(b))

  if err != nil {
    return fmt.Errorf("Failed while initializing repo: %s", err)
  }

  base.Log("[InitRepo] Repo created: ", repoName)
  return nil

}

func (c *ApiClient) CloneRepo(repoName, branchName string) (commitId string, err error) {
  var commit_id string
  repo_req := c.repoInfo.Verb("GET")
  repo_req.Param("repoName", repoName)
  repo_req.Param("branchName", branchName)
  resp := repo_req.Do()
  
  body, err := resp.Raw()

  if err != nil {
    return commit_id, fmt.Errorf("Error occured in get_repo: %s", err)
  }

  getRepoResponse := schema.GetRepoResponse{}
  err = json.Unmarshal(body, &getRepoResponse)

  if err != nil {
    return commit_id, fmt.Errorf("Failed to decode get_repo response %s", err)
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
    return commit_id, err
  }
  fmt.Println("clone size:", s)
  commitId = getRepoResponse.CommitId
  return
}

// push code updates and then call run 
func (c *ApiClient) RunTask(repoName string, branchName string, commitId string, cmdStr string) (openCommitId string, task_status string, fnError error) {
  var err error
  var status_str string 
  var commit_id string
  commit_id, err = c.PushRepo(repoName, branchName, commitId) 
  
  if err != nil {
    base.Log("[ApiClient.RunTask] Failed to push code updates to server: ", err)
    return commit_id, status_str, err
  }

  api_req := c.flowIo.Verb("POST") 

  task_req := flow_pkg.NewFlowLaunchRequest {
    Repo: ws.Repo {
        Name: repoName,
      },
    Branch: ws.Branch {
      Name: branchName,
    },
    Commit: ws.Commit{
      Id: commitId,
      },
    CmdString: cmdStr,
  }

  task_req_json, _ := json.Marshal(&task_req) 
  _ = api_req.SetBodyReader(ioutil.NopCloser(bytes.NewReader(task_req_json)))

  http_resp := api_req.Do()
  api_resp, err := http_resp.Raw()  

  task_resp :=  flow_pkg.NewFlowLaunchResponse{}
  err = json.Unmarshal(api_resp, &task_resp)
  status_str = task_resp.TaskStatusStr
  fmt.Println("response: ", status_str, task_resp.Flow)
  base.Log("[RunTask] Flow Id: ", task_resp.Flow.Id)

  return commit_id, status_str, err
}
 

func (c *ApiClient) pushCode(repoName, branchName, commitId, fullFilePath string) (int64, error) {
  // get relative path to file 

  var upld_size int64

  base.Debug("[RepoFs.PushObject] Pushing File: ", fullFilePath)
  repo_path:= c.repoFS.GetWorkingDir()
  rel_path, err := filepath_pkg.Rel(repo_path, fullFilePath)

  file_io, err := os.Open(fullFilePath)
  if err != nil {
    return upld_size, err
  } 

  buf := utils.GetBuffer()
  defer utils.PutBuffer(buf)

  //TODO: add this inside loop to send or add multi part writer 
  w, err := c.PutObjectWriter(repoName, branchName, commitId, rel_path)
  defer w.Close()

  for {
        
      read_len, err := file_io.Read(buf)
      if read_len == 0 && err != nil {
        if err == io.EOF {
          return upld_size, nil
        }
        return upld_size, err
      }
      
      wrt_len, err := w.Write(buf[:read_len])

      upld_size := upld_size + int64(wrt_len)

      if err != nil {
        return upld_size, err
      }

    }

  return upld_size, nil

}

func (c *ApiClient) pushCodeUpdates(repoName, branchName, commitId string) error {
  var upload_size uint64
  var file_len int64
  var eg errgroup.Group
  
  commit_map, err := c.fetchCommitMap(repoName, commitId) 
  if err != nil {
    base.Log("[ApiClient.pushCodeUpdates] Error :", err)
    return err
  }

  var repo_path string = c.repoFS.GetWorkingDir()
  
  // TODO : check / 
  if err:= filepath_pkg.Walk(repo_path, func(current_path string, file_osinfo os.FileInfo, err error) error {

    if (file_osinfo.Mode() & file_osinfo.Mode() > os.ModeNamedPipe) {
      base.Log("[ApiClient.pushCodeUpdates] Found Named pipe. Skipping.. ", current_path)
      return nil
    }

    if (file_osinfo.Mode() & file_osinfo.Mode() > os.ModeSymlink) {
      base.Log("[ApiClient.pushCodeUpdates] Found Named Symlink. Skipping.. ", current_path)
      return nil
    }

    file_commit_info, ok := commit_map.Entries[current_path]
    _ = file_commit_info

    if !ok {

      // file doesnt exist. Push
      eg.Go(func() (upldError error){  
        file_len, err = c.pushCode(repoName, branchName, commitId, current_path) 
        if err != nil {
          return err
        }

        atomic.AddUint64(&upload_size, uint64(file_len))
        return nil   
      })    
       
    }

    //if (file_commit_info.size == file_osinfo.Size()) {
      // file not changed. Skip the file 
    //  return nil
    //}

    // update code 
    eg.Go(func() (upldError error){  
      file_len, err = c.pushCode(repoName, branchName, commitId, current_path)
      if err != nil {
        return err
      }
      atomic.AddUint64(&upload_size, uint64(file_len))
      return nil
    })
    return nil 

  }); err != nil {
    base.Log("[ApiClient.pushCodeUpdates] Walkthrough completed with an error: ", err)
  }

  fnError := eg.Wait()
  // TODO: upload size 
  base.Log("[ApiClient.pushCodeUpdates] Upload size: ", upload_size)
  return fnError

}


// Need branch here?
// should compare file sizes and decide to move or not ?
// Push always pushes to the head of commit graph on branch
// 
func (c *ApiClient) PushRepo(repoName, branchName, commitId string) (string, error) {
  var open_commit_id string 
  var err error  

  commit_cli, err := rest_client.NewRESTClient(c.ServerAddr, c.config.CommitUriPath, http.DefaultClient)

  commit_req := commit_cli.Get()
  commit_req.Param("repoName", repoName)
  commit_req.Param("branchName", branchName)
  commit_req.Param("commitId", commitId)

  base.Debug("[ApiClient.PushRepo] Commit Url: ", commit_req.URL())
  commit_resp := commit_req.Do()
  commit_data, err := commit_resp.Raw()

  // commit ID validated
  if err != nil {
    base.Log("[ApiClient.PushRepo] Failed to retrieve an open repo commit: ", err)
    return open_commit_id, err
  }

  commit_attrs := &ws.CommitAttrs{}
  err = json.Unmarshal(commit_data, &commit_attrs)

  open_commit_id = commit_attrs.Commit.Id

  // got an open commit, now push code 
  err =  c.pushCodeUpdates(repoName, branchName, open_commit_id)
  if err != nil {
    return "", err
  }

  return open_commit_id, nil
}
 

func (c *ApiClient) PutObjectWriter(repoName string, branchName string, commitId string, fpath string) (io.WriteCloser, error) {
  r := c.vfs.VerbSp("PUT", "put_file")
  r.Param("repoName", repoName)
  r.Param("branchName", branchName)

  r.Param("commitId", commitId)
  r.Param("path", fpath)
  
  hw := &httpWriter {
    r: r,
  }  

  return hw, nil
}




