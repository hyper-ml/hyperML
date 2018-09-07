package flow

import (
  "fmt"
  "time"
  "hyperview.in/server/core/utils"
  "golang.org/x/sync/errgroup"

  "testing"
)
  

func Test_StartFlowEngineMaster(t *testing.T) {
  var err error
  
  d, err := utils.FakeDb()
  if err != nil {
    fmt.Println("[Test_StartFlowEngineMaster] Failed to acquire FakeDb connection:", err)
    t.Fail()
  }

  q := NewQueryServer(d)
  
  flow_rec, task_rec := dummyFlow(q)

  engine:= NewFlowEngine(q, d)
  
  quit := make(chan int)
  fmt.Println("Starting master ...")
  go engine.master(quit)

  var eg errgroup.Group
  eg.Go(func() error {
      _, err = engine.StartFlow(flow_rec.Flow.Id, task_rec.Task.Id)
      return nil
    })

  eg.Wait()
  time.Sleep(300 * time.Second)
  close(quit)
}

