package flow

import (
  "strings"
  "fmt"
)

func scrubError(e error) error {
  if e != nil {
    if strings.Contains(e.Error(), "exists") {
      return nil
    }
  }

  return e
}

func InvalidTaskError(taskId string) error{
  return fmt.Errorf("invalid_task: This task does not exist in curernt flow. Invalid Task Id: %s", taskId)
}

func InvalidTaskStatusError() error{
  return fmt.Errorf("invalid_status_error: This task is already complete with status Success or failed")
}

func InvalidWorkerFlowCombo() error {
  return fmt.Errorf("invalid_worker_flow_combo: Invalid worker-flow-task combination. ")
}

func InvalidFlowId() error {
  return fmt.Errorf("invalid_flow_id: Flow Id is either missing or invalid")
}

func InvalidTaskId() error {
  return fmt.Errorf("invalid_task_id: Task Id is either missing or invalid")
}