package flow

import ( 
  "fmt"
  "time"
  //"context"
  "golang.org/x/sync/errgroup"


  //"hyperview.in/server/base/backoff"
  task_pkg "hyperview.in/server/core/tasks"
  "hyperview.in/server/base"
  "hyperview.in/server/core/storage"
  db_pkg "hyperview.in/server/core/db"
)

type FlowEngine interface{
  StartFlow(flowId, taskId string) (*FlowAttrs, error)
  LaunchFlow(repoName string, branchName string, commitId string, cmdString string) (*FlowAttrs, error)
}


type flowEngine struct {
  qs *queryServer
  db *db_pkg.DatabaseContext
  wpool WorkerPool
  namespace string
  defaultImage string
  dockerPullPolicy string

  // storage details - add later

}

func NewFlowEngine(qs *queryServer, db *db_pkg.DatabaseContext, logger storage.ObjectAPIServer) *flowEngine{
  
  wp := NewWorkerPool(db, logger)

  return &flowEngine{
    qs: qs,
    db: db,
    wpool: wp,
  }
}

func TaskWorkerExistsError(flowId, taskId string) error{
  return fmt.Errorf("[flowEngine] Worker already executing this flow task: %s %s", flowId, taskId)
}

func ErrTaskComplete() error {
  return fmt.Errorf("invalid_task_update: task already complete")
}

func InvalidFlowIdError(flowId string) error {
  return fmt.Errorf("[flowEngine] Invalid flow Id: %s", flowId)
}

func InvalidFlowParamsError(flowId string) error {
  return fmt.Errorf("[flowEngine] Invalid flow parameter: %s", flowId)
}

 
// monitor new messages from the worker pod or update on flow status
// end pods or mark flow completion 

func (fe *flowEngine) master(quit chan int) {
  
  event_chan := make(chan interface{})
  fe.db.Listener.RegisterObject(event_chan, Flow{})
  base.Log("[flowEngine.master] Starting Flow Master")

  w_chan:= NewWorkPoolWatcher()
  
  go fe.wpool.Watch(w_chan)


  for {
    select {
      case evtval, ok := <- event_chan:
        if !ok {
          return
        }

        flow_attrs, ok := evtval.(*FlowAttrs)
        
        if !ok {
          base.Log("[flowEngine.master] Oops not a flow record")
          break
        } 

        base.Debug("[flowEngine.master] Change event on  Flow/task Status. Id: ", flow_attrs.Flow.Id)
        
        

        // stop worker if flow is completed with success or error 
        if flow_attrs.isComplete() {
          base.Info("[flowEngine.master] Releasing worker as flow status > 11")
          _ = fe.wpool.SaveWorkerLog(Worker{}, flow_attrs.Flow)
          _ = fe.wpool.ReleaseWorker(flow_attrs.Flow)
        }

      case w_evt, ok := <- w_chan: 
        base.Debug("[flowEngine.master] Received kube event: ", w_evt.Type) 
        if !ok {
          break
        }

        flow := w_evt.Flow
        //worker := w_evt.Worker

        /*if worker.PodId != "" {
          // TODO: multiple writes happening
          err := fe.wpool.SaveWorkerLog(worker, flow)
          if err != nil {
            base.Debug("[flowEngine.master] Save worker log error: ", err)
          }
        }*/
        
        if w_evt.Type == "ERROR" || w_evt.Type == "DELETED" || 
           w_evt.Type == WorkerError || w_evt.Type == WorkerDeleted {
          base.Debug("[FlowEngine.Master] Event type error found ", w_evt.Type, flow.Id)
          
          if flow.Id != "" {
            var task_id string
            // retrieve task Id 
            task_id = flow.Id 
            fe.processFlowError(flow.Id, task_id, string(w_evt.Type))
            // call update on flow. Set status to message
          }
        }

        // on failure shutdown and update status. Check why error is not being caputred
        
      case <-quit:
        base.Log("[flowEngine.master] Quiting flow Engine master..")
        return
    }
  }

  return

  /*op := func() error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    eventCh := make(chan interface{})

    fe.db.Listener.RegisterObject(eventCh, Flow{})

    for {
      select {
        case val, ok := <- eventCh:
          base.Log("Received update:", val)
      }
    }

    return nil
  }

  backoff.RetryNotify(op, 
    backoff.NewExponentialBackOff(), 
    func(error, time.Duration) {
    })*/
}


 
func (fe *flowEngine) StartFlow(flowId, taskId string) (*FlowAttrs, error) {
  flow_attrs, err:= fe.qs.GetFlowAttr(flowId)

  if err != nil {
    return nil, InvalidFlowIdError(flowId)
  }

  if !fe.wpool.WorkerExists(flowId, taskId) {
    err = fe.wpool.AssignWorker(taskId, flow_attrs)
    
    if err != nil {      
      if err = fe.processFlowError(flowId, taskId, err.Error()); err != nil {
        return flow_attrs, err
      }
      return flow_attrs, err
    }

    return flow_attrs, nil

  } else {
    // TODO: check if worker is active if not then flush it and restart a new worker
    return nil, TaskWorkerExistsError(flowId, taskId)
  } 
  return flow_attrs, nil

}

func (fe *flowEngine) processFlowError(flowId, taskId, message string) error {
 
  var eg errgroup.Group
  
  f := Flow { 
        Id: flowId,
    }

  //release worker from pool
  // check if worker is assigned. Release only then
  eg.Go(func () error {  
      if task_worker := fe.qs.GetWorkerByTaskId(flowId, taskId); task_worker == nil {
        base.Debug("[flowEngine.processFlowError] No task worker for this flow task :", flowId, taskId)
        return nil
      }
      
      err := fe.wpool.ReleaseWorker(f)
      return err 
    })
  
  // update status of flow and remove worker assignment 
  eg.Go(func () error {  
      err := fe.updateTaskStatus(flowId, taskId, task_pkg.TASK_FAILED)
      err = fe.workerCleanUp(f)
      return err 
    })

  return eg.Wait()
}

func (fe *flowEngine) Logworker(flow Flow, worker Worker) error { 
  
  //var w io.Writer 
  err := fe.wpool.SaveWorkerLog(worker, flow)
  return err
}

func (fe *flowEngine) workerCleanUp(flow Flow) error {
  return fe.wpool.ReleaseWorker(flow)
}

// create flow with a single task
func (fe *flowEngine) createSimpleFlow(repoName string, branchName string, commitId string, cmdString string) (*FlowAttrs, error) {
  wdir := "/workspace"
 
  mount_map := task_pkg.NewMountConfig(repoName, branchName, commitId, wdir, 0)
  flow_config := &FlowConfig{
    MountMap: mount_map,
  }

  // new flow attr rec
  new_flow :=  NewFlowAttrs(flow_config) 

  // insert task
  task_config:= task_pkg.NewTaskConfig(cmdString, nil, wdir, mount_map)
  
  // add task 
  _ = new_flow.AddTask(task_config)

  err :=  fe.qs.InsertFlow(new_flow)

  return new_flow, err
}


func (fe *flowEngine) LaunchFlow(repoName string, branchName string, commitId string, cmdString string) (*FlowAttrs, error) {
  // 1. create a new flow - task
  // 2. start flow 

  var err error 
  flow_attrs, err := fe.createSimpleFlow(repoName, branchName, commitId, cmdString)
  
  if err != nil {
    base.Error("[fe.LaunchFlow] Failed to create simple flow: ", err)
    return nil, err
  }

  task_attrs:= flow_attrs.FirstTask()
  
  if task_attrs == nil {
    return nil, fmt.Errorf("[LaunchFlow] No task to run")
  }

  flow_attrs, err = fe.StartFlow(flow_attrs.Flow.Id, task_attrs.Task.Id)

  if err != nil {
    return nil, err
  }

  return flow_attrs, nil
}

/*
func (fe *flowEngine) updateWorkerTaskStatus(workerId string, flowId string, taskId string, newStatus  task_pkg.TaskStatus) (error) {
  
  w:= Worker { Id: workerId }
  f:= Flow { Id: flowId }
  _ = fe.Logworker(f, w)

  return fe.updateTaskStatus(flowId, taskId, newStatus)
} 
*/

func (fe *flowEngine) updateTaskStatus(flowId string, taskId string, newStatus task_pkg.TaskStatus) (error) {
   
  task_attrs, err  := fe.qs.GetTaskByFlowId(flowId, taskId)
  if err != nil {
    return err
  }

  if task_attrs.Status == task_pkg.TASK_COMPLETED {
    return ErrTaskComplete()
  }
  
  task_attrs.Status = newStatus

  switch s := newStatus; s {
  
  case task_pkg.TASK_CREATED:
    task_attrs.Created = time.Now()
  
  case task_pkg.TASK_COMPLETED:
    //TODO: should come in the request from worker
    task_attrs.Completed = time.Now()
  
  case task_pkg.TASK_INITIATED:
    if task_attrs.Completed.IsZero() {
      task_attrs.Started = time.Now()
    } else {
      return errorCompletedTask() 
    }
  
  case task_pkg.TASK_FAILED:
    if task_attrs.Completed.IsZero() {
      task_attrs.Failed = time.Now()
    } else {
      return errorCompletedTask()
    }
  } 

  if err := fe.qs.UpdateTaskByFlowId(flowId, *task_attrs); err == nil {
    return  nil
  }

  return err
}





