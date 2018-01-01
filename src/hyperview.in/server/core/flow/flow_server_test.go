package flow


import (
  "fmt"
  "time"
  "testing"
  ws "hyperview.in/server/core/workspace"

  "hyperview.in/server/core/utils"
  . "hyperview.in/server/core/tasks"
)

func dummyFlow(q *queryServer) (*FlowAttrs, *TaskAttrs) {
  flow_rec, task_rec  := dummyRec()
  fmt.Println("Dummy Flow Rec: ", flow_rec)

  err := q.InsertFlow(flow_rec)
  if err != nil {
    fmt.Println("Insert Error: ", err)
  }

  return flow_rec, task_rec
}

func Test_StartFlowServer(t *testing.T) {
  d, _ := utils.FakeDb()
  fs := NewFlowServer(d, "flow_test", nil, nil)
  
  q := NewQueryServer(d)  
  new_flow, _ := dummyFlow(q)

  err := q.UpdateFlow(new_flow.Flow.Id, new_flow)
  if err!= nil {
    fmt.Println("Error: ", err)
    t.Fail()
  }
  time.Sleep(3* time.Second)
  fs.Close()

  err = q.DeleteFlow(new_flow.Flow.Id)
}

func Test_CreateFlowOutRepo(t *testing.T) {
  d, _ := utils.FakeDb()
  ws_api, _ := ws.NewApiServer(d, nil)

  fs := NewFlowServer(d, "flow_test", nil, ws_api)
  // create flow
  flow := Flow{Id: "2321ewqd3243d"}
  r_attrs, _ := fs.GetOrCreateOutRepo(flow)
  if r_attrs == nil {
    t.Fail()
  }
}

 