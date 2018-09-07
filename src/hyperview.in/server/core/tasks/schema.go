package tasks

import (
  "time"
  "strings"
  "hyperview.in/server/core/utils"
  
)
 
const (
  MOUNT_TYPE_DEFAULT int = iota 
  MOUNT_TYPE_PIPE
  MOUNT_TYPE_FUSE
)

// details of mount file system
// this is usually tied to a repo or commit key
type MountConfig struct {

  // commit ID for file map
  CommitId string

   // target path for mount
  Target string 

  // PIPE, FUSE or DEFAULT
  MountType int 

}

type MountMap map[string]MountConfig

func NewMountConfig(repoName string, commitId string, targetPath string, mountType int) MountMap{
  m_type := MOUNT_TYPE_DEFAULT
  if mountType != 0 {
    m_type = mountType
  } 
  
  var mmap map[string]MountConfig
  mmap = make(map[string]MountConfig)
  mmap[repoName] = MountConfig{Target: targetPath, MountType: m_type, CommitId: commitId} 
  return mmap
}

type TaskConfig struct {
  // repo to mount path mapping
  MountMap MountMap

  Cmd string
  CmdArgs []string

  WorkDir string   
}



func NewTaskConfig(cmd string, cargs []string, wdir string, mmap MountMap) *TaskConfig{
  var head string
  var cmd_args []string

  if len(cargs) == 0 {
    cmd_parts := strings.Fields(cmd)
    head = cmd_parts[0]
    cmd_args = cmd_parts[1:len(cmd_parts)]
  } else {
    head = cmd
    cmd_args = cargs
  }

  return &TaskConfig{
    MountMap: mmap,
    Cmd: head,
    CmdArgs: cmd_args,
    WorkDir: wdir,
  }
  
}

type TaskStatus int
const (
  TASK_CREATED TaskStatus = iota
  TASK_WAITINGTOSTART
  TASK_STARTING 
  TASK_ASSIGNED
  TASK_ACCEPTED
  TASK_INITIATED
  TASK_CONFIGURED
  TASK_WAITING
  TASK_FAILED
  TASK_STOPPING
  TASK_STOPPED
  TASK_RUNNING 
  TASK_COMPLETING
  TASK_COMPLETED
  TASK_CANCELLING
  TASK_CANCELLED
)

var TaskStatusKey = map[int]string {
  0: "TASK_CREATED",
  1: "TASK_WAITINGTOSTART",
  2: "TASK_STARTING",
  3: "TASK_ASSIGNED",
  4: "TASK_ACCEPTED",
  5: "TASK_INITIATED",
  6: "TASK_CONFIGURED",
  7: "TASK_WAITING",
  8: "TASK_FAILED",
  9: "TASK_STOPPING",
  10: "TASK_STOPPED",
  11: "TASK_RUNNING",
  12: "TASK_COMPLETING",
  13: "TASK_COMPLETED",
  14: "TASK_CANCELLING",
  15: "TASK_CANCELLED",
}

func TaskStatusByKey(key TaskStatus) string {
  int_val := int(key)
  return TaskStatusKey[int_val]
}

type Task struct {
  Id string
}


type TaskAttrs struct {
  Task *Task
  // map of repo to target mounted paths
  // used during unmounting
  OpenMounts map[string]string

  // add time   
  WorkDir string
  Cmd string
  CmdArgs []string 

  Status TaskStatus
  Created time.Time
    
  Started time.Time
  Completed time.Time

  Failed time.Time
  FailureReason string 

  TaskConfig *TaskConfig

  WorkerPref map[string]string

}

func NewTaskKey() string {
  return utils.NewUUID()
}


func NewTaskAttrs(tc *TaskConfig) *TaskAttrs {

  return &TaskAttrs {
    Task: &Task {
      Id: NewTaskKey(),
    },
    WorkDir: tc.WorkDir,
    Cmd: tc.Cmd,
    CmdArgs: tc.CmdArgs, 
    TaskConfig: tc,
    Status: TASK_CREATED,
  }
}

func (taskAttrs *TaskAttrs) GetTaskConfig() (*TaskConfig) {
  return taskAttrs.TaskConfig
}
