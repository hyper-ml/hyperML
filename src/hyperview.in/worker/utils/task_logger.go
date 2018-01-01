package utils

import (
  "os"
  "fmt"
 "hyperview.in/server/base"
)

const (
  LOG_DIRECTORY = "log"
  LOG_PERM os.FileMode = 0755
)
type LoggerInterface interface {
  Write(p []byte) (int, error)
  GetLogPath() string
  GetLogFile() *os.File
  Close()
}

type TaskLogger struct {
  taskId string
  logDir string 
  logPath string
  log *os.File
}

func NewTaskLogger(homeDir string, taskId string) (LoggerInterface){
  base.Info("[utils.NewTaskLogger] homeDir, taskId: ", homeDir, taskId)
  log_dir := homeDir + "/" + LOG_DIRECTORY
  log_perm := LOG_PERM
  log_path := log_dir + "/" + taskId + ".log"
  
  if err := MkDirAll(log_dir, log_perm); err != nil {
    base.Warn("[utils.NewTaskLogger] Failed to create log dir for worker", err)
    return nil
  }

  f, err := os.Create(log_path)
  
  if err != nil {
    base.Warn("[utils.NewTaskLogger] Failed to create log file for worker", err)
    return nil
  }
 
  return &TaskLogger {
    taskId: taskId,
    logDir: log_dir,
    logPath: log_path,
    log: f,
  }
}

func (l *TaskLogger) Write(p []byte) (_ int, retErr error) {
  
  fmt.Println("[taskLogger: " + l.taskId + "] " + string(p))
  _, err := l.log.Write(p)
  
  if err != nil {
    return 0, err
  }

  return 0, nil
}

func (l *TaskLogger) Close() {
  defer l.log.Close()  
}

func (l *TaskLogger) GetLogPath() string {
  return l.logPath
}

func (l *TaskLogger) GetLogFile() *os.File {
  return l.log
}




