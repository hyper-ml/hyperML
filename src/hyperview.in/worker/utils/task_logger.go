package utils

import (
  "fmt"
)

type Logger interface {
  Write(p []byte) (int, error)
}

type taskLogger struct {
  taskId string
}

func NewTaskLogger(id string) Logger{
  return &taskLogger {
    taskId: id,
  }
}

func (tl *taskLogger) Write(p []byte) (_ int, retErr error) {
  fmt.Println("[taskLogger: " + tl.taskId + "] " + string(p))
  return 0, nil
}








