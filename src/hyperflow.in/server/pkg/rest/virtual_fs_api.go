package rest 
/* -- need to get rid of this as FUSE is unlikely to be used in future versions

import (  
  "fmt" 
  "net/http"
  "hyperflow.in/server/pkg/base"
  ws "hyperflow.in/server/pkg/workspace"
  "hyperflow.in/server/pkg/base/structs"

)

func (h *Handler) handleListDir() error {
  var response map[string]interface{}
  repo_name := h.getQuery("repoName")
  branch_name := h.getQuery("branchName")

  commit_id := h.getQuery("commitId")
  dir_path := h.getQuery("path")

  base.Info("[Handler.handleListDir]: %s %s %s", repo_name, commit_id, dir_path)

  if repo_name == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
  }

  if commit_id == "" { 
    base.Debug("No Commit Id passed. This request will use open commit.")
  }

  finfo_map, err := h.server.vfsAPI.ListDir(repo_name, branch_name, commit_id, dir_path)
  if err != nil {
    base.Log("[handleListDir] Failed to retrieve file in dir: %s %s %s %s", repo_name, branch_name, commit_id, dir_path)
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



func (h *Handler) handleFileLookup() error {

  var response map[string]interface{}
  repo_name := h.getQuery("repoName")
  branch_name := h.getQuery("branchName")
  commit_id := h.getQuery("commitId")
  fpath := h.getQuery("path")

  if repo_name == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
  }

  if branch_name == "" { 
    base.Warn("[Handler.handleFileLookup] No branch_name passed in request")
  }
  
  if commit_id == "" { 
    base.Warn("[Handler.handleFileLookup] No Commit Id passed. This request will use open commit.")
  }

  finfo, err := h.server.vfsAPI.Lookup(repo_name, branch_name, commit_id, fpath)

  if err != nil {
    base.Warn("[Handler.handleFileLookup] Failed to retrieve file in dir: %s %s %s", repo_name, commit_id, fpath)
    return err
  }
 

  if finfo != nil { 
    response =  structs.Map(finfo) 
  } else {
    //TODO: get last element and check if file. return empty file or error

    dummy := ws.NewDirInfo(nil, "", 0)
    response = structs.Map(dummy)
  } 

  base.Debug("[Handler.handleFileLookup] Response on handleFileLookup: ", response)
  h.writeJSON(response)

  return nil
} 



func (h *Handler) handleVfsPutFile() error {

  if (h.rq.Method != "PUT") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "handleVfsPutFile(): Invalid method %s", h.rq.Method)
  }
  
  var response map[string]interface{}
  var file_attrs *ws.FileAttrs
  var err error 
  var written int64

  repo_name := h.getQuery("repoName")
  branch_name := h.getQuery("branchName")
  commit_id := h.getQuery("commitId")
  file_path := h.getQuery("path")
  object_hash := h.getQuery("hash")

  if h.rq.Body == nil {
    h.writeJSON(response)
    return nil
  }

  if object_hash == "" {
    file_attrs, written, err = h.server.wsAPI.PutFile(repo_name, branch_name, commit_id, file_path, h.rq.Body) 
  } else {
    written, err = h.server.vfsAPI.PutFileByHash(object_hash, h.rq.Body)
  }

  if err != nil {
    base.Error("[Handler.handleVfsPutFile] Failed to write file on to server:", object_hash, repo_name, commit_id, file_path, err)

    
    return base.HTTPErrorf(http.StatusBadRequest, err.Error()) 
  }

  response = map[string]interface{}{
    "file_attrs" : structs.Map(file_attrs),
    "written" : written,
  }
  h.writeJSON(response)
  return nil
}


func (h *Handler) handleVfsGetFile() error {
  return h.handleGetObject()
}



*/





