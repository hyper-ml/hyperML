package rest_server

import(
  "flag"
)
var DefaultInterface = ":9999"
var DefaultMaxIncomingConnections = 0
var ServerReadTimeout = 200
var ServerWriteTimeout = 200

func ParseCommandLine() (config *ServerConfig) {
  addr := flag.String("interface", DefaultInterface, "Address to bind to")
  flag.Parse()
  return NewServerConfig(addr)
}

func RunServer(config *ServerConfig) {
  server_context := NewServerContext(config)
  config.Serve(*config.Interface, CreatePublicHandler(server_context))
}


func WorkerServer() {
  config := ParseCommandLine()
  RunServer(config)
}