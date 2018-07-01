package rest 

import (
  "fmt"
  "net/http"
  "hyperview.in/server/base"
  "golang.org/x/net/context"
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

  srcFile, srcHeader, err := h.rq.FormFile("file")

  if err != nil {
    errString := fmt.Sprintf("File Error: %s", err.Error())
    return base.HTTPErrorf(http.StatusInternalServerError, errString)
  }

  if srcFile == nil {
    return base.HTTPErrorf(http.StatusBadRequest, err.Error())
  }

  defer srcFile.Close()

  file_hash, byteswriten, err := h.server.objectAPI.PutObject(context.Background(), srcFile, false)
  
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
  }

  h.writeJSON(response)
  return nil
}


func (h *handler) handleReadObject() error {

  if (h.rq.Method != "GET") {
    return base.HTTPErrorf(http.StatusMethodNotAllowed, "Invalid method %s", h.rq.Method)
  }
  file_hash := h.getQuery("file_hash")

  datum, bytes, err := h.server.objectAPI.GetObject(file_hash)

  if err != nil {
    base.Log("Failed to fetch file object: ", file_hash, err)
    return base.HTTPErrorf(http.StatusBadRequest, "Error occurred while fetching the given object.")
  }

  response := map[string]interface{}{
    "file_hash": file_hash,
    "data": datum,
    "bytes": bytes,
  }

  h.writeJSON(response)

  return nil

}



