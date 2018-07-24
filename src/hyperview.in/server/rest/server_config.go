package rest

import (
  "net/http"
  "flag"
  "hyperview.in/server/base"
)

type DatabaseConfig struct {
  Name string
  User string
  Pass string
}

type ServerConfig struct {
  Interface *string
  AdminInterface *string
  BaseDir string
  DatabaseConfig *DatabaseConfig

  // Add logging
}

var config *ServerConfig

var DefaultInterface = ":8888"
var DefaultAdminInterface = "127.0.0.1:8889"
var DefaultMaxIncomingConnections = 0
var ServerReadTimeout = 200
var ServerWriteTimeout = 200

func (config *ServerConfig) Serve(addr string, handler http.Handler) {
  err := ListenAndServeHTTP(addr, DefaultMaxIncomingConnections, ServerReadTimeout, ServerWriteTimeout, handler)
  
  if err != nil {
    base.Log("Failed to start HTTP Server on %s: %v", addr, err)
  }
}

func ParseCommandLine() {
  addr := flag.String("interface", DefaultInterface, "Address to bind to")
  adminAddr := flag.String("adminInterface", DefaultAdminInterface, "Address to bind admin interface to")
  

  flag.Parse()

  dbConfig := &DatabaseConfig{
    Name: "amp_db",
    User: "apple",
    Pass: "",
  }
  
  config = &ServerConfig{
      Interface:        addr,
      AdminInterface:   adminAddr,
      BaseDir:          "hyperview",
      DatabaseConfig:   dbConfig,
  }
}

func RunServer(config *ServerConfig) {
  sc := NewServerContext(config)

  base.LogInfo("Starting server on %s ...", *config.Interface)
  config.Serve(*config.Interface, CreatePublicHandler(sc))
}

func ServerMain() {
  ParseCommandLine()
  RunServer(config)
}