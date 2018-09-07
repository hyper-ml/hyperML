package flow

import(
  "fmt"
  "testing"
  
  tsk_pkg "hyperview.in/server/core/tasks"  
  "hyperview.in/server/core/utils"
)
 

func dummyRec() (*FlowAttrs, *tsk_pkg.TaskAttrs) {
  cmd:= "who"
  wdir:= "/workspace"
  
  mmap := tsk_pkg.NewMountConfig("test_repo", "7759ff96ae2046bbbf89de9bf14ab381", "/workspace", 0)

  flow_config := &FlowConfig{
    MountMap: mmap,
  }

  flow_rec :=  NewFlowAttrs(flow_config) 
  
  task_config:= tsk_pkg.NewTaskConfig(cmd, nil, wdir, mmap)
  task_rec := flow_rec.AddTask(task_config)

  return flow_rec, task_rec
}


func Test_InsertFlow(t *testing.T) {

  d, err := utils.FakeDb()
  if err != nil {
    fmt.Println("[Test_InsertFlow] Failed to acquire FakeDb connection:", err)
    t.Fail()
  }
  q := NewQueryServer(d)

  dummy_rec, _ := dummyRec()

  err = q.InsertFlow(dummy_rec)
  if err != nil {
    fmt.Println("[Test_InsertFlow] Insert Flow Error: ", err)
  }
}

