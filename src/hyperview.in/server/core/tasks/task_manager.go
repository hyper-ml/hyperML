package tasks

import (
  "time"
  "fmt"

  db_pkg "hyperview.in/server/core/db"
)

type Tasker Interface {
  CreateTask(config *TaskConfig) error
  GetTask(Id string) (*TaskInfo, error)
}

type shellTasker struct {
  db *db_pkg.DatabaseContext
  started time.Time
  idle int64
  //keep track of launched tasks 
}

func NewShellTasker(db *db_pkg.DatabaseContext) *Tasker {
  return &shellTasker {
    db: db,
    started: time.Now()
  }
}


func (t *shellTasker) taskKey(taskId string) string {
  return "task:" + taskId
}

func (t *shellTasker) CreateTask(config *TaskConfig) error{
  task_info := NewTaskInfo(config)
  return q.db.Insert(task_info.Task.Id, task_info)
}

func (t *shellTasker) GetTask(Id string) (*TaskInfo, error) {
  var err error
  if Id == nil {
    return nil, fmt.Errorf("Task Id is required parameter to GetTaskInfo")
  }
  data, err := q.db.Get(taskKey(Id))
  
  taskInfo :=  &TaskInfo{} 
  err = json.Unmarshal(data, &taskInfo)
  
  return taskInfo, err
}

