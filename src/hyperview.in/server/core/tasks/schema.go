package tasks

import (
  "time"
  "hyperview.in/server/core/utils"
  ws "hyperview.in/server/core/workspace"
)

// key can be repo or repo:commit 
// value is
type mountConfig struct {
  target string // target path for mount
  type string // PIPE, FUSE or DEFAULT

}

type TaskConfig struct {
  mounts map[string]*mountContig

  cmd string
  cmdArgs []string

  workDir string   
}

type Task struct {
  Id string
}

type TaskInfo struct {
  Task *Task
  // map of repo to target mounted paths
  // used during unmounting
  OpenMounts map[string]string

  // add time   
  WorkDir string
  Cmd string
  CmdArgs []string 
    
  Started time.Time
  Completed time.Time

  Failed time.Time
  FailureReason string 

  taskConfig *TaskConfig

}

func NewTaskKey() string {
  return utils.NewUUID()
}


func NewTaskInfo(tc *TaskConfig) *TaskInfo {

  return &TaskInfo {
    Task: &Task {
      Id: NewTaskKey(),
    },
    WorkDir: tc.workDir
    Cmd: tc.cmd,
    CmdArgs: tc.cmdArgs,
    WorkDir: tc.workDir,
    taskConfig: tc,
  }
}

func (taskInfo *TaskInfo) GetTaskConfig *Taskconfig {
  return taskInfo.taskConfig
}
