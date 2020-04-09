package api_client

import (
  "io" 
  "strconv" 
  "io/ioutil"
  "encoding/json"
 
  ws "github.com/hyper-ml/hyperml/server/pkg/workspace"
  "github.com/hyper-ml/hyperml/server/pkg/base"  
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
    base.Error("[WorkerClient.LookupFile] Error looking up file:", reponame, commitId, p, err)
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
    base.Error("Failed to get handle on HTTP Request during GetFileObject:", commitId, fpath)
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
      base.Error("Failed to pull object file into local writer", commitId, fpath)
      return err
    }
  }

  return nil

}

func (wc *WorkerClient) PutFileWriter(repoName string, commitId string, fpath string) (io.WriteCloser, error) {
  r := wc.vfs.VerbSp("PUT", "put_file")
  r.Param("repoName", repoName)
  r.Param("commitId", commitId)
  r.Param("path", fpath)
  
  hw := &httpFileWriter {
    r: r,
  }  

  return hw, nil
}