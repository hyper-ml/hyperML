package worker

import (
  "fmt"
  "time"
  "context"

  "hyperview.in/server/base"
  flw "hyperview.in/server/core/flow"
  tsk "hyperview.in/server/core/tasks"

  "hyperview.in/worker/utils"  
  //"hyperview.in/worker/pkg/exec"
)


type WorkHorse struct {
  flowId string
  taskId string

  flowAttrs *flw.FlowAttrs
  taskAttrs *tsk.TaskAttrs

  workingDir string
  
  workerIp string

  tasks map[string]*tsk.TaskAttrs
  tlogs map[string]utils.Logger
  
  // TODO: add task environments

  started time.Time
  wc *WorkerClient
  attrs *flw.WorkerAttrs
  ctx context.Context
}

//
func NewWorkHorse(serverAddr string, flowId string, taskId string, workerIp string, workingDir string) *WorkHorse {
  wc, _ := NewWorkerClient(serverAddr)
  tasks := make(map[string]*tsk.TaskAttrs) 
  tlogs := make(map[string]utils.Logger) 

  return &WorkHorse {
    flowId: flowId,
    taskId: taskId,
    workerIp: workerIp,
    workingDir: workingDir,

    wc: wc,
    tasks: tasks,
    tlogs: tlogs,

    attrs: &flw.WorkerAttrs{}, 
    started: time.Now(),
    ctx: context.Background(),
  }
}

func (w *WorkHorse) Register() (error) {
  worker_attrs, err := w.wc.RegisterWorker(w.flowId, w.taskId, w.workerIp)
  if worker_attrs!= nil {
    base.Debug("[worker.Register] Registered Worker Attributes: ", *worker_attrs)
  }
  w.attrs = worker_attrs 

  return err
}


func (w *WorkHorse) Do() error {
  return nil
}



func (w *WorkHorse) init() error {
  
  if err := w.setFlowAttrs(); err != nil {
    return err
  }

  return nil
}

func (w *WorkHorse) mountVolume() error {

  return nil
}

func (w *WorkHorse) setFlowAttrs() error {
  flow_attrs, err:= w.wc.FetchFlowAttrs(w.flowId)

  if err != nil || flow_attrs == nil {
    base.Log("[WorkHorse.setFlowAttrs] Failed to fetch flow attributes: ", w.flowId, err)
    return err
  }

  if flow_attrs == nil || flow_attrs.Tasks == nil {
    return fmt.Errorf("[WorkHorse.setFlowAttrs] This flow has no tasks.", w.flowId)
  }

  task_attrs, ok := flow_attrs.Tasks[w.taskId]
  if !ok {
    return fmt.Errorf("[WorkHorse.setFlowAttrs] Assigned task does not exist in this flow has no tasks.", w.flowId, w.taskId)
  }
  w.flowAttrs = flow_attrs
  w.taskAttrs = &task_attrs
  return nil
}


/*
func (w *WorkHorse) StartWorker(tasksIds []string) error {
  var err error
  
  // TODO: add err group
  for _, id := range tasksIds {
    err = w.AddTask(id)
    if err != nil {
      base.Log("[w.StartWorker] Failed to add tasks in worker queue: ", id, err)
      return err
    }
  }

  return nil
}


func (w *WorkHorse) AddTask(Id string) error {
  task_attrs, err := w.wc.FetchTaskAttrs(Id)
  if err != nil {
    base.Log("[Worker.AddTask] Failed to fetch task info: ", Id)
    return err
  }

  w.tasks[Id] = task_attrs
  w.tlogs[Id] = utils.NewTaskLogger(Id)
  return nil
}


func (w *WorkHorse) CompleteTask(Id string) error {
  
  task_attrs, err := w.wc.FetchTaskAttrs(Id)
  if err != nil {
    base.Log("[w.StartTask] Fetch Task Info Error: ", err)
  }

  task_logger := w.tlogs[Id]
  ctx := w.ctx

  if task_attrs.Cmd == "" {
    base.Log("[w.StartTask] This task has no command. ", task_attrs)
    return fmt.Errorf("[StartTask] This task has no command")
  }

  c_handle := exec.CommandContext(ctx, task_attrs.Cmd, task_attrs.CmdArgs...)
  c_handle.Dir = task_attrs.WorkDir

  c_handle.Stdout = task_logger
  c_handle.Stderr = task_logger
  c_handle.Env = w.getTaskUserEnv(Id)

  err = c_handle.Start()

  if err != nil {
    base.Debug("[w.StartTask] Failed to initiate task command Id: ", Id, c_handle, err)
    return fmt.Errorf("[StartTask] Failed to initiate task command: %v %v", c_handle, err)
  }

  base.Debug("[w.StartTask] Task initiated with no errors. Going to wait on process now.")

  procState, err := c_handle.Process.Wait()
  if err != nil {
    base.Debug("[w.StartTask] Failed while waiting for task command Id: ", Id, c_handle, err)
    return fmt.Errorf("[StartTask] Failed while waiting for task command: %v %v", c_handle, err)
  }

  if ctx != nil {
    select {
      case <- ctx.Done():

        base.Debug("[w.StartTask] Was waiting on process. Received Done")
        if err = ctx.Err(); err != nil {
          return err
        }
      default: 
        //do nothing
    }
  }

  base.Debug("[w.StartTask] Initiating IO close on command handle")
  err = c_handle.WaitForIOClose(procState, err)

  base.Debug("[w.StartTask] IO close complete")

  return nil
} 



func (w *WorkHorse) getTaskUserEnv(Id string) []string{
  return nil
}



 */
 






