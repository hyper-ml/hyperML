package worker

import (
  "time"
  ts "hyperview.in/server/core/tasks"
)




type worker struct {
  tasks []ts.Task
  started time.Time
}



func Worker() *worker {
  return &worker {}
}


