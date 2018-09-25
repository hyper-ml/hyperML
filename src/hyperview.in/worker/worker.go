package worker

import (
  "fmt"
  "time"
  "context"
  "io"  
  "io/ioutil"
  "path/filepath"
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
  DefaultOutPath = "/out"
  DefaultModelPath = "/saved_models"
  UserVolume = "/wh_data" 

  DefaultWorkSpacePerm = 0755
  DefaultModelPerm = 0755
  DefaultOutPerm = 0755
  
  IgnoreEmptyTask bool = true
  IgnoreMountError bool = true
  DefaultHomeDir = "/wh_data"

)


type WorkHorse struct {
  Id string

  flowId string
  taskId string

  flowAttrs *flw.FlowAttrs
  taskAttrs *tsk.TaskAttrs

  homeDir string 
  outDir string
  modelDir string
  workingDir string
  workerIp string

  repoFs *fsys.RepoFs
  outFs *fsys.RepoFs
  modelFs *fsys.RepoFs

  tasks map[string]*tsk.TaskAttrs
  tlogs map[string]utils.LoggerInterface
  
  // TODO: add task environments

  started time.Time
  wc *WorkerClient
  attrs *flw.WorkerAttrs
  ctx context.Context
}

func emptyTaskError(flowId, taskId string) error {
  return fmt.Errorf("empty_task: The task is empty: %s %s", flowId, taskId)
}

func InvalidFlowTaskError() error {
  return fmt.Errorf("empty_task_flow_ids: Empty Flow or task assigned to the worker")
}

func InvalidFlowTaskStatus(status tsk.TaskStatus) error {
  return fmt.Errorf("invalid_status: Invalid Status (%d) for running the task. Check docs for valid statuses. ", int(status))
}

func errTaskHasFailures() error {
  return fmt.Errorf("task_has_failures: This task has failures.")
}

//
func NewWorkHorse(client *WorkerClient, serverAddr string, homeDir string, flowId string, taskId string, workerIp string, workingDir string) *WorkHorse {
  wc := client
  home_dir := homeDir
  if home_dir == "" {
    home_dir = DefaultHomeDir
  }

  if wc == nil {
    wc, _ = NewWorkerClient(serverAddr)
  }

  tasks := make(map[string]*tsk.TaskAttrs) 
  tlogs := make(map[string]utils.LoggerInterface) 
  
  if taskId != "" { 
    tlogs[taskId]  = utils.NewTaskLogger(home_dir, taskId)
  }  

  return &WorkHorse {
    flowId: flowId,
    taskId: taskId,
    workerIp: workerIp,
    workingDir: workingDir,
    homeDir: home_dir,

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
  
  if err := w.createDirs(); err != nil {
    base.Error("[WorkHorse.Init] Failed to create workspace directories: ", err)
    return err
  }

  if err := w.MountVolumes(); err != nil {
    base.Log("[WorkHorse.Init] Failed to mount repo volums: ", err)
    return err
  }

  return nil
}

func (w *WorkHorse) createDirs() error {
  
  if err := utils.MkDirAll(w.workingDir, DefaultWorkSpacePerm); err != nil {
    base.Error("[WorkHorse.createDirs] Failed to create workspace dir: ", err)
  }

  model_dir := filepath.Join(UserVolume, DefaultModelPath)
  if err := utils.MkDirAll(model_dir, DefaultModelPerm); err != nil {
    base.Error("[WorkHorse.createDirs] Failed to create model dir: ", err)
  }

  out_dir := filepath.Join(UserVolume, DefaultOutPath)
  if err := utils.MkDirAll(out_dir, DefaultOutPerm); err != nil {
    base.Error("[WorkHorse.createDirs] Failed to create out dir: ", err)
  }

  files, err := ioutil.ReadDir(UserVolume)
  if err != nil {
    base.Error("[WorkHorse.createDirs] Error reading directory: ", UserVolume, err)
  } else {
    base.Info("Listing files from user directory: ", UserVolume)
    for _, f := range files {
      base.Info(f.Name())
    }
  } 

  w.modelDir = model_dir
  w.outDir = out_dir

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
  var has_failures bool

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
  base.Debug("[WorkHorse.DoWork] Command string: ", w.taskAttrs.Cmd)
  base.Debug("[WorkHorse.DoWork] Command Args: ", w.taskAttrs.CmdArgs)

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
          w.PushTaskStatusWithMessage(tsk.TASK_STOPPED, err.Error())  
          return err
        }
      default: 
        //do nothing
    }
  }

  // initiate i/o close 
  err = command_h.WaitForIOClose(progress, err) 
  if err != nil {
    base.Warn("[WorkHorse.DoWork] Command Error: ", err) 
    has_failures = true
    w.PushTaskStatusWithMessage(tsk.CMD_FAILED, "Task command line failed")
    // continue capturing log and output files
  }
  
  err = w.pushLog() 
  if err != nil {
    base.Warn("[WorkHorse.DoWork] pushLog failed: ", err)
    has_failures = true
    w.PushTaskStatusWithMessage(tsk.LOG_UPLOAD_FAILED, "Failed to capture task log")
  }

  err = w.pushSavedModels()
  if err != nil {
    base.Warn("[WorkHorse.DoWork] pushSavedModels failed: ", err, tsk.TASK_FAILED)
    has_failures = true
    w.PushTaskStatusWithMessage(tsk.MODEL_UPLOAD_FAILED, "Failed to upload/commit saved_models directory to repo")
  }

  err = w.pushOutput()
  if err != nil {
    base.Warn("[WorkHorse.DoWork] pushOutput failed: ", err)
    has_failures = true
    w.PushTaskStatusWithMessage(tsk.OUTPUT_UPLOAD_FAILED, "Failed to upload/commit out directory to repo")
  }
  
  if has_failures  {
    w.PushTaskStatusWithMessage(tsk.TASK_WARNING, "Failures during execution. Check log for details.")
    if err != nil {
      return err 
    }  
    return errTaskHasFailures()
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
    repoFs:= fsys.NewRepoFs(w.workingDir, FileOpsConcurrency, m_config.RepoName, m_config.BranchName, m_config.CommitId, w.wc)
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
  base.Debug("[WorkHorse.PushTaskStatusWithMessage] Updating task Status: ", status)
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
 
func (w *WorkHorse) getRepoFromFlowConfig() (string, string, string, error){
  if w.flowAttrs == nil {
    return "", "", "", fmt.Errorf("Repo Information is unavailble to generate model repo")
  }

  mount_map := w.flowAttrs.FlowConfig.MountMap
  if len(mount_map) == 0 {
    return "", "", "", nil
  }

  if len(mount_map) > 1 {
    return "", "", "", fmt.Errorf("[WorkHorse.getRepoFromFlowConfig] was expecting just one repo associated with this task.")
  }

  // expecting just one 
  for _, mount_config := range mount_map {
    return mount_config.RepoName, mount_config.BranchName, mount_config.CommitId, nil
  }
  return "","","", nil
}

func (w *WorkHorse) setModelRepo() error {
  source_repo, source_branch, source_commit, err := w.getRepoFromFlowConfig()
  if err != nil {
    base.Warn("[WorkHorse.setModelRepo] Failed to retrieve master repo from flow config: ", err)
    return err 
  }

  repo, branch, commit, err := w.wc.GetOrCreateModelRepo(source_repo, source_branch, source_commit)
  if err != nil {
    base.Warn("[WorkHorse.setModelRepo] Failed to get model repo for master repo/commit: ", source_repo, source_commit, err)
    return err
  }

  model_path := w.modelDir
  if model_path == "" {
    model_path = filepath.Join(UserVolume, DefaultModelPath)
  }

  base.Log("Model Directory: ", model_path)

  files, err := ioutil.ReadDir(model_path)
  if err != nil {
    base.Error("[WorkHorse.setModelRepo] Error reading directory: ", model_path, err)
  } else {
    base.Info("[WorkHorse.setModelRepo] Listing files from workspace directory: ", model_path)
    for _, f := range files {
      base.Info(f.Name())
    }
  } 

  fs := fsys.NewRepoFs(model_path, FileOpsConcurrency, repo.Name, branch.Name, commit.Id, w.wc)  
  w.modelFs = fs

  return nil
}

func (w *WorkHorse) pushSavedModels() error {

  if err := w.setModelRepo(); err!= nil {
    base.Error("[WorkHorse.pushSavedModels] Could not set model repo: ", err)
    return err
  } 

  dir_size, err := w.modelFs.PushModelDir()
  if err != nil {
    base.Error("[WorkHorse.pushSavedModels] Failed to commit output files to repo: ", err)
    return err
  }
  
  if err := w.modelFs.CloseCommit(); err != nil {
    base.Error("[WorkHorse.pushSavedModels] Failed to close model commit")
    return err
  }
  base.Log("[WorkHorse.pushSavedModels] saved_models updated (size in bytes): ", dir_size)
  
  return nil
}

func (w *WorkHorse) setOutRepo() error {
  repo, branch, commit, err := w.wc.GetOrCreateFlowOutRepo(w.flowId)
  if err != nil {
    return err
  }
  
  out_path := w.outDir
  if out_path == "" {
    out_path = filepath.Join(UserVolume, DefaultOutPath)
  }

  out_repo := fsys.NewRepoFs(out_path, FileOpsConcurrency, repo.Name, branch.Name, commit.Id, w.wc)  
  w.outFs = out_repo
  return nil
}

func (w *WorkHorse) pushOutput() error {
  // create out fs and then push output dir?? 
  if err := w.setOutRepo(); err!= nil {
    return err
  } 

  out_size, err := w.outFs.PushOutputDir()
  if err != nil {
    base.Log("[WorkHorse.pushOutput] Failed to commit output files to repo: ", err)
    return err
  }
  
  base.Log("[WorkHorse.pushOutput] Uploaded out directory: ", out_size)
  
  if err := w.outFs.CloseCommit(); err != nil {
    base.Error("[WorkHorse.pushOutput] Failed to close output commit")
    return err
  }

  return nil
}


func (w *WorkHorse) closeLogFiles() error {
 
 for _, logger := range w.tlogs {
  logger.Close()
 }

 return nil 
}

func (w *WorkHorse) pushLog() error {
  var upload int64
  logger:= w.tlogs[w.taskId]
  
  if logger == nil {
    base.Log("[WorkHorse.pushLog] Logger not found for flowId:  ", w.taskId, w.tlogs)
    return nil
  }

  err := w.closeLogFiles()
  log_path:= logger.GetLogPath()

  log_io, err := utils.Open(log_path) 
  if err != nil {
    base.Log("[WorkHorse.pushLog] Failed to open log file: ", log_path)
    return err 
  }

  writer, err := w.wc.PostLogWriter(w.taskId)
  if err != nil {
    base.Warn("[WorkHorse.pushLog] Failed to get log http writer: ", err)
    return err
  }

  defer writer.Close()

  buf := utils.GetBuffer()
  defer utils.PutBuffer(buf) 
  for {
        
      rbuf, err := log_io.Read(buf)
 

      if rbuf == 0 && err != nil {
        if err == io.EOF {
          return nil
        }
        return  err
      }
       
      wbuf, err := writer.Write(buf[:rbuf])
      if err != nil {
        return err
      }

      upload = upload + int64(wbuf)  
  }
 
  base.Debug("[WorkHorse.pushLog] log size uploaded: ", upload)
  
  return nil
}


