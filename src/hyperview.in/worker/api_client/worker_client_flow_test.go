package api_client


import(
  "testing"
  tsk_pkg "hyperview.in/server/core/tasks" 

  flw_pkg "hyperview.in/server/core/flow" 

)

const (
  TEST_FLOW_ID = "053087fbbb0e43acaef07d3024b82a5a"
  TEST_TASK_ID = "f2da9c8a26bf4b9d9f274cf185395dc3"
)

func onError(t *testing.T, err error) {
  if err != nil{
    t.Fail()
  }

}

func Test_UpdateTaskStatus(t *testing.T) {
  wc, err := NewWorkerClient("")
  onError(t, err)
  tsr:= &flw_pkg.TaskStatusChangeRequest {
    Flow:  flw_pkg.Flow {
      Id: TEST_FLOW_ID,
    },
    Task: tsk_pkg.Task {
      Id: TEST_TASK_ID,
    },
    TaskStatus: tsk_pkg.TASK_INITIATED,
  }

  _, err = wc.UpdateTaskStatus("", "f2da9c8a26bf4b9d9f274cf185395dc3", tsr)
  onError(t, err)
}

func Test_GetFlowOutRepo(t *testing.T) {
  wc, err := NewWorkerClient("")
  onError(t, err)

  _, _, _, _ = wc.GetFlowOutRepo(TEST_FLOW_ID)

}