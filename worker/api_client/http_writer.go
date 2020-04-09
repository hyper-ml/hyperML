package api_client

import(
  "fmt"
  "strconv"  
  "io/ioutil"
  "bytes"

  "encoding/json"
  "github.com/hyper-ml/hyperml/worker/rest_client"

  "github.com/hyper-ml/hyperml/server/pkg/base"

  local_schema "github.com/hyper-ml/hyperml/worker/schema"

  ws "github.com/hyper-ml/hyperml/server/pkg/workspace"
)


// add mutex to synchronize writes   
type httpFileWriter struct {
  r *rest_client.Request
  object_hash string
}

func (h *httpFileWriter) setHash(hash string) {
  h.object_hash = hash 
}

func (h *httpFileWriter) Write(p []byte) (n int, err error) {
  p_len := len(p)
  
  h.r.Param("size", strconv.Itoa(p_len))
  h.r.Param("hash", h.object_hash)

  _ = h.r.SetBodyReader(ioutil.NopCloser(bytes.NewReader(p)))

  resp := h.r.Do()

  if resp.Error()!= nil {
    base.Error("Encountered an error while writing object to server: ", h.object_hash, err)
    _= h.r.PrintParams()
    return 0, err
  } 

  pfr := local_schema.PutFileResponse{}
  err = json.Unmarshal(resp.Body(), &pfr)

  if err != nil {
    base.Error("[httpFileWriter.write] Write Object Error:", h.object_hash, err)
    return 0, err
  }

  if pfr.Error != "" {
    return 0, fmt.Errorf(pfr.Error)
  }

  if pfr.FileAttrs.Object != nil {   
    h.setHash(pfr.FileAttrs.Object.Hash) 
  } 
  
  //todo: keep track of object hash in return values
  return p_len, nil

}

func (h *httpFileWriter) Close() error {
  // Close body here?  
  return nil
}



// add mutex to synchronize writes   
type httpObjectWriter struct {
  r *rest_client.Request 
}
  
func (h *httpObjectWriter) Write(p []byte) (n int, err error) {
  
  h.r.Param("size", strconv.Itoa(len(p)))

  _ = h.r.SetBodyReader(ioutil.NopCloser(bytes.NewReader(p)))

  resp := h.r.Do()
  body, err := resp.Raw()

  if err != nil {
    base.Warn("[httpObjectWriter.Write] object writing error: ", err)
    return 0, err
  }

  object:= ws.Object{} 
  err = json.Unmarshal(body, &object)

  return object.Size, nil
}

func (h *httpObjectWriter) Close() error {
  return nil
}
