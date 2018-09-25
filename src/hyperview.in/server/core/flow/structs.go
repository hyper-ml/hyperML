package flow

import (
  "time"
  "sync"
  . "hyperview.in/server/core/tasks"
  "hyperview.in/server/base"
 
)
 

type FlowStatus int

const (
  FLOW_CREATED FlowStatus = iota
  FLOW_WAITINGTOSTART
  FLOW_STARTING 
  FLOW_STARTED
  FLOW_WAITING
  FLOW_RUNNING
  FLOW_COMPLETING
  FLOW_CANCELLING  
  FLOW_STOPPING
  FLOW_FAILED
  FLOW_STOPPED
  FLOW_COMPLETED
  FLOW_CANCELLED
)


var FlowStatusKey = map[int]string {
   0: "FLOW_CREATED",
   1: "FLOW_WAITINGTOSTART",
   2: "FLOW_STARTING",
   3: "FLOW_STARTED",
   4: "FLOW_WAITING",
   5: "FLOW_RUNNING",
   6: "FLOW_COMPLETING",
   7: "FLOW_CANCELLING",
   8: "FLOW_STOPPING",
   9: "FLOW_FAILED",
   10: "FLOW_STOPPED",
   11: "FLOW_COMPLETED",
   12: "FLOW_CANCELLED",
}


func FlowStatusToString(key FlowStatus) string {
  int_val := int(key)
  return FlowStatusKey[int_val]
}
  
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

func FlowRef(id string) Flow {
  return Flow {Id: id}
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
    Status: FLOW_CREATED,
    Created: time.Now(),
    Tasks: make(map[string]TaskAttrs),
  }
} 

/* temporary function to retrieve master repo 
 * Need a better way to track master repo 
 */

func (f *FlowAttrs) masterRepo() (repoName, branchName, commitId string) {
  if task_attrs := f.FirstTask(); task_attrs != nil {
    return task_attrs.MasterRepo()
  }
  return 
}

func (f *FlowAttrs) isComplete() bool {

  if f.Status >= FLOW_FAILED {
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
  if fi.Status == FLOW_CREATED {
    return true
  }
  return false
}


func (fi *FlowAttrs) IsStarted() bool {
  if fi.Status == FLOW_STARTED {
    return true
  }
  return false
}


func (fi *FlowAttrs) IsFailed() bool {
  if fi.Status == FLOW_FAILED {
    return true
  }
  return false
}


func (fi *FlowAttrs) IsStarting() bool {
  if fi.Status == FLOW_STARTING {
    return true
  }
  return false
}


func (fi *FlowAttrs) IsCompleted() bool {
  if fi.Status == FLOW_COMPLETED {
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


