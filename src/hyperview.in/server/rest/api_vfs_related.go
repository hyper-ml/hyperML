package rest 

import (  
  "fmt" 
  "net/http"
  "hyperview.in/worker/utils/structs"

  "hyperview.in/server/base"
  ws "hyperview.in/server/core/workspace"


)

func (h *handler) handleListDir() error {
  var response map[string]interface{}
  repo_name := h.getQuery("repoName")
  commit_id := h.getQuery("commitId")
  dir_path := h.getQuery("path")

  base.Debug("handleListDir: %s %s %s", repo_name, commit_id, dir_path)

  if repo_name == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
  }

  if commit_id == "" { 
    base.Debug("No Commit Id passed. This request will use open commit.")
  }

  finfo_map, err := h.server.vfs.ListDir(repo_name, commit_id, dir_path)
  if err != nil {
    base.Log("Failed to retrieve file in dir: %s %s %s", repo_name, commit_id, dir_path)
    return err
  }

  info_map := &ws.FileAttrsMap{Entries: finfo_map}
  
  if err == nil {
    response =  structs.Map(info_map)//map[string]interface{} {
    //  finfo_map, // structs.Map(finfo_map) 
    //}
  } else {
    return err
  }

  fmt.Println("response on handleListDir: ", response)
  h.writeJSON(response)

  return nil
} 



func (h *handler) handleFileLookup() error {

  var response map[string]interface{}
  repo_name := h.getQuery("repoName")
  commit_id := h.getQuery("commitId")
  fpath := h.getQuery("path")

  base.Debug("[handler.handleFileLookup] ", repo_name, commit_id, fpath)

  if repo_name == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
  }

  if commit_id == "" { 
    base.Debug("[handler.handleFileLookup] No Commit Id passed. This request will use open commit.")
  }

  finfo, err := h.server.vfs.Lookup(repo_name, commit_id, fpath)

  if err != nil {
    base.Log("[handler.handleFileLookup] Failed to retrieve file in dir: %s %s %s", repo_name, commit_id, fpath)
    return err
  }
 

  if finfo != nil { 
    base.Debug("[handler.handleFileLookup] finfo is not null ",)

    response =  structs.Map(finfo) 
  } else {
    //TODO: get last element and check if file. return empty file or error

    dummy := ws.NewDirInfo(nil, "", 0)
    response = structs.Map(dummy)
  } 

  base.Debug("[handler.handleFileLookup] Response on handleFileLookup: ", response)
  h.writeJSON(response)

  return nil
} 



func (h *handler) handleVfsPutFile() error {

  if (h.rq.Method != "PUT") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "handleVfsPutFile(): Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  var file_attrs *ws.FileAttrs
  var err error 
  var written int64

  repo_name := h.getQuery("repoName")
  commit_id := h.getQuery("commitId")
  file_path := h.getQuery("path")
  object_hash := h.getQuery("hash")

  if h.rq.Body == nil {
    h.writeJSON(response)
    return nil
  }

  if object_hash == "" {
    base.Debug("[handler.handleVfsPutFile] Calling workspaceApi.PutFile() ", repo_name, commit_id, file_path)
    file_attrs, written, err = h.server.workspaceApi.PutFile(repo_name, commit_id, file_path, h.rq.Body) 
  } else {
    base.Debug("[handler.handleVfsPutFile] calling vfs.  PutFileByHash()", object_hash)
    written, err = h.server.vfs.PutFileByHash(object_hash, h.rq.Body)
  }

  if err != nil {
    base.Error("[handler.handleVfsPutFile] Failed to write file on to server:", object_hash, repo_name, commit_id, file_path, err)

    response = map[string]interface{}{
      "file_attrs" : &ws.FileAttrs{},
      "written" : 0,
      "error": "Failed to write file on to server:",
    }

    h.writeJSON(response)
    return err 
  }

  response = map[string]interface{}{
    "file_attrs" : structs.Map(file_attrs),
    "written" : written,
  }
  h.writeJSON(response)
  return nil
}


func (h *handler) handleVfsGetFile() error {
  return h.handleGetObject()
}









