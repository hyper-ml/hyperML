package worker_test

import(
  "fmt"
  "testing"
  . "hyperview.in/worker"
)

const (
  TEST_FLOW_ID = "1d533970742b4ee88544b30c0a81ed5b"
  TEST_TASK_ID = "1d533970742b4ee88544b30c0a81ed5b"
  TEST_WORKER_IP = "192.1.1.94"
  TEST_WORK_DIR = "/Users/apple/MyProjects/stash/1d533970742b4ee88544b30c0a81ed5b"
)


func failOnError(t *testing.T, err error) {
  if err != nil {
    fmt.Println("worker_test.go: Test failed with error: ", err)
    t.Fail()
  }
}

func Test_WorkerCycle(t *testing.T) {

  w := NewWorkHorse("", TEST_FLOW_ID, TEST_TASK_ID, TEST_WORKER_IP, TEST_WORK_DIR)
  
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