package flow

import(
  "io"
  tsk_pkg "github.com/hyper-ml/hyperml/server/pkg/tasks"
  db_pkg "github.com/hyper-ml/hyperml/server/pkg/db"
  "github.com/hyper-ml/hyperml/server/pkg/storage"
  "github.com/hyper-ml/hyperml/server/pkg/config"
)

 
type WorkerEventType string

const (
  WorkerInitError WorkerEventType = "WORKER_INIT_FAILED"
  WorkerAdded    WorkerEventType = "ADDED"
  WorkerModified WorkerEventType = "MODIFIED"
  WorkerDeleted  WorkerEventType = "DELETED"
  WorkerError    WorkerEventType = "ERROR"
  WorkerFailed   WorkerEventType = "FAILED"
  WorkerSucceeded WorkerEventType = "SUCCEEDED"

  DefaultChanSize int32 = 100
)

func WorkerEventFromStr(evt string) WorkerEventType{
  return WorkerEventType(evt)
}

type WorkerEvent struct {
  Type WorkerEventType

  Worker Worker
  Flow Flow
  Task tsk_pkg.Task
}

type emptyWatch chan WorkerEvent

type WorkerPool interface {
  WorkerExists(flowId, taskId string) bool
  AssignWorker(taskId string, flowAttrs *FlowAttrs, masterIp string, masterPort, masterExtPort int32) error
  ReleaseWorker(flow Flow ) error
  
  Watch(eventCh chan WorkerEvent)
  CloseWatch() 

  SaveWorkerLog(worker Worker, flow Flow) (error)
  LogStream(flowId string) (io.ReadCloser, error)
}

func NewWorkPoolWatcher() (chan WorkerEvent) {
  evt := make(chan WorkerEvent) 
  return evt
}


//TODO: add worker limits
func NewWorkerPool(config *config.Config, db db_pkg.DatabaseContext, logger storage.ObjectAPIServer) (*PodKeeper, error) {
  return NewDefaultPodKeeper(config, db, logger)
}


