package client

import(
  "fmt"
  "strconv"  
  "io/ioutil"
  "bytes"

  "encoding/json"
  "hyperview.in/client/rest_client"

  "hyperview.in/server/base"

  local_schema "hyperview.in/client/schema"

)


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
