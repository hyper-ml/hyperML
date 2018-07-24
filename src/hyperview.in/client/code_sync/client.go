package code_sync

import ( 
  "fmt"
  "net/url" 
  "golang.org/x/sync/errgroup"
  "hyperview.in/client/fs"
)

type Client struct {
  serverAddr string
  objectsUrlPath string
  localFS fs.FS
}

func NewClient(serverAddr string, localFS fs.FS) *Client{
  return &Client {
    serverAddr: serverAddr,
    objectsUrlPath: "object",
    localFS: localFS,
  }
}

func (c *Client) PullObject(repoName string, commitId string, fileName string) (int64, error) {
  var file_size int64

  //TODO: check the spaces escape chars are handled
  url_string := c.serverAddr 
  url, err := url.Parse(url_string) 
  
  //fmt.Println("url", url)
  file_op := NewFileOp(nil, url, c.objectsUrlPath, c.localFS)
  file_op = file_op.RemoteURLParams("commitId", commitId)
  file_op = file_op.RemoteURLParams("repoName", repoName)
  file_op = file_op.RemoteURLParams("filePath", fileName)
  
  //call driver here
  file_size, err = file_op.GetFileByUrl(fileName) 
  fmt.Println("file_size", file_size)

  if err != nil {
    return 0, err
  }

  return file_size, nil
}

func (c *Client) PullRemoteRepo(pullRepoRequest *PullRepoRequest, parallel int) (int64, error){
  op_limiter := NewOpLimiter(parallel)
  var total_bytes int64
  var eg errgroup.Group

  for key, value := range pullRepoRequest.FileMap { 
    fmt.Println("Key:", key, "Value:", value)
    //bytes_written, err:= d.GetFile(key)
    op_limiter.Ask()

    eg.Go(func() (retError error){
      defer op_limiter.Release()
      download_size, err := c.PullObject(pullRepoRequest.RepoName, pullRepoRequest.CommitId, key)
      
      //TODO: need to improve this with mutex 
      total_bytes = total_bytes + download_size
      return err
      })
  }
  
  fmt.Println("total_bytes:", total_bytes)
  return total_bytes, eg.Wait()
}

