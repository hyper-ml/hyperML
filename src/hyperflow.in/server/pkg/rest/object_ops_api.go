package rest

import ( 
  "net/http"  

  db_pkg "hyperflow.in/server/pkg/db" 
  "hyperflow.in/server/pkg/base"
 
)

// create a new hash for first request from vfs server 
// updae commit with hash 


// if object hash is available then just check if commit is open and get on with write 
// if obj has is null then call put file writer which may return existing or new obj hash 
// need really bare minimum code here. Manage some how so other network methods can be 
// added in future. Let API Server manage most of the workload ?? make calls depend on values of 
// obj has and send body reader 

// send obj has to client through file info so next time client 
// can send hash and improve write time 



func (h *Handler) handlePutObject() error {

  if (h.rq.Method != "PUT") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  var response map[string]interface{}

  repo_name := h.getQuery("repoName")
  commit_id := h.getQuery("commitId")
  file_path := h.getQuery("path")

  // TODO: use commit map or raise error if commt is not open 

  // TODO: create file info and object cache 

  file_attrs, err := h.server.wsAPI.GetFileAttrs(repo_name, commit_id, file_path)
  
  if err != nil {
    if !db_pkg.IsErrRecNotFound(err) {
      base.Log("failed to retrieve file info", err)
      return err
    }
     
    base.Debug("File not found in commit map. Creating a new entry", repo_name, commit_id, file_path)  
  }
   
  object_hash := file_attrs.Object.Path

  if h.rq.Body != nil {
    obj_path, chksum, n, err := h.server.objAPI.SaveObject(object_hash, "objects", h.rq.Body, false) 
    if err != nil {
      base.Log("Failed to update object on server", file_path)
      return err
    } 
    response = map[string]interface{}{
      "obj_path": obj_path,
      "size": n,
      "checksum": chksum, 
    }
     h.writeJSON(response)
    return nil
  }

  response = map[string]interface{}{
      "obj_path": "",
      "size": 0,
  }

  h.writeJSON(response)
  return nil
}