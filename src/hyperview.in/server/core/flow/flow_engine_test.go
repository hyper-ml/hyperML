package flow

import (
  "fmt"
  "hyperview.in/server/core/utils"

  "testing"
)
  

func Test_StartMaster(t *testing.T) {
  d, _ := utils.FakeDb()

  engine:= NewFlowEngine(nil, d)
  
  qch := make(chan int)
  fmt.Println("Starting master ...")
  engine.master(qch)

}

