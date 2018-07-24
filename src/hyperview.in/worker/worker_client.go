package worker

// what: Client to access apis and process

import (
  "net/http" 
  "fmt"
  "net/url" 
  "encoding/json"
  "strconv"
  "io"
  "bytes" 
  "io/ioutil"

  "hyperview.in/server/base" 

  "hyperview.in/worker/config" 
  "hyperview.in/worker/rest_client" 
  ws "hyperview.in/server/core/workspace"

  local_schema "hyperview.in/worker/schema"
)


type FsConfig struct {
  repoDir string
}

type ServerConfig struct {
  server_http string
  base_uri string
  repo_path string
}

type WorkerClient struct {

  // Rest Client to fetch info from server   
  RepoInfo rest_client.Interface

  // Rest client for commit info
  CommitInfo rest_client.Interface

  // Rest client for branch info
  BranchInfo rest_client.Interface

  // File Info client
  FileInfo rest_client.Interface

  // object client
  ContentIo rest_client.Interface

  // Virtual FS Client 
  vfs rest_client.Interface

  //TODO: add stats 
}

func NewWorkerClient() (*WorkerClient, error) {
  
  c, err := config.ReadFromFile()

  if err != nil {
    fmt.Println("Failed to read config file")
    c = config.Default()
  } 
  
  server_addr, err := url.Parse(c.ServerAddr) 
  if err != nil {
    return nil, err
  }

  repo_info, err := rest_client.NewRESTClient(server_addr, c.RepoInfoUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  commit_info, err := rest_client.NewRESTClient(server_addr, c.CommitInfoUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  branch_info, err := rest_client.NewRESTClient(server_addr, c.BranchInfoUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  file_info, err := rest_client.NewRESTClient(server_addr, c.FileInfoUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  vfs, err := rest_client.NewRESTClient(server_addr, c.VfsUriPath, http.DefaultClient)
  if err != nil {
    return nil, err
  }

  contents, err := rest_client.NewRESTClient(server_addr, c.ObjectUriPath, http.DefaultClient)

  return &WorkerClient {
    RepoInfo: repo_info,
    BranchInfo: branch_info,
    CommitInfo: commit_info,
    FileInfo: file_info,
    vfs: vfs,
    ContentIo: contents,
  }, nil

} 


func (wc *WorkerClient) FetchBranchInfo(repoName, branchName string) (*ws.BranchInfo, error) {
  var branch_info ws.BranchInfo

  brq := wc.BranchInfo.Get() 
  brq.Param("repoName", repoName)
  brq.Param("branchName", branchName)

  resp := brq.Do()
  body, err := resp.Raw()
  if err != nil {
    return nil, err
  } 

  branch_info =  ws.BranchInfo{}
  err = json.Unmarshal(body, &branch_info) 

  return &branch_info, nil
}

func (wc *WorkerClient) FetchCommitInfo(repoName, commitId string) (*ws.CommitInfo, error) {
  var commit_info ws.CommitInfo

  crq := wc.CommitInfo.Get() 
  crq.Param("repoName", repoName)
  crq.Param("commitId", commitId)
  resp := crq.Do()
  body, err := resp.Raw()
  if err != nil {
    return nil, err
  }
  err = json.Unmarshal(body, &commit_info)
  return &commit_info, nil
}
 
func (wc *WorkerClient) FetchFileInfo(repoName string, commitId string, fpath string) (*ws.FileInfo, error){
  var file_info *ws.FileInfo 
  crq := wc.FileInfo.Verb("GET") 
  crq.Param("repoName", repoName)
  crq.Param("commitId", commitId)
  crq.Param("path", fpath)

  resp := crq.Do()
  body, err := resp.Raw()
  if err != nil {
    return nil, err
  }
  
  err = json.Unmarshal(body, &file_info)
  return file_info, nil

}

func (wc *WorkerClient) ListDir(reponame string, commitId string, path string) (map[string]*ws.FileInfo, error) {
  //var file_list map[string]*ws.FileInfo

  crq := wc.vfs.VerbSp("GET", "list_dir") 
  crq.Param("repoName", reponame)
  crq.Param("commitId", commitId)
  crq.Param("path", path)

  resp := crq.Do()
  body, err := resp.Raw()
  if err != nil {
    return nil, err
  }
  //base.Debug("Received list dir body:", string(body))
  finfo_map := ws.FileInfoMap{}
  err = json.Unmarshal(body, &finfo_map)
  //base.Debug("Received list dir body:", finfo_map)
  
  //return &commit_info, nil
  return finfo_map.Entries, nil
}



func (wc *WorkerClient) LookupFile(reponame string, commitId string, p string) (*ws.FileInfo, error) {
  //var file_list map[string]*ws.FileInfo

  crq := wc.vfs.VerbSp("GET", "lookup") 
  crq.Param("repoName", reponame)
  crq.Param("commitId", commitId)
  crq.Param("path", p)

  resp := crq.Do()
  body, err := resp.Raw()

  if err != nil {
    base.Log("[WorkerClient.LookupFile] Error looking up file:", reponame, commitId, p, err)
    return nil, err
  }

  //base.Debug("Received list dir body:", string(body))
  
  f_info := ws.FileInfo{}
  err = json.Unmarshal(body, &f_info)
    
  return &f_info, nil
}

func (wc *WorkerClient) GetFileObject(repoName string, commitId string, fpath string, offset int64, size int64, writebuf io.Writer) error {
  base.Debug("GetFileObject(): repoName, commitId, fpath, offset, size: ")
  base.Debug(repoName, commitId, fpath, offset, size)
  var err error
  var n int

  r := wc.vfs.VerbSp("GET", "get_file")
  r.Param("repoName", repoName)
  r.Param("commitId", commitId)
  r.Param("filePath", fpath)
  r.Param("offset", strconv.Itoa(int(offset)))
  r.Param("size", strconv.Itoa(int(size)))

  obj_reader, err:= r.ReadResponse() 
  
  if err != nil {
    base.Log("Failed to get handle on HTTP Request during GetFileObject:", commitId, fpath)
    return err
  }

  w, err := ioutil.ReadAll(obj_reader)
  base.Log("w:", string(w))
  
  defer obj_reader.Close()

  buffer := make([]byte, 1024) //2 * 1024 * 1024

  for {
    n, err = obj_reader.Read(buffer)

    if err != nil {
      //TODO: handle ErrUnexpectedEOF
      if err == io.EOF && n == 0 {
        return nil
      }
      if err != io.EOF {
        return err
      }
    }   

    //TODO: return bw from this method
    
    _, err := writebuf.Write(buffer[:n]) 
    if err != nil {
      base.Log("Failed to pull object file into local writer", commitId, fpath)
      return err
    }
  }

  return nil

}

// add mutex to synchronize writes   
type httpWriter struct {
  r *rest_client.Request
  object_hash string
}

func (h *httpWriter) setHash(hash string) {
  h.object_hash = hash 
}

// TODO: Write at will be better. But figure it out later
//
func (h *httpWriter) Write(p []byte) (n int, err error) {
 
  h.r.Param("size", strconv.Itoa(len(p)))
  h.r.Param("hash", h.object_hash)

  _ = h.r.SetBodyReader(ioutil.NopCloser(bytes.NewReader(p)))

  resp := h.r.Do()

  if resp.Error()!= nil {
    base.Log("Encountered an error while writing object to server: ", h.object_hash, err)
    _= h.r.PrintParams()
    return 0, err
  } 

  pfr := local_schema.PutFileResponse{}
  err = json.Unmarshal(resp.Body(), &pfr)

  if err != nil {
    base.Log("Invalid response from server for PutFileResponse:", err)
    return 0, err
  }

  if pfr.Error != "" {
    return 0, fmt.Errorf(pfr.Error)
  }

  if pfr.FileInfo.Object != nil {   
    base.Debug("Received File Info. Caching object hash with writer for future use: ", pfr.FileInfo.Object.Hash)
    h.setHash(pfr.FileInfo.Object.Hash) 
  } 

  return int(pfr.Written), nil
}

func (h *httpWriter) Close() error {
  // Close body here?  
  return nil
}
 

func (wc *WorkerClient) PutObjectWriter(repoName string, commitId string, fpath string) (io.WriteCloser, error) {
  r := wc.vfs.VerbSp("PUT", "put_file")
  r.Param("repoName", repoName)
  r.Param("commitId", commitId)
  r.Param("path", fpath)
  
  hw := &httpWriter {
    r: r,
  }  

  return hw, nil
}
 