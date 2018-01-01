package flow



import(
  "fmt"
  "os"
  "time"
  "testing"
  "hyperview.in/server/core/utils"
)


func setEnv(){
  os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/Users/apple/MyProjects/creds/hw-storage-75d060e8419a.json")
  os.Setenv("GOOGLE_STORAGE_BUCKET", "hyperview001")
}

// start pod 
// cancel pod
// complete pod task

func Test_AssignWorker(t *testing.T) {
  db, _:= utils.FakeDb()
  pk := NewDefaultPodKeeper(db)
  qs := NewQueryServer(db)
 
  FlowAttrs, _:= dummyFlow(qs)

  var task_id string
  for k, _:= range FlowAttrs.Tasks {
    task_id = k
  } 

  fmt.Println("Flow Id: ", FlowAttrs.Flow.Id)
  fmt.Println("Task Id: ", task_id)

  err:= pk.AssignWorker(task_id, FlowAttrs)
  if err != nil {
    fmt.Println("Error creating worker: ", err)
    _ = qs.DeleteFlow(FlowAttrs.Flow.Id)  
  }
  time.Sleep(10* time.Second)
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


func Test_SavePodWithDeployId(t *testing.T) {
  setEnv()
  db, _:= utils.FakeDb()

  pk := NewDefaultPodKeeper(db)
  err:= pk.SaveWorkerLog(Worker{}, Flow{Id: "af84fdf93b5741568e8dfb413d3c3047"})
  fmt.Println("err: ", err)
  if err != nil {
    t.Fatalf("Failure: %s", err)
  }
}

func Test_GetPodLog(t *testing.T) {
  setEnv()
  db, _:= utils.FakeDb()

  pk := NewDefaultPodKeeper(db)
  w := Worker { PodId: "e23dfa1225fc40249d4915d8b6f52b6f-86bc799555-bc8vs" }
  f := Flow { Id: "e23dfa1225fc40249d4915d8b6f52b6f" }
  err := pk.SaveWorkerLog(w, f)
  fmt.Println("error: ", err)
}