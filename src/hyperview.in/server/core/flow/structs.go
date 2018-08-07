package flow

import (
  "time"
  "sync"
  . "hyperview.in/server/core/tasks"
  
)

type FlowStatus int

const (
  CREATED FlowStatus = iota
  WAITINGTOSTART
  STARTING 
  STARTED
  WAITING
  RUNNING
  FAILED
  STOPPED
  STOPPING
  COMPLETING
  COMPLETED
  CANCELLING
  CANCELLED
)

type WorkerStatus int 

const (
  WORKER_REGISTERED WorkerStatus = iota
  WORKER_RUNNING
  WORKER_FAILED
  WORKER_COMPLETED
)

type FlowConfig struct {
  MountMap MountMap
}

type Flow struct {
  Id string
  Version string
}

// TODO: Add version to flow at somepoint
type FlowAttrs struct {
  Flow Flow


  // mounted file systems
  OpenMounts map[string]string

  // value - look at constants above 
  Status FlowStatus 

  Created time.Time
  Started time.Time
  Completed time.Time
  Failed time.Time

  // support multiple tasks in future releases
  Tasks map[string]TaskAttrs `json:"Tasks"`

  MountConfig MountConfig `json:"MountConfig"`

  // lock during worker assignment
  waLock sync.RWMutex
}

type Worker struct {
  Id string
}

type WorkerAttrs struct {
  Worker Worker `json:"Worker"`
  Flow Flow `json:"Flow"`
  Ip string `json:"ip"`
  Task Task `json:"Task"`
  Started time.Time `json:"started"`
  Completed time.Time `json:"completed"`
  Error string `json:"error"`
  Status WorkerStatus `json:"WorkerStatus"` // REGISTERED, RUNNING, FAILED, STOPPED  
}

type FlowTaskWorker struct {
  Worker Worker `json:"Worker"`
  Flow Flow `json:"Flow"`
  Task Task `json:"Task"`
  Created time.Time `json:"created"`
}


func NewFlowAttrs(fc *FlowConfig) *FlowAttrs {
  new_flow_id := NewTaskKey()

  return &FlowAttrs {
    Flow: Flow {
      Id: new_flow_id,
    },
    Status: CREATED,
    Created: time.Now(),
    Tasks: make(map[string]TaskAttrs),
  }
}

func (f *FlowAttrs) AddTask(taskConfig *TaskConfig) {
  new_task_id := NewTaskKey()
  new_task := TaskAttrs {
    Task: &Task {
      Id: new_task_id,
    },
    WorkDir: taskConfig.WorkDir,
    Cmd: taskConfig.Cmd,
    CmdArgs: taskConfig.CmdArgs,    
  } 
  f.Tasks[new_task_id] = new_task 
} 


func (fi *FlowAttrs) IsCreated() bool {
  if fi.Status == CREATED {
    return true
  }
  return false
}


func (fi *FlowAttrs) IsStarted() bool {
  if fi.Status == STARTED {
    return true
  }
  return false
}


func (fi *FlowAttrs) IsFailed() bool {
  if fi.Status == FAILED {
    return true
  }
  return false
}


func (fi *FlowAttrs) IsStarting() bool {
  if fi.Status == STARTING {
    return true
  }
  return false
}


func (fi *FlowAttrs) IsCompleted() bool {
  if fi.Status == COMPLETED {
    return true
  }
  return false
}


