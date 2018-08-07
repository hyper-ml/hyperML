package flow



import(
  "fmt"
  "time"
  "testing"
  "hyperview.in/server/core/utils"
)


// start pod 
// cancel pod
// complete pod task

func Test_AssignWorker(t *testing.T) {
  pk := NewDefaultPodKeeper()
  db, _:= utils.FakeDb()
  qs := NewQueryServer(db)
 
  FlowAttrs:= dummyFlow(qs)

  var task_id string
  for k, _:= range FlowAttrs.Tasks {
    task_id = k
  } 

  fmt.Println("Flow Id: ", FlowAttrs.Flow.Id)
  fmt.Println("Task Id: ", task_id)

  err:= pk.AssignWorker(task_id, *FlowAttrs)
  if err != nil {
    fmt.Println("Error creating worker: ", err)
    _ = qs.DeleteFlow(FlowAttrs.Flow.Id)  
  }
  time.Sleep(2* time.Second)
  err = pk.ReleaseWorker(FlowAttrs.Flow)
  if err != nil {
    fmt.Println("Error releasing/deleting worker: ", err)
    _ = qs.DeleteFlow(FlowAttrs.Flow.Id)  
  }

  _ = qs.DeleteFlow(FlowAttrs.Flow.Id)  
  
}

/*
func Test_ReleaseWorker(t *testing.T) {
  pk := NewDefaultPodKeeper()
  db, _:= utils.FakeDb()
  qs := NewQueryServer(db)
 
  FlowAttrs, _:= qs.GetFlowAttr("095321853de742cf90a0d9ecf8e2d285")

}*/