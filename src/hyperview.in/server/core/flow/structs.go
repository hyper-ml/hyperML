package flow

import (
  "time"
  "sync"
  . "hyperview.in/server/core/tasks"
  "hyperview.in/server/base"
)

type FlowStatus int

const (
  CREATED FlowStatus = iota
  WAITINGTOSTART
  STARTING 
  STARTED
  WAITING
  RUNNING
  COMPLETING
  CANCELLING  
  STOPPING
  FAILED
  STOPPED
  COMPLETED
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

  FlowConfig *FlowConfig `json:"FlowConfig"`

  // lock during worker assignment
  waLock sync.RWMutex
}



func NewFlowAttrs(fc *FlowConfig) *FlowAttrs {
  new_flow_id := NewTaskKey()
  flow_config := copyConfig(fc)

  return &FlowAttrs {
    FlowConfig: flow_config,
    Flow: Flow {
      Id: new_flow_id,
    },
    Status: CREATED,
    Created: time.Now(),
    Tasks: make(map[string]TaskAttrs),
  }
}

func (f *FlowAttrs) isComplete() bool {

  if f.Status >= FAILED {
    return true
  }

  if len(f.Tasks) > 0 {
    for _, t := range f.Tasks {
      base.Debug("[FlowAttrs.isComplete] task status: ", t.Status)
      if t.Status >= TASK_FAILED {
        return true
      } 
    }
  } 

  return false 
}

func (f *FlowAttrs) AddTask(taskConfig *TaskConfig) *TaskAttrs{
  new_task_id := f.Flow.Id

  if len(f.Tasks) > 0 {
    new_task_id = NewTaskKey() 
  } 
  
  new_task := TaskAttrs {
    Task: &Task {
      Id: new_task_id,
    },
    WorkDir: taskConfig.WorkDir,
    Cmd: taskConfig.Cmd,
    CmdArgs: taskConfig.CmdArgs,   
    TaskConfig: taskConfig,  // TODO: re-think
  } 
  
  f.Tasks[new_task_id] = new_task 
  return &new_task
} 

func (f *FlowAttrs) FirstTask() *TaskAttrs {
  if len(f.Tasks) >0 {
    for _, v := range f.Tasks {
      return &v
    }
  }

  return nil
}

func copyConfig(fc *FlowConfig) *FlowConfig {
  return &FlowConfig {
    MountMap: fc.MountMap,
  }
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


type Worker struct {
  Id string
  PodId string
  PodPhase string
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


