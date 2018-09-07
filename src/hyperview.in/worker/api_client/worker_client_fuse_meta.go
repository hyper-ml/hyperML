package api_client

import (
  "io"
  "fmt"
  "strconv"
  "bytes"
  "io/ioutil"
  "encoding/json"

  "hyperview.in/worker/rest_client"
  ws "hyperview.in/server/core/workspace"
  "hyperview.in/server/base" 
  local_schema "hyperview.in/worker/schema"
)
 
func (wc *WorkerClient) ListDir(reponame string, commitId string, path string) (map[string]*ws.FileAttrs, error) {
  //var file_list map[string]*ws.FileAttrs

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
  finfo_map := ws.FileAttrsMap{}
  err = json.Unmarshal(body, &finfo_map)
  //base.Debug("Received list dir body:", finfo_map)
  
  //return &commit_attrs, nil
  return finfo_map.Entries, nil
}




func (wc *WorkerClient) LookupFile(reponame string, commitId string, p string) (*ws.FileAttrs, error) {
  //var file_list map[string]*ws.FileAttrs

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
  
  f_info := ws.FileAttrs{}
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

  if pfr.FileAttrs.Object != nil {   
    base.Debug("Received File Info. Caching object hash with writer for future use: ", pfr.FileAttrs.Object.Hash)
    h.setHash(pfr.FileAttrs.Object.Hash) 
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