package utils

import (
  "os"
  "fmt"
  filepath_pkg "path/filepath"
  "hyperflow.in/server/pkg/base"

)

const ( 
  LOG_PERM os.FileMode = 0755
)
 

type LogHandler struct {
  listening bool
  logpath string
  logf *os.File

  rcv chan interface{}

  done chan int
}

func setLog(logpath string) (*os.File, error) {
  log_dir := filepath_pkg.Dir(logpath)

  if err := MkDirAll(log_dir, LOG_PERM); err != nil {
    base.Warn("[utils.NewTaskLogger] Failed to create log dir for worker", logpath, err)
    return nil, err
  }
 
  return os.Create(logpath)
}

func NewLogHandler(logpath string, bufflen int) *LogHandler {
  f, err := setLog(logpath)
  if err != nil {
    base.Error("Failed to open log file: ", err)
    return nil
  }

  lh := &LogHandler {
    rcv: make(chan interface{}, bufflen),
    done: make(chan int),
    listening: false,
    logf: f,
    logpath: logpath,
  }

  go lh.Listen() 
  return lh
}
 
func (lh *LogHandler) GetLogPath() string {
  return lh.logpath
}


func (lh *LogHandler) Record(m []byte) error {
  lh.rcv <- m
  return nil
} 
 

func (lh *LogHandler) Write(m []byte) (_ int, retErr error) {
  retErr = lh.Record(m)
  return 
}

func (lh *LogHandler) Close() {
  lh.done <- 0
}
 
func (lh *LogHandler) Listen() {
   
  if (lh.listening != false) {
    base.Log("[ChangeListener.Listen] Already one listener is active")
    return 
  } 

  lh.listening = true

  for { 
    select {
    case msg := <- lh.rcv: 
      m := string(msg.([]byte))

      fmt.Println(m)
    case <- lh.done:
      break
    }
  }
  return
}