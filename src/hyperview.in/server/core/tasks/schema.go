package tasks

import (
  "time"
  "strings"
  "hyperview.in/server/core/utils"
    ws "hyperview.in/server/core/workspace"

)
 
const (
  MOUNT_TYPE_DEFAULT int = iota 
  MOUNT_TYPE_PIPE
  MOUNT_TYPE_FUSE
)

// details of mount file system
// this is usually tied to a repo or commit key
type MountConfig struct {
  //repo
  RepoName string

  //repo_type
  RepoType ws.RepoType

  //branch 
  BranchName string

  // commit ID for file map
  CommitId string

   // target path for mount
  Target string 

  // PIPE, FUSE or DEFAULT
  MountType int 

}

type MountMap map[string]MountConfig

func NewMountConfig(repoName string, branchName string, commitId string, targetPath string, mountType int) MountMap{
  m_type := MOUNT_TYPE_DEFAULT
  if mountType != 0 {
    m_type = mountType
  } 
  
  var mmap map[string]MountConfig
  mmap = make(map[string]MountConfig)
  mmap[repoName] = MountConfig{RepoType: ws.STANDARD_REPO, Target: targetPath, MountType: m_type, RepoName: repoName, BranchName: branchName, CommitId: commitId} 
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
  TASK_CREATED TaskStatus = 0
  TASK_WAITINGTOSTART TaskStatus = 10
  TASK_STARTING TaskStatus = 20
  TASK_ASSIGNED TaskStatus = 30
  TASK_ACCEPTED TaskStatus = 40
  TASK_INITIATED TaskStatus = 50
  TASK_CONFIGURED TaskStatus = 60
  TASK_WAITING TaskStatus = 70
  TASK_RUNNING TaskStatus = 80
   

  CMD_COMPLETE TaskStatus = 90
  CMD_FAILED  TaskStatus = 100
  CMD_WARNING  TaskStatus = 110
  LOG_UPLOAD_FAILED  TaskStatus = 120
  LOG_UPLOAD_COMPLETE  TaskStatus = 130
  MODEL_UPLOAD_COMPLETE TaskStatus = 140
  MODEL_UPLOAD_FAILED TaskStatus = 150
  OUTPUT_UPLOAD_COMPLETE TaskStatus = 160
  OUTPUT_UPLOAD_FAILED TaskStatus = 170
  TASK_COMPLETING TaskStatus = 180
  TASK_STOPPING TaskStatus = 190
  TASK_CANCELLING TaskStatus = 200
  TASK_FAILED TaskStatus = 300
  TASK_STOPPED TaskStatus = 310
  TASK_COMPLETED TaskStatus = 320
  TASK_WARNING TaskStatus = 330
  TASK_CANCELLED TaskStatus = 340
)

var TaskStatusKey = map[int]string {
   0: "TASK_CREATED",
  10: "TASK_WAITINGTOSTART",
  20: "TASK_STARTING",
  30: "TASK_ASSIGNED" ,
  40: "TASK_ACCEPTED" ,
  50: "TASK_INITIATED"  ,
  60: "TASK_CONFIGURED" ,
  70: "TASK_WAITING"  ,
  80: "TASK_RUNNING"  ,

  90: "CMD_COMPLETE",
  100: "CMD_FAILED",
  110: "CMD_WARNING",
  
  120: "LOG_UPLOAD_FAILED",
  130: "LOG_UPLOAD_COMPLETE",
  
  140: "MODEL_UPLOAD_COMPLETE",
  150: "MODEL_UPLOAD_FAILED",
  
  160: "OUTPUT_UPLOAD_COMPLETE",
  170: "OUTPUT_UPLOAD_FAILED",
  
  180: "TASK_COMPLETING",
  190: "TASK_STOPPING",
  200: "TASK_CANCELLING",
  300: "TASK_FAILED",
  310: "TASK_STOPPED",
  320: "TASK_COMPLETED",
  330: "TASK_WARNING",
  340: "TASK_CANCELLED",
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

func (t *TaskAttrs) GetTaskConfig() (*TaskConfig) {
  return t.TaskConfig
}

func (t *TaskAttrs) MasterRepo() (repoName, branchName, commitId string) {
  if t.TaskConfig == nil {
    return 
  }

  for _, mount_config := range t.TaskConfig.MountMap {
    if mount_config.RepoType == ws.STANDARD_REPO {
      repoName = mount_config.RepoName
      branchName = mount_config.BranchName
      commitId = mount_config.CommitId
      return
    }
  }

  return
}





