package flow

import(
  tsk_pkg "hyperview.in/server/core/tasks"
  db_pkg "hyperview.in/server/core/db"
  
)

 
type WorkerEventType string

const (
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


// Launch and manages workers 

type WorkerPool interface {
  WorkerExists(flowId, taskId string) bool
  AssignWorker(taskId string, flowAttrs *FlowAttrs) error
  ReleaseWorker(flow Flow ) error
  
  Watch(eventCh chan WorkerEvent)
  CloseWatch() 

  SaveWorkerLog(worker Worker, flow Flow) (error)
}

func NewWorkPoolWatcher() (chan WorkerEvent) {
  evt := make(chan WorkerEvent) 
  return evt
}


//TODO: add worker limits
func NewWorkerPool(db *db_pkg.DatabaseContext) *PodKeeper {
  return NewDefaultPodKeeper(db)
}


