package rest


import (
	"sync"
  "net/http"
  "time"
  "strconv"
  "hyperview.in/server/core/storage"
  "hyperview.in/server/base"

)

type ServerContext struct {
	config *ServerConfig
	lock sync.RWMutex
	statsTicker *time.Ticker
	HTTPClient  *http.Client
  objectAPI storage.StorageServer
}

// TODO: add error logging and response
// accessed by HTTP handlers so needs to be thread safe 
func NewServerContext(config *ServerConfig) *ServerContext{

  dir := "test_dir"
  cacheStr := base.GetEnv("OBJECT_CACHE_BYTES")
  var cacheBytes int64
  var err error

  if (cacheStr != "") {
    cacheBytes, err = strconv.ParseInt(cacheStr, 10, 64)
    if err == nil {
      base.Log("invalid cache bytes.")
    }
  }

  //TODO: add config variable for storage option
  oapi, _ := storage.NewStorageServer(dir, cacheBytes, "GCS")  

  sc := &ServerContext{
    config: config,
    HTTPClient: http.DefaultClient,
    objectAPI: oapi,
  }
  return sc
}