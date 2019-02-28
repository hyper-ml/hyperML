package rest 

import (
  "fmt"
  "time"
  "io"
  "io/ioutil"
  "strconv"
  "net/http"
  "encoding/json"
  "hyperflow.in/server/pkg/base/structs"
  "hyperflow.in/server/pkg/base"
  "hyperflow.in/server/pkg/auth"
 
)

func raiseError(error_msg string) error{
  return base.HTTPErrorf(http.StatusInternalServerError, error_msg)
}


func ValidateMethod(request, expected string) error {
  if expected != request {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", request)
  }
  return nil
}

func (h *Handler) handleRoot() error {
	response := map[string]interface{}{
    "hyperflow": "version 0.1",
  }

  h.writeJSON(response)
  return nil
}


func (h *Handler) handleBasicAuth() error {
  base.Info("[Handler.handleBasicAuth] started at :", time.Now())
  if err := ValidateMethod(h.rq.Method, "POST"); err != nil{
    return err
  }

  auth_data, err := ioutil.ReadAll(h.rq.Body)
  if err != nil {    
    base.Error("[Handler.handleBasicAuth] Invalid data input: ", auth_data, err)
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid data input")
  }

  auth_request := auth.AuthRequest{}
  if err := json.Unmarshal(auth_data, &auth_request); err != nil {
    base.Error("[Handler.handleBasicAuth] Invalid data input: ", auth_data, err)
    return base.HTTPErrorf(http.StatusBadRequest, "Invalid json input")
  }

  jwt, user_attrs, err := h.server.authAPI.CreateSession(auth_request.UserName, auth_request.Password)
  if err != nil {
    base.Error("[Handler.handleBasicAuth]Failed to create session: ", err)
    return base.HTTPErrorf(http.StatusBadRequest, ": {" + err.Error() + "}")
  }

  auth_response := &auth.AuthResponse {
    Jwt: jwt,
    UserAttrs: user_attrs,
  }

  response :=  structs.Map(auth_response)
  h.writeJSON(response)

  return nil  
}

/* TODO: Add target dir 
*/
/*! Unused - to be removed
func (h *Handler) handleCreateObject() error {
  var err error 

  if (h.rq.Method != "POST") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  parentDir := "objects"
  srcFile, srcHeader, err := h.rq.FormFile("file")

  if err != nil {
    errString := fmt.Sprintf("File Error: %s", err.Error())
    return base.HTTPErrorf(http.StatusInternalServerError, errString)
  }

  if srcFile == nil {
    return base.HTTPErrorf(http.StatusBadRequest, err.Error())
  }

  defer srcFile.Close()

  file_hash, chksum, byteswriten, err := h.server.objects.CreateObject(parentDir, srcFile, false)
  
  if err != nil {
    base.Log("Failed to write file:\n", err) 
    return base.HTTPErrorf(http.StatusInternalServerError, "Failed to write file on to server")
  } 

  if !h.server.objects.CheckObject(file_hash) {
    base.Log("Failed to find written file on server. File not found on storage server. %s %s", file_hash, srcHeader.Filename)
    return base.HTTPErrorf(http.StatusInternalServerError, "Failed to find written file on to server")
  }

  response := map[string]interface{}{
    "status": "OBJ_CREATED",
    "input_file_name": srcHeader.Filename,
    "file_hash": file_hash,
    "bytes": byteswriten,
    "chksum": chksum,
  }

  h.writeJSON(response)
  return nil
}
*/ 

func (h *Handler) handleGetObject() error {
  var err error
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  repo_name := h.getQuery("repoName")
  commit_id := h.getQuery("commitId")
  file_path := h.getQuery("filePath")
  offset_str := h.getQuery("offset")
  size_str := h.getQuery("size")

  base.Info("[Handler.handleGetObject] Params: ", repo_name, commit_id, file_path, offset_str, size_str)

  var offset int64
  var size int64

  if offset_str != "" {
    //string to int64 base 10
    offset, err = strconv.ParseInt(offset_str, 10, 64); 
  }

  if size_str != "" {
    //string to int64 base 10
    size, err = strconv.ParseInt(size_str, 10, 64); 
  }
 
  file_attrs, err := h.server.wsAPI.GetFileAttrs(repo_name, commit_id, file_path)

  if err != nil {
    base.Error("[Handler.handleGetObject] Failed to retrieve file info: ", repo_name, commit_id, file_path, err)
    return base.HTTPErrorf(http.StatusBadRequest, err.Error())
  }
  
  object_hash := file_attrs.Object.Path
  base.Info("[Handler.handleGetObject] object_hash: ", object_hash)
  rs, err:= h.server.objAPI.ReadSeeker(object_hash, offset, size)

  if err != nil {
    if  err != io.EOF {
      base.Log("Failed to fetch file object: ", object_hash, err)
      return base.HTTPErrorf(http.StatusBadRequest, "Error occurred while fetching the given object.")
    }
    base.Log("handleGetObject(): Reached EOF. Nothing to read. ", object_hash)
    return nil
  }

  h.setHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", file_path))
  h.setHeader("Content-Type", "application/octet-stream")
  http.ServeContent(h.response, h.rq, file_path, time.Time{}, rs)
  
  return nil
}

 
 

func (h *Handler) handleGetRepoAttrs() error {
  var response map[string]interface{}
  repoName := h.getQuery("repoName")

  if repoName == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
  }
  
  //TODO: handle error
  repo_attrs, err := h.server.wsAPI.GetRepoAttrs(repoName)


   
  if err == nil {
    response = structs.Map(repo_attrs)
  } else {
    return err
  }

  h.writeJSON(response)

  return nil
} 


func (h *Handler) handleGetBranchAttrs() error {
  var response map[string]interface{}
  repoName := h.getQuery("repoName")

  if repoName == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName: %s", repoName )
  }

  branchName := h.getQuery("branchName")

  if branchName == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid branch param - branchName: %s", branchName)
  }

  //TODO: handle error
  branch_attr, err := h.server.wsAPI.GetBranchAttrs(repoName, branchName)
 
  if err == nil {
    response = structs.Map(branch_attr) 
  } else {
    return err
  }

  fmt.Println("response on handleGetRepoAttrs: ", response)
  h.writeJSON(response)

  return nil
} 
 


func (h *Handler) handleGetCommitMap() error {

  var response map[string]interface{}
  repoName := h.getQuery("repoName")
  commitId := h.getQuery("commitId")

  if repoName == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
  }

  if commitId == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid commitId param - commitId")
  }
    
  //TODO: handle error
  commit_map, err := h.server.wsAPI.GetCommitMap(repoName, commitId)
  
  if err == nil {
    response = structs.Map(commit_map) 
  } else {
    return err
  }
  h.writeJSON(response)

  return nil
} 





func (h *Handler) handleGetFileAttrs() error {
  var response map[string]interface{}
  repoName := h.getQuery("repoName")
  commitId := h.getQuery("commitId")
  fpath := h.getQuery("path")

  if repoName == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
  }

  if commitId == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - commitId")
  }
    
  file_attrs, err := h.server.wsAPI.GetFileAttrs(repoName, commitId, fpath)
  
  if err == nil {
    //TODO: handle nil file info
    response = structs.Map(file_attrs) 
  } else {
    return err
  }

  h.writeJSON(response)
  return nil
} 