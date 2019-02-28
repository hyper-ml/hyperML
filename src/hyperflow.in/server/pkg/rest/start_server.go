package rest

import(
  "hyperflow.in/server/pkg/base"
  hfConfig "hyperflow.in/server/pkg/config"
)

// todo: ssl
func StartServer(listenIp, listenPort, configPath string) {
   
  c, err := hfConfig.NewConfig(listenIp, listenPort, configPath)
  if err != nil {
    raisePanic(err)
  }

  m := NewMasterServer(c) 
  runServer(m)  
}

func runServer(m *Master) {
  
  sc, err := NewServerContext(m.Config)  
  if err != nil {
    raisePanic(err)
  } 

  base.Println("Starting server on %s ...", m.ListenAddress())
  m.Serve(CreatePublicHandler(sc))
}


func raisePanic(err error) {
  panic(err)
}

