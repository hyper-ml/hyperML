package rest_server

import (
  "net/http"
  "github.com/hyper-ml/hyperml/server/pkg/base"

)

type ServerConfig struct {
  Interface *string
}

func NewServerConfig(addr *string) *ServerConfig{
  return &ServerConfig{
    Interface: addr,
  }
}

func (config *ServerConfig) Serve(addr string, Handler http.Handler) {
   err := ListenAndServeHTTP(addr, DefaultMaxIncomingConnections, ServerReadTimeout, ServerWriteTimeout, Handler)
  
  if err != nil {
    base.Log("Failed to start HTTP Server on %s: %v", addr, err)
  }
}
 
type ServerContext struct {
  config *ServerConfig
  HTTPClient  *http.Client

}
  

func NewServerContext(config *ServerConfig) *ServerContext {
  s := &ServerContext{
    config: config,
  }

  return s
}