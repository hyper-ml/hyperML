package rest 

import (
  "fmt"
  "time"
  "io"
  //"encoding/json"
  "strconv"
  "net/http"
  "hyperview.in/worker/utils/structs"
  "hyperview.in/server/base"
 
)


func (h *handler) handleRoot() error {
	response := map[string]interface{}{
    "hyperview": "version 0.1",
  }

  h.writeJSON(response)
  return nil
}

/* TODO: Add target dir 
*/
func (h *handler) handleCreateObject() error {
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

  file_hash, chksum, byteswriten, err := h.server.objectAPI.CreateObject(parentDir, srcFile, false)
  
  if err != nil {
    base.Log("Failed to write file:\n", err) 
    return base.HTTPErrorf(http.StatusInternalServerError, "Failed to write file on to server")
  } 

  if !h.server.objectAPI.CheckObject(file_hash) {
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

func (h *handler) handleGetObject() error {
  var err error
  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }

  repo_name := h.getQuery("repoName")
  commit_id := h.getQuery("commitId")
  file_path := h.getQuery("filePath")
  offset_str := h.getQuery("offset")
  size_str := h.getQuery("size")

  base.Debug("handleGetObject(): ", repo_name, commit_id, file_path, offset_str, size_str)

  var offset int64
  var size int64

  // convert strings to int64 
  if offset, err = strconv.ParseInt(offset_str, 10, 64); err != nil {
    base.Log("Failed to convert offset string to int64", offset_str)
    return err 
  }
  if size, err = strconv.ParseInt(size_str, 10, 64); err!= nil {
    base.Log("Failed to convert offset string to int64", offset_str)
    return err
  }

  file_info, err := h.server.workspaceApi.GetFileInfo(repo_name, commit_id, file_path)

  if err != nil {
    fmt.Println("failed to retrieve file info", err)
    return err
  }
  
  object_hash := file_info.Object.Path
  base.Debug("handleGetObject(): object_hash")
  rs, err:= h.server.objectAPI.ReadSeeker(object_hash, offset, size)

  if err != nil {
    if  err != io.EOF {
      base.Log("Failed to fetch file object: ", object_hash, err)
      return base.HTTPErrorf(http.StatusBadRequest, "Error occurred while fetching the given object.")
    }
    base.Log("handleGetObject(): Reached EOF. Nothing to read. ", object_hash)
    return nil
  }

  h.setHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", file_path))
  
  http.ServeContent(h.response, h.rq, file_path, time.Time{}, rs)
  
  return nil
}

 

// What:  Gets Repo, Master Branch and Commit Info
//

func (h *handler) handleGetRepo() error {
  var response map[string]interface{}
  repoName := h.getQuery("repoName")
  
  //TODO: handle error
  repo_info, branch_info, commit_info, fmap, _ := h.server.workspaceApi.DownloadRepo(repoName)

  // Repo Response:
  //    Repo
  //    Branch
  //    CommitId <- Head
  //    FileMap

  if commit_info != nil {
    response = map[string]interface{}{
      "repo": repo_info.Repo, 
      "branch": branch_info.Branch,
      "commit_id": commit_info.Commit.Id,
      "file_map": fmap, 
    }
  }

  h.writeJSON(response)

  return nil
} 

func (h *handler) handleGetRepoInfo() error {
  var response map[string]interface{}
  repoName := h.getQuery("repoName")

  if repoName == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
  }
  
  //TODO: handle error
  repo_info, err := h.server.workspaceApi.GetRepoInfo(repoName)


  fmt.Println("repo_info:", repo_info)
  
  if err == nil {
    response = map[string]interface{}{
      "repo_info": repo_info, 
    }
  } else {
    return err
  }

  fmt.Println("response on handleGetRepoInfo: ", response)
  h.writeJSON(response)

  return nil
} 


func (h *handler) handleGetBranchInfo() error {
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
  branch_info, err := h.server.workspaceApi.GetBranchInfo(repoName, branchName)

  fmt.Println("branch_info:", branch_info)
  
  if err == nil {
    response = structs.Map(branch_info) 
  } else {
    return err
  }

  fmt.Println("response on handleGetRepoInfo: ", response)
  h.writeJSON(response)

  return nil
} 

func (h *handler) handleGetCommitInfo() error {
  var response map[string]interface{}
  repoName := h.getQuery("repoName")
  commitId := h.getQuery("commitId")

  if repoName == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - repoName")
  }

  if commitId == "" { 
    return base.HTTPErrorf(http.StatusInternalServerError, "Invalid repo param - commitId")
  }
    
  //TODO: handle error
  commit_info, err := h.server.workspaceApi.GetCommitInfo(repoName, commitId)
  
  if err == nil {
    response = structs.Map(commit_info) 
  } else {
    return err
  }

  fmt.Println("response on handleGetRepoInfo: ", response)
  h.writeJSON(response)

  return nil
} 



func (h *handler) handleGetFileInfo() error {
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
    
  file_info, err := h.server.workspaceApi.GetFileInfo(repoName, commitId, fpath)
  
  if err == nil {
    //TODO: handle nil file info
    response = structs.Map(file_info) 
  } else {
    return err
  }

  h.writeJSON(response)
  return nil
} 