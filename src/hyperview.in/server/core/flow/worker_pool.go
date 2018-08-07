package flow
 


// Launch and manages workers 

type WorkerPool interface {
  AssignWorker(taskId string, flowAttrs *FlowAttrs) error
  DestroyWorker(taskId string, flowAttrs *FlowAttrs) error
}




