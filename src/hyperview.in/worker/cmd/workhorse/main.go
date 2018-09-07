package main


import (
  "os"
  "fmt"
  "strings"
  "hyperview.in/server/base"  
  "hyperview.in/worker/utils/env_util"

  "hyperview.in/worker"
) 


type workerEnv struct {

  ServerAddr string `env:"SERVER_ADDR"`
  ApiServerIp string `env:"API_SERVER_IP"`
  ApiServerPort string `env:"API_SERVER_PORT"`
  ApiServerProtocol string `env:"API_SERVER_PROTOCOL" envDefault:"http://"` 

  WorkerIp string `env:"WORKER_IP"`
  FlowId string `env:"FLOW_ID"`
  TaskId string `env:"TASK_ID"`
  WorkingDir string `env:"WORKSPACE_DIR" envDefault:"/workspace"`

}

func main() {
  wenv := workerEnv{}

  err := env_util.Parse(&wenv)
  if err != nil {
    base.Log("[whWorker.main] Failed to parse : ", err)
    cmdError(err)
  }
  var api_server_addr string

  if wenv.ApiServerIp != "" && wenv.ApiServerPort != "" {
    api_server_addr = wenv.ApiServerProtocol + wenv.ApiServerIp + ":" + wenv.ApiServerPort
  }
  
  if  wenv.ServerAddr != "" {
    api_server_addr = wenv.ServerAddr
  }


  base.Log("[whWorker.main] Environment vars: ", wenv)
  
  wh := worker.NewWorkHorse(api_server_addr, wenv.FlowId, wenv.TaskId, wenv.WorkerIp, wenv.WorkingDir)
  err = wh.Init()
  
  if err != nil {
    
    if strings.Contains(err.Error(), "Invalid task status") {
      base.Log("[main] Skipping task as it is already complete.")
      os.Exit(0)
    }

    base.Debug("[main] Worker Initialization Error: ", err)
    cmdError(err)
  }

  // TODO: Add backoff here
  err = wh.DoWork()
  if err != nil {
    base.Debug("[main] Worker Failed doing work. Error: ", err)
    cmdError(err)    
  }

  err = wh.Shutdown()

  if err != nil {
    base.Debug("[main] Worker Failed during shutdown. Error: ", err)
    cmdError(err)    
  }  

}

func cmdError(err error) {
  if errString := strings.TrimSpace(err.Error()); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}
 