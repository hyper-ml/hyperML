package worker

import (
  "fmt"
  "time"
  "context"

  "hyperview.in/server/base"
  flw "hyperview.in/server/core/flow"
  tsk "hyperview.in/server/core/tasks"

  "hyperview.in/worker/utils"  
  fsys "hyperview.in/worker/file_system"
  "hyperview.in/worker/pkg/exec"
  . "hyperview.in/worker/api_client"

)

const (
  FileOpsConcurrency int = 5
  IgnoreEmptyTask bool = true
  IgnoreMountError bool = true
)


type WorkHorse struct {
  Id string

  flowId string
  taskId string

  flowAttrs *flw.FlowAttrs
  taskAttrs *tsk.TaskAttrs

  workingDir string
  repoFs *fsys.RepoFs
  workerIp string

  tasks map[string]*tsk.TaskAttrs
  tlogs map[string]utils.Logger
  
  // TODO: add task environments

  started time.Time
  wc *WorkerClient
  attrs *flw.WorkerAttrs
  ctx context.Context
}

func emptyTaskError(flowId, taskId string) error {
  return fmt.Errorf("The task is empty: %s %s", flowId, taskId)
}

func InvalidFlowTaskError() error {
  return fmt.Errorf("Empty Flow or task assigned to the worker")
}

func InvalidFlowTaskStatus(status tsk.TaskStatus) error {
  return fmt.Errorf("Invalid Status (%d) for running the task. Check docs for valid statuses. ", int(status))
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



func (w *WorkHorse) Init() error {

  if err := w.register(); err != nil {
    base.Log("[WorkHorse.Init] Failed to register worker for this flow-task: ", w.flowId, w.taskId, w.workerIp)
    base.Log("[WorkHorse.Init] Error: ", err)
    return err
  }

  if err := w.pullFlowAttrs(); err != nil {
    base.Log("[WorkHorse.Init] Failed to pull flow-task attributes: ", w.flowId, w.taskId, err)
    return err
  }

  if err := w.PushTaskStatus(tsk.TASK_ACCEPTED); err != nil {
    base.Log("[WorkHorse.Init] Failed to update task status to accepted: ", w.flowId, w.taskId, err)
    return err
  }

  if err := w.MountVolumes(); err != nil {
    base.Log("[WorkHorse.Init] Failed to mount repo volums: ", err)
    return err
  }

  return nil
}

// Does following:
// 1) marks task as running
// 2) Execute task command 
// 3) Waits for command completion
// 4) Marks task either failed or completed  

// TODO: Upload out dir and command std io/err 

func (w *WorkHorse) DoWork() error {
  var err error
  if w.flowId == "" || w.taskId == "" || w.taskAttrs == nil {
    return InvalidFlowTaskError()
  }

  if w.taskAttrs.Status >= tsk.TASK_RUNNING {
    base.Log("[WorkHorse.DoWork] The status of this task is not suitable for execution.", w.taskAttrs.Status)
    return InvalidFlowTaskStatus(w.taskAttrs.Status)
  }

  base.Debug("[WorkHorse.DoWork] Starting Work on flow-task : ", w.flowId, w.taskId)
  w.PushTaskStatus(tsk.TASK_RUNNING)

  task_logger := w.tlogs[w.taskId]
  ctx := w.ctx

  if w.taskAttrs.Cmd == "" {
    
    if IgnoreEmptyTask {
      base.Log("[WorkHorse.DoWork] This task has no commands to execute. ", w.taskAttrs)
      return w.PushTaskStatus(tsk.TASK_COMPLETED)
    }
    w.PushTaskStatusWithMessage(tsk.TASK_FAILED, "The task is empty.")
    return emptyTaskError(w.flowId, w.taskId)
  }

  command_h := exec.CommandContext(ctx, w.taskAttrs.Cmd, w.taskAttrs.CmdArgs...)
  command_h.Dir = w.workingDir //TODO: append task relative directory w.taskAttrs.WorkDir
  command_h.Stdout = task_logger
  command_h.Stderr = task_logger
  command_h.Env = w.getTaskUserEnv(w.taskId)

  base.Debug("[WorkHorse.DoWork] current directory: ", w.workingDir)
  err = command_h.Start()
  
  if err != nil {
    base.Log("[WorkHorse.DoWork] Failed to start task command for flow-task: ", w.flowId, w.taskId, w.taskAttrs.Cmd)
    base.Log("[WorkHorse.DoWork] Error : ",  err)

    w.PushTaskStatusWithMessage(tsk.TASK_FAILED, err.Error())

    return err
  }

  progress, err := command_h.Process.Wait()
  if err != nil {
    base.Log("[WorkHorse.DoWork] Failed while waiting task command for flow-task: ", w.flowId, w.taskId, w.taskAttrs.Cmd)
    w.PushTaskStatusWithMessage(tsk.TASK_FAILED, err.Error())    
    return err
  }

  if ctx != nil {
    select {
      case <- ctx.Done():

        base.Debug("[WorkHorse.DoWork] Was waiting on process. Received Done")
        if err = ctx.Err(); err != nil {
          return err
        }
      default: 
        //do nothing
    }
  }

  // initiate i/o close 
  err = command_h.WaitForIOClose(progress, err) 
  if err != nil {
    base.Log("[WorkHorse.DoWork] Failed while Closing I/O channel: ", err) 
    return err
  }

  err = w.handleOutcome()
  if err != nil {
    w.PushTaskStatusWithMessage(tsk.TASK_FAILED, "Failed to upload/commit out directory to repo")
    return err
  }

  return  w.PushTaskStatus(tsk.TASK_COMPLETED)
}

func (w *WorkHorse) Shutdown() error {
  return w.detach()
}

func (w *WorkHorse) register() error {
  worker_attrs, err := w.wc.RegisterWorker(w.flowId, w.taskId, w.workerIp)
  if err != nil {
    return err
  }

  if worker_attrs!= nil {
    base.Debug("[worker.Register] Registered Worker Attributes: ", *worker_attrs)
  }
  w.attrs = worker_attrs 
  w.Id = worker_attrs.Worker.Id
  base.Debug("[worker.Register] Worker ID: ", w.Id)
  return err
}


func (w *WorkHorse) detach() error {
  err := w.wc.DetachWorker(w.flowId, w.taskId, w.Id)
  return err
}

// 1. register 
// 2. set worker attributes 
// 3. mount volume
// 4. change task status to running
// 5. Run task commands
// 6. Commit results to server
// 7. Update task status 


// TODO: 
// 1) mounts task level repo fs
// 2) at this point only one mount is allowed.
// 3) Add a repo tracker for mounted fs to destroy the mount after task completion
// 4) Add cancel option for longer running mounts

func (w *WorkHorse) MountVolumes() error {
  var mountErr error 
  
  if w.flowAttrs.FlowConfig == nil {
    base.Log("[WorkHorse.MountVolume] Failed to retrieve flow config.")
    
    if IgnoreMountError {
      base.Debug("[WorkHorse.MountVolume] Ignoring mount config as it doesnt exist for this flow. IgnoreMountError is set to : ", IgnoreMountError)
      return nil
    }

    return fmt.Errorf("[WorkHorse.MountVolume] Flow Config not available")
  } 

  mount_map := w.flowAttrs.FlowConfig.MountMap
  fmt.Println("task mount config: ", mount_map)

  if len(mount_map) == 0 {
    base.Log("[WorkHorse.MountVolume] No mount config found for this flow.")
    return nil
  }

  for mount_repo, m_config := range mount_map {
    repoFs:= fsys.NewRepoFs(w.workingDir, FileOpsConcurrency, mount_repo, m_config.CommitId, w.wc)
    mountErr = repoFs.Mount()
    
    if mountErr != nil {
      base.Log("[WorkHorse.Mount] Failed to mount Repo FS: ", mount_repo, mountErr)
      return mountErr
    }
    _, _ = repoFs.ListFiles("", false)
    w.repoFs = repoFs
  }
  
  return nil
}

func (w *WorkHorse) pullFlowAttrs() error {
  flow_attrs, err:= w.wc.FetchFlowAttrs(w.flowId)

  if err != nil || flow_attrs == nil {
    base.Log("[WorkHorse.setFlowAttrs] Failed to fetch flow attributes: ", w.flowId, err)
    return err
  }

  if flow_attrs == nil || flow_attrs.Tasks == nil {
    return fmt.Errorf("[WorkHorse.setFlowAttrs] This flow has no tasks. %s", w.flowId)
  }

  task_attrs, ok := flow_attrs.Tasks[w.taskId]
  if !ok {
    return fmt.Errorf("[WorkHorse.setFlowAttrs] Assigned task does not exist in this flow has no tasks. %s %s", w.flowId, w.taskId)
  }
  w.flowAttrs = flow_attrs
  w.taskAttrs = &task_attrs
  return nil
}

func (w *WorkHorse) PushTaskStatus(status tsk.TaskStatus) error {
  return w.PushTaskStatusWithMessage(status, "")
}

func (w *WorkHorse) PushTaskStatusWithMessage(status tsk.TaskStatus, msg string) error {
  if w.taskAttrs == nil {
    return fmt.Errorf("Task Attributes unavailable for this worker")
  }

  tsr := &flw.TaskStatusChangeRequest {
    Flow: w.flowAttrs.Flow, 
    Task: *w.taskAttrs.Task,
    TaskStatus: status,
    Message: msg, 
  } 
  base.Log("[WorkHorse.PushTaskStatusWithMessage] Worker ID: ", w.Id)
  tsr_resp, err := w.wc.UpdateTaskStatus(w.Id, w.taskId, tsr)
  
  if err != nil {
    return err 
  }

  if tsr_resp.FlowAttrs != nil {
    task_val := tsr_resp.FlowAttrs.Tasks[w.taskId]
    w.taskAttrs = &task_val
    w.flowAttrs = tsr_resp.FlowAttrs
  }

  return nil
} 

// TODO
func (w *WorkHorse) getTaskUserEnv(Id string) []string{
  return nil
}


func (w *WorkHorse) handleOutcome() error {
  odir_size, err := w.repoFs.PushOutputDir()
  if err != nil {
    base.Log("[WorkHorse.handleOutcome] Failed to commit output files to repo: ", err)
    return err
  }
  base.Log("[WorkHorse.handleOutcome] Uploaded out directory: ", odir_size)
  return nil
}



