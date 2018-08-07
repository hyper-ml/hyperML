package worker

import(
  "fmt"
  "testing"
)

const (
  TEST_TASK_ID = "64d5783f211944faafbb262c844b8c8c"
)


func Test_CompletTask(t *testing.T) {
  w := NewWorkeHorse()
  err := w.AddTask(TEST_TASK_ID)
  if err != nil {
    fmt.Println("[w.AddTask] ", err)
    t.Fail()
  }
  err = w.CompleteTask(TEST_TASK_ID)
  if err != nil {
    fmt.Println("[w.CompleteTask] ", err)
    t.Fail()
  }
}