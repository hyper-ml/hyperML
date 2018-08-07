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
  WorkerIp string `env:"WORKER_IP"`
  FlowId string `env:"FLOW_ID"`
  TaskId string `env:"TASK_ID"`
  WorkingDir string `env:"WORKSPACE_DIR" envDefault:"\workspace"`

}

func main() {
  wenv := workerEnv{}

  err := env_util.Parse(&wenv)
  if err != nil {
    base.Log("[whWorker.main] Failed to parse : ", err)
    cmdError(err)
  }
  base.Log("[whWorker.main] Environment vars: ", wenv)
  
  wh := worker.NewWorkHorse(wenv.ServerAddr, wenv.FlowId, wenv.TaskId, wenv.WorkerIp, wenv.WorkingDir)
  err = wh.Register()
  
  if err != nil {
    base.Debug("[main] Worker Registration Error: ", err)
    cmdError(err)
  }

  // TODO: Add backoff here
  wh.Do()
}

func cmdError(err error) {
  if errString := strings.TrimSpace(err.Error()); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}
 