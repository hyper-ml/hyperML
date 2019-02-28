package worker

import (
  "io"  
  "fmt"
  "time"
  "strconv"
  "strings"
  "bufio"
  "context"
  "path/filepath"
  "hyperflow.in/server/pkg/base"
  flw "hyperflow.in/server/pkg/flow"
  tsk "hyperflow.in/server/pkg/tasks"

  "hyperflow.in/worker/utils"  
  fsys "hyperflow.in/worker/file_system"
  "hyperflow.in/worker/pkg/exec"
  . "hyperflow.in/worker/api_client"

)

const (
  FileOpsConcurrency int = 5
  DefaultOutPath = "/out"
  DefaultModelPath = "/saved_models"
  DefaultSrcPath = "/src"

  DefaultWorkSpacePerm = 0755
  DefaultModelPerm = 0755
  DefaultOutPerm = 0755
  
  IgnoreEmptyTask bool = true
  IgnoreMountError bool = true

  DefaultWorkspaceDir = "workspace"

  CondaSpecName = "environment.yml"
  DefaultCondaEnv = "python"
  PipReqFileName = "requirements.txt"
)


type WorkHorse struct {
  Id string

  flowId string
  taskId string

  flowAttrs *flw.FlowAttrs
  taskAttrs *tsk.TaskAttrs
 
  outDir string
  modelDir string
  srcDir string
  workingDir string
  workerIp string

  srcRepo *fsys.RepoFs
  outFs *fsys.RepoFs
  modelFs *fsys.RepoFs

  tasks map[string]*tsk.TaskAttrs  
  logHandlr *utils.LogHandler

  shutd chan int

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

func discardIOError(err error) error {
  if strings.Contains(err.Error(), "short write") {
    fmt.Println("found short write in error", err)
    return nil
  }
  return err
}
  
 
//
func NewWorkHorse(client *WorkerClient, serverAddr string, flowId string, taskId string, workerIp string, appDir string, logHandlr *utils.LogHandler) *WorkHorse {
  
  wdir, err := createWorkspace(appDir)
  if err != nil {
    //TODO:
    panic(err)
  } 

  wc := client 
  if wc == nil {
    wc, _ = NewWorkerClient(serverAddr)
  }

  tasks := make(map[string]*tsk.TaskAttrs)  

  return &WorkHorse {
    flowId: flowId,
    taskId: taskId,
    workerIp: workerIp,
    workingDir: wdir, 

    wc: wc,
    tasks: tasks,
    logHandlr: logHandlr,
    shutd: make(chan int),
    attrs: &flw.WorkerAttrs{}, 
    started: time.Now(),
    ctx: context.Background(),
  }
}
 

func (w *WorkHorse) Init() error {
  
  if err := w.register(); err != nil {
    base.Error("[WorkHorse.Init] Failed to register worker for this flow-task: ", w.flowId, w.taskId, w.workerIp)
    base.Error("[WorkHorse.Init] Error: ", err)
    return err
  }

  if err := w.pullFlowAttrs(); err != nil {
    base.Error("[WorkHorse.Init] Failed to pull flow-task attributes: ", w.flowId, w.taskId, err)
    return err
  }

  if err := w.PushTaskStatus(tsk.POD_ASSIGNED); err != nil {
    base.Error("[WorkHorse.Init] Failed to update task status to accepted: ", w.flowId, w.taskId, err)
    return err
  }
  

  return nil
}

func (w *WorkHorse) InitEnvironment() error {

  if err := w.createDirs(); err != nil {
    base.Error("[WorkHorse.Init] Failed to create workspace directories: ", err)
    return err
  }

  if err := w.MountVolumes(); err != nil {
    base.Error("[WorkHorse.Init] Failed to mount repo volums: ", err)
    return err
  }

  if err := w.setModelRepo(); err != nil {
    base.Error("[WorkHorse.Init] Failed to initiate model repo")
    return err
  }
  return nil
}

func (w *WorkHorse) createDirs() error {
  base_path := w.workingDir

  src_path := filepath.Join(base_path, DefaultSrcPath)
  if err := utils.MkDirAll(src_path, DefaultWorkSpacePerm); err != nil {
    base.Error("[WorkHorse.createDirs] Failed to create source dir: ", err)
  }
 
  model_dir := filepath.Join(base_path, DefaultModelPath)
  if err := utils.MkDirAll(model_dir, DefaultModelPerm); err != nil {
    base.Error("[WorkHorse.createDirs] Failed to create model dir: ", err)
  }

  out_dir := filepath.Join(base_path, DefaultOutPath)
  if err := utils.MkDirAll(out_dir, DefaultOutPerm); err != nil {
    base.Error("[WorkHorse.createDirs] Failed to create out dir: ", err)
  }

  /* todo: Could enable this when debug info is ON
  files, err := ioutil.ReadDir(base_path)
  if err != nil {
    base.Error("[WorkHorse.createDirs] Error reading directory: ", base_path, err)
  } else {
    base.Println("Listing files from user directory: ", w.workingDir)
    for _, f := range files {
      base.Println(f.Name())
    }
  } */

  w.modelDir = model_dir
  w.outDir = out_dir
  w.srcDir = src_path

  return nil
}

func (w *WorkHorse) getCondaEnv() (string, error) {
  conda_spec := filepath.Join(w.srcDir, CondaSpecName)

  if utils.PathExists(conda_spec) {
    base.Println("Creating a new env with file :", conda_spec)
    return utils.GetOrCreateCondaEnvWithSpec(conda_spec)  
  } else {
    return "", nil
  }

  return "", fmt.Errorf("Unknown error")
} 

func (w *WorkHorse) getPipReqFile() string {
  req_spec := filepath.Join(w.srcDir, PipReqFileName)
  if utils.PathExists(req_spec) {
    return req_spec
  }

  return ""
}


func copyOutput(r io.Reader) {
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }
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
    base.Error("[WorkHorse.DoWork] The status of this task is not suitable for execution.", w.taskAttrs.Status)
    return InvalidFlowTaskStatus(w.taskAttrs.Status)
  }

  w.PushTaskStatus(tsk.TASK_RUNNING)
  ctx := w.ctx
  if w.taskAttrs.Cmd == "" {
    
    if IgnoreEmptyTask {
      base.Warn("[WorkHorse.DoWork] This task has no commands to execute. ", w.taskAttrs)
      return w.PushTaskStatus(tsk.TASK_COMPLETED)
    }

    w.PushTaskStatusWithMessage(tsk.TASK_FAILED, "The task is empty.")
    return emptyTaskError(w.flowId, w.taskId)
  }
  
  py_cmd := "cd " + w.srcDir + ";"

  env, err := w.getCondaEnv()
  if err != nil {
    base.Error("Failed to setup conda environemnt: ", err)  
    w.PushTaskStatusWithMessage(tsk.TASK_FAILED, err.Error())
    return err
  }

  if env != "" {
    py_cmd = py_cmd + "source activate " + env + ";"
  }

  if req_spec := w.getPipReqFile(); req_spec != "" {
    py_cmd = py_cmd + "pip " + req_spec + ";"
  }

  py_cmd = py_cmd + w.taskAttrs.Cmd + " " + strings.Join(w.taskAttrs.CmdArgs, " ") + ";"

  command_h := exec.CommandContext(ctx, "bash","-c", py_cmd)
  command_h.Dir = w.srcDir  
  //command_h.Stdout = w.logHandlr
  //command_h.Stderr = w.logHandlr
  command_h.Env = w.getTaskUserEnv(w.taskId) 
  
  stdout, _ := command_h.StdoutPipe()
  stderr, _ := command_h.StderrPipe()

  err = command_h.Start()

  fmt.Println("Command Initiated.")
  go copyOutput(stdout)
  go copyOutput(stderr)

  if err != nil {
    base.Error("[WorkHorse.DoWork] Failed to start task command for flow-task: ", w.flowId, w.taskId, w.taskAttrs.Cmd)
    base.Error("[WorkHorse.DoWork] Error : ",  err)

    w.PushTaskStatusWithMessage(tsk.TASK_FAILED, err.Error())
    return err
  }
  w.PushTaskStatus(tsk.CMD_STARTED)
 
  progress, err := command_h.Process.Wait()
  if err != nil {
    base.Error("[WorkHorse.DoWork] Failed while waiting task command for flow-task: ", w.flowId, w.taskId, w.taskAttrs.Cmd)
    w.PushTaskStatusWithMessage(tsk.TASK_FAILED, err.Error())    
    return err
  } 
 
  w.PushTaskStatus(tsk.CMD_COMPLETE)
 
  if ctx != nil {
    select {
      case <- ctx.Done():

        base.Println("Task received a cancel request")
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
    if discardIOError(err) != nil {
      base.Warn("[WorkHorse.DoWork] Command Error: ", err) 
      has_failures = true
      w.PushTaskStatusWithMessage(tsk.CMD_FAILED, "Task command line failed")
    }
  } 
  fmt.Print("** end of command log ** \n")

  err = w.pushLog() 
  if err != nil {
    base.Warn("[WorkHorse.DoWork] pushLog failed: ", err)
    has_failures = true
    w.PushTaskStatusWithMessage(tsk.LOG_UPLOAD_FAILED, "Failed to capture task log")
  }

  // push saved models and close commit 
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
  w.shutd <- 1
  return w.detach()
}

func (w *WorkHorse) register() error {
  worker_attrs, err := w.wc.RegisterWorker(w.flowId, w.taskId, w.workerIp)
  if err != nil {
    return err
  }

  if worker_attrs!= nil {
    // worker created  
  }

  w.attrs = worker_attrs 
  w.Id = worker_attrs.Worker.Id
  base.Println("Worker Assigned to the task")
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
  src_path := filepath.Join(w.workingDir, DefaultSrcPath)

  if w.flowAttrs.FlowConfig == nil {
    base.Error("[WorkHorse.MountVolume] Failed to retrieve flow config.")
    
    if IgnoreMountError {
      base.Debug("[WorkHorse.MountVolume] Ignoring mount config as it doesnt exist for this flow. IgnoreMountError is set to : ", IgnoreMountError)
      return nil
    }

    return fmt.Errorf("[WorkHorse.MountVolume] Flow Config not available")
  } 

  mount_map := w.flowAttrs.FlowConfig.MountMap

  if len(mount_map) == 0 {
    base.Error("[WorkHorse.MountVolume] No mount config found for this flow.")
    return nil
  }

  for mount_repo, m_config := range mount_map {

    srcRepo:= fsys.NewRepoFs(src_path, FileOpsConcurrency, m_config.RepoName, m_config.BranchName, m_config.CommitId, w.wc)
    mountErr = srcRepo.Mount()
    
    if mountErr != nil {
      base.Error("[WorkHorse.Mount] Failed to mount Repo FS: ", mount_repo, mountErr)
      return mountErr
    }
    w.srcRepo = srcRepo
  }
  
  return nil
}

func (w *WorkHorse) pullFlowAttrs() error {
  flow_attrs, err:= w.wc.FetchFlowAttrs(w.flowId)

  if err != nil || flow_attrs == nil {
    base.Error("[WorkHorse.setFlowAttrs] Failed to fetch flow attributes: ", w.flowId, err)
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
  
  if w.flowAttrs == nil {
    return utils.ErrMissingFlowAttrs()
  }

  flow_id := w.flowAttrs.Flow.Id
  repo, branch, commit, err := w.wc.GetOrCreateModelRepo(flow_id)
  if err != nil {
    base.Error("[WorkHorse.setModelRepo] Failed to get model repo for the task: ", flow_id, err)
    return err
  }

  model_path := w.modelDir
  if model_path == "" {
    model_path = filepath.Join(w.workingDir, DefaultModelPath)
  }
    
  fs := fsys.NewRepoFs(model_path, FileOpsConcurrency, repo.Name, branch.Name, commit.Id, w.wc)  
  w.modelFs = fs
  
  go fsys.SyncRepo(fs, w.shutd)

  return nil
}

func (w *WorkHorse) pushSavedModels() error {

  dir_size, err := w.modelFs.PushModelDir()
  if err != nil {
    base.Error("[WorkHorse.pushSavedModels] Failed to commit output files to repo: ", err)
    return err
  }
  
  if err := w.modelFs.CloseCommit(); err != nil {
    base.Error("[WorkHorse.pushSavedModels] Failed to close model commit")
    return err
  }
  base.Println("Model checked into repo (size in bytes): " +  w.modelFs.GetRepoId() + "(" + strconv.FormatInt(dir_size, 10) +")")
  
  return nil
}

func (w *WorkHorse) setOutRepo() error {
  repo, branch, commit, err := w.wc.GetOrCreateFlowOutRepo(w.flowId)
  if err != nil {
    return err
  }
  
  out_path := w.outDir
  if out_path == "" {
    out_path = filepath.Join(w.workingDir, DefaultOutPath)
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
    base.Error("[WorkHorse.pushOutput] Failed to commit output files to repo: ", err)
    return err
  }
  
  base.Println("Uploaded out directory: ", out_size)
  
  if err := w.outFs.CloseCommit(); err != nil {
    base.Error("[WorkHorse.pushOutput] Failed to close output commit")
    return err
  }

  return nil
}


func (w *WorkHorse) closeLogFiles() error {
  w.logHandlr.Close()
  return nil 
}


func (w *WorkHorse) pushLog() error {
  var upload int64
  logger := w.logHandlr
  
  if logger == nil {
    base.Warn("[WorkHorse.pushLog] Logger not found for flowId:  ", w.taskId)
    return nil
  }

  err := w.closeLogFiles()
  log_path:= logger.GetLogPath()

  log_io, err := utils.Open(log_path) 
  if err != nil {
    base.Error("[WorkHorse.pushLog] Failed to open log file: ", log_path)
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
 
  base.Println("Pushed log to server. Size (in bytes): ", upload)
  
  return nil
}

// utility routines
func createWorkspace(path string) (wspath string, err error) {

  wspath = path
  if wspath == "" {
    wspath, err = base.HomeDir()
    if err != nil {
      return wspath, err
    }
  }

  wspath = filepath.Join(wspath, DefaultWorkspaceDir)
   
  if err := utils.MkDirAll(wspath, DefaultWorkSpacePerm); err != nil {
    base.Error("[WorkHorse.createWorkspace] Failed to create workspace dir: ", err)
    return wspath, err
  }

  return
}

func cleanupWorkspace(path string) error {
  //TODO: 
  return fmt.Errorf("unimplemnted feature")
}

