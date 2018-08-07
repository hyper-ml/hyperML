package flow

import(
  "fmt"
  "testing"
  
  tsk_pkg "hyperview.in/server/core/tasks"  
  "hyperview.in/server/core/utils"
)


func dummyRec() *FlowAttrs{
  cmd:= "who"
  wdir:= "/Users/apple/MyProjects/stash"

  task_config:= tsk_pkg.NewTaskConfig(cmd, nil, wdir, nil)
  flow_rec :=  NewFlowAttrs(nil) 
  flow_rec.AddTask(task_config) 
  return flow_rec
}


func Test_InsertFlow(t *testing.T) {

  d, err := utils.FakeDb()

  q:= NewQueryServer(d)

  dummy_rec := dummyRec()

  fmt.Println("Dummy Flow Rec: ", dummy_rec)
  
  err = q.InsertFlow(dummy_rec)
  if err != nil {
    fmt.Println("Insert Error: ", err)
  }
}

