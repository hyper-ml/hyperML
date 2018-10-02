package rest

import (
  "net/http"
  "flag"
  "hyperview.in/server/base"
)
 

var DefaultInterface = ":8888"
var DefaultMaxIncomingConnections = 0
var ServerReadTimeout = 200
var ServerWriteTimeout = 200
 

type RunConfig struct {
  *config.Config
  BaseDir string
}

var config *RunConfig

func (config *RunConfig) Serve(addr string, Handler http.Handler) {
  err := ListenAndServeHTTP(addr, DefaultMaxIncomingConnections, ServerReadTimeout, ServerWriteTimeout, Handler)
  
  if err != nil {
    base.Log("Failed to start HTTP Server on %s: %v", addr, err)
  }
}
  
func StartServer(addr string) {
  c, err := config.GetConfig()
  if err != nil {
    panic(err)
  }
  
  c.Interface = addr
  runServer(c)  
}


func runServer(config *RunConfig) {
  sc := NewServerContext(config)
  base.Info("Starting server on %s ...", *config.Interface)
  config.Serve(*config.Interface, CreatePublicHandler(sc))
}

  