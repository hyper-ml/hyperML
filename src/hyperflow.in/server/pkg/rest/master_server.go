package rest

import (
  "net/http"
  "hyperflow.in/server/pkg/config"
  "hyperflow.in/server/pkg/base"
)
  
var DefaultMaxIncomingConnections = 0
var ServerReadTimeout = 200
var ServerWriteTimeout = 200
  
var master *Master

type Master struct {
  Config *config.Config
}

func NewMasterServer(c *config.Config) *Master {
  return &Master {
    Config: c,
  }
}
 
func (m *Master) ListenAddress() string{
  return m.Config.GetListenAddr()
}

func (m *Master) Serve(Handler http.Handler) {
  err := ListenAndServeHTTP(m.ListenAddress(), DefaultMaxIncomingConnections, ServerReadTimeout, ServerWriteTimeout, Handler)
  
  if err != nil {
    base.Log("failed to start HTTP Server on %s: %v", m.ListenAddress(), err)
  }
}
   


  