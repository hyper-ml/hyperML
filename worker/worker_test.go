package worker_test

import(
  "fmt"
  "testing"
  . "github.com/hyper-ml/hyperml/worker"
)

const (
  TEST_FLOW_ID = ""
  TEST_TASK_ID = ""
  TEST_WORKER_IP = "192.1.1.94"
  TEST_WORK_DIR = "/var/tmp/stash/"
)


func failOnError(t *testing.T, err error) {
  if err != nil {
    fmt.Println("worker_test.go: Test failed with error: ", err)
    t.Fail()
  }
}

func Test_WorkerCycle(t *testing.T) {

  w := NewWorkHorse(nil, "", "/var/tmp", TEST_FLOW_ID, TEST_TASK_ID, TEST_WORKER_IP, TEST_WORK_DIR)
  
  fmt.Println("Calling Init..")
  err := w.Init() 
  failOnError(t, err)
  
  if err == nil {

    fmt.Println("Calling DoWork..")
    err = w.DoWork() 
    failOnError(t, err)
    
    if err == nil {
      
      fmt.Println("Calling Shutdown..")
      err = w.Shutdown()
      failOnError(t, err)
    }  
  } 
  
}

