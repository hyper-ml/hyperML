package flow

import (  
  "fmt"
  "time"

  "hyperview.in/server/base"
  "hyperview.in/server/core/storage"
  tasks_pkg "hyperview.in/server/core/tasks"
  db_pkg "hyperview.in/server/core/db"
  ws "hyperview.in/server/core/workspace"
) 

type FlowServer struct {
  db *db_pkg.DatabaseContext
  fe FlowEngine	
  qs *queryServer
  obj storage.ObjectAPIServer 
  wsapi ws.ApiServer
  ns string
  quit chan int
}

func NewFlowServer(db *db_pkg.DatabaseContext, 
  kubens string, 
  obj storage.ObjectAPIServer,
  wsapi ws.ApiServer) *FlowServer {
    
  qs:= NewQueryServer(db)
  fe:= NewFlowEngine(qs, db) 
  
  quit:= make(chan int)
  
  go fe.master(quit)
    
  fs := &FlowServer {
    db: db,
    ns: kubens,
    qs: qs,
    fe: fe,
    quit: quit,
    obj: obj,
    wsapi: wsapi,
  }

  return fs
    
}

func (fs *FlowServer) Close() {
  close(fs.quit)
}
 
func errorCompletedTask() error{
  return fmt.Errorf("Invalid Status Update. The task is already completed.")
}

func errInvalidWorkerForTask(workerId string) error {
  return fmt.Errorf("Invalid worker for this task: %s", workerId)
}

func (fs *FlowServer) GetFlowAttr(flowId string) (*FlowAttrs, error) {
  return fs.qs.GetFlowAttr(flowId)
}


func (fs *FlowServer) RegisterWorker(flowId string, taskId string, ipaddr string) (*WorkerAttrs, error) {
  work_attr, err := fs.qs.registerW(flowId, taskId, ipaddr)
  
  if err != nil {
    base.Log("[FlowServer.RegisterWorker] Failed to register worker attributes: ", err)
    return nil, err
  }

  return work_attr, nil
}


func (fs *FlowServer) DetachTaskWorker(workerId, flowId, taskId string) error {
  return fs.qs.DetachTaskWorker(workerId, flowId, taskId)
}

func (fs *FlowServer) LaunchFlow(flr NewFlowLaunchRequest) (NewFlowLaunchResponse, error) {
  var commit_id string = flr.Commit.Id

  if  commit_id == "" {
    // get or open commit 
    
  }

  resp:= NewFlowLaunchResponse{}

  flow, task_status, err := fs.fe.LaunchFlow(flr.Repo.Name, flr.Branch.Name, flr.Commit.Id, flr.CmdString)
  if err != nil {
    resp.TaskStatus = tasks_pkg.TASK_FAILED
    return resp, err
  }

  resp.TaskStatusStr = tasks_pkg.TaskStatusByKey(task_status)
  resp.Flow = flow
  
  return resp, nil
}

func (fs *FlowServer) UpdateWorkerTaskStatus(worker Worker, tsr *TaskStatusChangeRequest) (*TaskStatusChangeResponse, error) {
  flow_attrs, err := fs.updateWorkerTaskStat(worker.Id, tsr.Flow.Id, tsr.Task.Id, tsr.TaskStatus)
  if err != nil {
    return nil, err
  }

  return &TaskStatusChangeResponse {
    FlowAttrs: flow_attrs,
  }, nil
}

func (fs *FlowServer) updateWorkerTaskStat(workerId string, flowId string, taskId string, newStatus tasks_pkg.TaskStatus) (*FlowAttrs, error) {
  task_worker:= fs.qs.GetWorkerByTaskId(flowId, taskId) 
  
  if task_worker.Worker.Id != workerId {
    base.Log("[FlowServer.updateWorkerTaskStat] Invalid Worker error (flowId, taskId, workerId): ", flowId, taskId, workerId)
    return nil, errInvalidWorkerForTask(workerId)
  }

  return fs.updateTaskStatus(flowId, taskId, newStatus)
}  

func (fs *FlowServer) updateTaskStatus(flowId string, taskId string, newStatus tasks_pkg.TaskStatus) (*FlowAttrs, error) {
  base.Log("[FlowServer.updateTaskStatus] newStatus: ", newStatus)
  
  task_attrs, err  := fs.qs.GetTaskByFlowId(flowId, taskId)
  if err != nil {
    return nil, err
  }
  
  task_attrs.Status = newStatus

  switch s := newStatus; s {
  
  case tasks_pkg.TASK_CREATED:
    task_attrs.Created = time.Now()
  

  case tasks_pkg.TASK_COMPLETED:
    //TODO: should come in the request from worker
    task_attrs.Completed = time.Now()
  
  case tasks_pkg.TASK_INITIATED:
    if task_attrs.Completed.IsZero() {
      task_attrs.Started = time.Now()
    } else {
      return nil, errorCompletedTask() 
    }
  
  case tasks_pkg.TASK_FAILED:
    if task_attrs.Completed.IsZero() {
      task_attrs.Failed = time.Now()
    } else {
      return nil, errorCompletedTask()
    }
  } 

  if err := fs.qs.UpdateTaskByFlowId(flowId, *task_attrs); err == nil {
    return  fs.qs.GetFlowAttr(flowId) 
  }

  return nil, err
}


/*func (fs *FlowServer) StartWorker(flowId, taskId string) error {
  return fs.fe.StartFlow(flowId, taskId)
}*/

func (fs *FlowServer) GetFlowLogPath(flowId string) string {
  return  "logs/flows/" + flowId + ".log"
}


func (fs *FlowServer) GetTaskLogPath(taskId string) string {
  return  "task_logs/" + taskId + ".log"
}

func (fs *FlowServer) GetTaskLog(flowId string) ([]byte, int, error) {
  //fs.obj.
  file_name := fs.GetFlowLogPath(flowId)
  return fs.obj.GetObject(file_name, 0, 0)
}

func getOutRepoName(flow_id string) string {
  return "flow/" + flow_id + "/out"
}

func (fs *FlowServer) createOutRepo(flow Flow) (*ws.RepoAttrs, error) {
  repo_name:= getOutRepoName(flow.Id) 
  repo_attrs, err:= fs.wsapi.InitRepo(repo_name)
  if err != nil {
    base.Log("[FlowServer.createOutRepo] Failed to create output repo: ", err)
    return nil, err
  }

  return repo_attrs, nil
}

func (fs *FlowServer) getOutRepo(flow Flow) (*ws.RepoAttrs) {
  repo_name := getOutRepoName(flow.Id)
  repo_attrs, _ := fs.wsapi.GetRepoAttrs(repo_name)

  return repo_attrs
}

func (fs *FlowServer) GetOrCreateOutRepo(flow Flow) (*ws.RepoAttrs, error) {
  if !fs.wsapi.CheckRepoExists(getOutRepoName(flow.Id)) {
    return fs.createOutRepo(flow)
  } 
  
  return fs.getOutRepo(flow), nil
}