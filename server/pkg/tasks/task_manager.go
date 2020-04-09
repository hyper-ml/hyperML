package tasks
/* -- commented as this is UNUSED -- 
import (
  "time"
  "fmt"
  "github.com/hyper-ml/hyperml/server/pkg/base"
  "encoding/json"
  db_pkg "github.com/hyper-ml/hyperml/server/pkg/db"
)

type Tasker interface {
  CreateTask(config *TaskConfig) (*TaskAttrs, error) 
  GetTaskAttrs(Id string) (*TaskAttrs, error)
}

type shellTasker struct {
  db *db_pkg.DatabaseContext
  started time.Time
  idle int64
  //keep track of launched tasks 
}

func NewShellTasker(db *db_pkg.DatabaseContext) Tasker {
  return &shellTasker {
    db: db,
    started: time.Now(),
  }
}


func taskKey(taskId string) string {
  return "task:" + taskId
}

func (t *shellTasker) CreateTask(config *TaskConfig) (*TaskAttrs, error) {
  task_attrs := NewTaskAttrs(config)
  task_key := taskKey(task_attrs.Task.Id)

  err := t.db.Insert(task_key, task_attrs)
  if err != nil {
    base.Log("[shellTasker.CreateTask] Failed to create task")
    return nil, err
  }

  return task_attrs, nil
}

func (t *shellTasker) GetTaskAttrs(Id string) (*TaskAttrs, error) {
  var err error
  if Id == "" {
    return nil, fmt.Errorf("Task Id is required parameter to GetTaskAttrs")
  }
  data, err := t.db.Get(taskKey(Id))
  
  task_attrs :=  &TaskAttrs{} 
  err = json.Unmarshal(data, &task_attrs)
  
  return task_attrs, err
}

*/
