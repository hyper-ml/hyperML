package flow

import(
  "fmt"
  "time"
  "hyperview.in/server/base"
  "encoding/json"
  "hyperview.in/server/core/utils"
  db_pkg "hyperview.in/server/core/db"
  
  . "hyperview.in/server/core/tasks"
)

type queryServer struct{
  db *db_pkg.DatabaseContext
}


func NewQueryServer(db *db_pkg.DatabaseContext) *queryServer {
  return &queryServer{
    db: db,
  }
}

func (qs *queryServer) flowKey(Id string) string {
  return "flow:" + Id
}
  

func (qs *queryServer) GetFlowAttr(flowId string) (*FlowAttrs, error) {
  flow_key := qs.flowKey(flowId)

  FlowAttrs_raw, err := qs.db.Get(flow_key)

  if err != nil {
    base.Log("[queryServer.GetFlowAttr] Flow Record for this ID does not exist. ", flowId)
    return nil, err
  }

  FlowAttrs:= FlowAttrs{}
  err = json.Unmarshal(FlowAttrs_raw, &FlowAttrs)
  if err != nil {
    base.Log("[queryServer.GetFlowAttr] Failed to convert raw object to Flow Info", flowId)
    return nil, err
  }

  return &FlowAttrs, nil
}


func (qs *queryServer) InsertFlow(flowAttr *FlowAttrs) error {
  flow_key := qs.flowKey(flowAttr.Flow.Id)

  err := qs.db.Insert(flow_key, flowAttr)

  if err != nil {
    base.Log("[flowServer.InsertFlow] Failed to Insert flow:", err)
    return err
  }

  return nil
}


func (qs *queryServer) UpdateFlow(flowId string, flowAttr *FlowAttrs) error {
  
  flow_key := qs.flowKey(flowId)
  t := Flow{}

  err := qs.db.UpdateAndTrack(flow_key, flowAttr, t)

  if err != nil {
    base.Log("[flowServer.UpdateFlow] Failed to start flow:", err)
    return err
  }

  return nil
}

func validForDelete(status FlowStatus) bool{
  switch status  {
    case CANCELLED,
         CREATED:
      return true
  }
  return false
}

func (qs *queryServer) DeleteFlow(flowId string) error {
  var err error
  flow_key :=  qs.flowKey(flowId)

  FlowAttrs, err:= qs.GetFlowAttr(flowId)
  if err != nil {
    base.Log("[queryServer.DeleteFlow] Invalid flow Id: ", flowId)
    return err
  }

  if !validForDelete(FlowAttrs.Status) {
    base.Log("[queryServer.DeleteFlow] The status of this flow is invalid for delete. Check docs for valid statuses.", flowId)
    return fmt.Errorf("Invalid flow Status for deletion: %s", flowId)
  }

  err = qs.db.SoftDelete(flow_key)
  return err
}



func taskWorkerKey(flowId, taskId string) string {
  return "flow:" + flowId + ":task:" + taskId + ":worker"
}

func (qs *queryServer) GetWorkerByTaskId(flowId, taskId string) *FlowTaskWorker {
  w_key:= taskWorkerKey(flowId, taskId)

  ftw_raw, err := qs.db.Get(w_key)

  if err != nil {
    base.Log("[queryServer.GetWorkerByTaskId] Failed to retrieve Flow Task Worker record: ", w_key)
    return nil
  } 

  if ftw_raw == nil {
    return nil
  }

  ftw := FlowTaskWorker{}
  err = json.Unmarshal(ftw_raw, &ftw)
  return &ftw
}

func (qs *queryServer) InsertTaskWorker(f Flow, t Task, w Worker) (*FlowTaskWorker, error) {
  w_key:= taskWorkerKey(f.Id, t.Id)

  ftw := &FlowTaskWorker {
      Worker: w,
      Flow: f,
      Task: t,
      Created: time.Now(),
    }

  err:= qs.db.Insert(w_key, ftw)
  if err != nil {
    return nil, err
  }

  return ftw, nil
}


func workerKey(workerId string) string {
  return "worker:" + workerId
}

func (qs *queryServer) GetWorkerAttrs(workerId string) (*WorkerAttrs, error) {
  w_key := workerKey(workerId)
  
  wattrs_raw, err := qs.db.Get(w_key)
  if err != nil {
    base.Log("[queryServer.GetWorkerAttrs] Failed to fetch worker attributes: ", err)
    return nil, err
  }

  work_attrs := &WorkerAttrs{}
  err = json.Unmarshal(wattrs_raw, work_attrs)
  
  return work_attrs, nil 
}
 

//register worker
func (qs *queryServer) registerW(flowId string, taskId string, ipAddress string) (*WorkerAttrs, error) {
  
  // check if a worker already exists agains this task
  // current flow task worker record
  cftw_rec := qs.GetWorkerByTaskId(flowId, taskId) 
  if cftw_rec != nil {
    if cftw_rec.Worker.Id != "" {

      // get current worker attributes
      cw_attrs, _ := qs.GetWorkerAttrs(cftw_rec.Worker.Id)

      if cw_attrs.Ip == ipAddress { 
        // requester is registering again. May be after a failure
        // send worker attributes for sync
        return cw_attrs, nil      
      } else {
        // requester is different that existing worker on task
        // This could be conflict or due to dead worker
        return nil, fmt.Errorf("[RegisterWorker] A worker is already assigned to this task. Wait for Reaper to collect the failed worker.")
      }
    }
  }  

  // TODO: ID no worker for this task. Proceed
  newWorkerId := utils.NewUUID()
  w_key := workerKey(newWorkerId)
 

  // check if a worker us already registered 
  w_attrs := &WorkerAttrs {
    Worker: Worker {
      Id: newWorkerId,
    },
    Flow: Flow {
      Id: flowId, 
    },
    Task: Task {
      Id: taskId,
    },
    Started: time.Now(),
    Ip: ipAddress,
    Status: WORKER_REGISTERED,
  }

  err := qs.db.Insert(w_key, w_attrs)
  if err != nil {
    return nil, err 
  }

  _, err = qs.InsertTaskWorker(w_attrs.Flow, w_attrs.Task, w_attrs.Worker)
  if err != nil {
    // TODO: delete worker too
    return nil, err
  }

  return w_attrs, nil
}


func (qs *queryServer) unregisterW(workerId string) error {
  
  worker_attrs, err := qs.GetWorkerAttrs(workerId)
  if err != nil {
    base.Log("[queryServer.unregisterW] Failed to retrieve worker: ", workerId)
    return err
  }

  // insert nil worker
  _, err = qs.InsertTaskWorker(worker_attrs.Flow, worker_attrs.Task, Worker{})
  
  if err != nil {
    return err
  }

  return nil
}





