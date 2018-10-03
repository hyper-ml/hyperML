package rest

import (
  "net/http"
  config_pkg "hyperview.in/server/config"
  "hyperview.in/server/base"
)
 

var DefaultInterface = ":8888"
var DefaultMaxIncomingConnections = 0
var ServerReadTimeout = 200
var ServerWriteTimeout = 200
 
const (
  STORAGE_BASE_DIR = "hyperview"
  LOG_DIR = "logs"
  SEP = "/"
)

type RunConfig struct {
  Current *config_pkg.Config
  StorageBaseDir string
  LogBaseDir string
  StorageDirSeparator string
}

func NewRunConfig(c *config_pkg.Config) *RunConfig {
  return &RunConfig {
    Current: c,
    StorageBaseDir: STORAGE_BASE_DIR,
    LogBaseDir: LOG_DIR,
    StorageDirSeparator: SEP,
  }
}

var rc *RunConfig

func (rc *RunConfig) Serve(Handler http.Handler) {
  err := ListenAndServeHTTP(rc.Current.Interface, DefaultMaxIncomingConnections, ServerReadTimeout, ServerWriteTimeout, Handler)
  
  if err != nil {
    base.Log("Failed to start HTTP Server on %s: %v", rc.Current.Interface, err)
  }
}
  
func StartServer(addr string) {
  c, err := config_pkg.GetConfig()
  if err != nil {
    panic(err)
  }
  
  switch {
  case addr != "":
    c.Interface = addr
  case c.Interface == "":
    c.Interface = DefaultInterface
  } 

  rc := NewRunConfig(c) 
  runServer(rc)  
}

//todo
func raisePanic(err error) {
  panic(err)
}

func runServer(rc *RunConfig) {
  sc, err := NewServerContext(rc)
  
  if err != nil {
    raisePanic(err)
  }

  base.Info("Starting server on %s ...", rc.Current.Interface)
  rc.Serve(CreatePublicHandler(sc))
}

  