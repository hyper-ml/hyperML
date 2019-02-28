package main


import (
  "os"
  "fmt"
  "strings"  
  filepath_pkg "path/filepath" 

  "hyperflow.in/server/pkg/base"  
  "hyperflow.in/worker/utils/env_util"
  "hyperflow.in/worker/utils"
  "hyperflow.in/worker"
  ws "hyperflow.in/worker/worker_server"
) 


const (
  LOG_DIRECTORY = "log"
  LOG_FILENAME = "task.log"
  FILE_SEP =  "/"
)

type workerEnv struct {

  ServerAddr string `env:"SERVER_ADDR"`
  ApiServerIp string `env:"API_SERVER_IP"`
  ApiServerPort string `env:"API_SERVER_PORT"`
  ApiServerProtocol string `env:"API_SERVER_PROTOCOL" envDefault:"http://"` 

  WorkerIp string `env:"WORKER_IP"`
  FlowId string `env:"FLOW_ID"`
  TaskId string `env:"TASK_ID"`
  WorkingDir string `env:"WORKSPACE_DIR"`
  HomeDir string `env:"HOME_DIR"`

}

func main() {
  var hfl_server string
  wenv := workerEnv{}

  // parse env variables
  err := env_util.Parse(&wenv)
  if err != nil {
    base.Log("[whWorker.main] Failed to parse : ", err)
    cmdError(err)
  }

  // set server conn addr
  if wenv.ApiServerIp != "" && wenv.ApiServerPort != "" {
    hfl_server = wenv.ApiServerProtocol + wenv.ApiServerIp + ":" + wenv.ApiServerPort
  }
  
  if wenv.ServerAddr != "" {
    hfl_server = wenv.ServerAddr
  } 
  
  // log router 

  log_handler := utils.NewLogHandler(getLogName(wenv.WorkingDir), 25)
  
  go func() {
    
    wh := worker.NewWorkHorse(nil, hfl_server, wenv.FlowId, wenv.TaskId, wenv.WorkerIp, wenv.WorkingDir, log_handler)
    err = wh.Init()
    
    if err != nil {
      
      if strings.Contains(err.Error(), "Invalid task status") {
        base.Println("[main] Skipping task as it is already complete.")
        return
      }

      base.Error("[main] Worker Initialization Error: ", err)
      // restart worker as registration failed
      cmdError(err)
    }

    if err := wh.InitEnvironment(); err != nil {
      return
    }
    
    // TODO: Add infinite backoff here
    err = wh.DoWork()
    if err != nil {
      base.Error("[main] Worker Failed doing work. Error: ", err)
      return
    }

    err = wh.Shutdown()

    if err != nil {
      base.Error("[main] Worker Failed during shutdown. Error: ", err)
    }  
    return
  }()

  ws.WorkerServer()
 
}

func getLogName(home string) string {
  log_path := filepath_pkg.Join(home, LOG_DIRECTORY)
  return log_path + FILE_SEP + LOG_FILENAME
}

func cmdError(err error) {
  if errString := strings.TrimSpace(err.Error()); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}
 