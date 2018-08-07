package tasks

import (
  "time"
  "hyperview.in/server/core/utils"
  
)

// details of mount file system
// this is usually tied to a repo or commit key
type MountConfig struct {
   // target path for mount
  target string

  // PIPE, FUSE or DEFAULT
  mountType string 

}

type MountMap map[string]MountConfig


type TaskConfig struct {
  // repo to mount path mapping
  MountMap MountMap

  Cmd string
  CmdArgs []string

  WorkDir string   
}


func NewTaskConfig(cmd string, cargs []string, wdir string, mmap MountMap) *TaskConfig{

  return &TaskConfig{
    MountMap: mmap,
    Cmd: cmd,
    CmdArgs: cargs,
    WorkDir: wdir,
  }
  
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
    
  Started time.Time
  Completed time.Time

  Failed time.Time
  FailureReason string 

  taskConfig *TaskConfig

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
    taskConfig: tc,
  }
}

func (taskAttrs *TaskAttrs) GetTaskConfig() (*TaskConfig) {
  return taskAttrs.taskConfig
}
