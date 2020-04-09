package flow

import (
	"fmt"
	"strings"
)

func scrubError(e error) error {
	if e != nil {
		if strings.Contains(e.Error(), "exists") {
			return nil
		}
	}

	return e
}

// InvalidTaskError : Raised when invalid task Id is passed to API
func InvalidTaskError(taskID string) error {
	return fmt.Errorf("invalid_task: This task does not exist in curernt flow. Invalid Task Id: %s", taskID)
}

// InvalidTaskStatusError : Error object
func InvalidTaskStatusError() error {
	return fmt.Errorf("invalid_status_error: This task is already complete with status Success or failed")
}

// InvalidWorkerFlowCombo : Error object
func InvalidWorkerFlowCombo() error {
	return fmt.Errorf("invalid_worker_flow_combo: Invalid worker-flow-task combination. ")
}

// InvalidFlowID : Error object
func InvalidFlowID() error {
	return fmt.Errorf("invalid_flow_id: Flow Id is either missing or invalid")
}

// InvalidTaskID : Error object
func InvalidTaskID() error {
	return fmt.Errorf("invalid_task_id: Task Id is either missing or invalid")
}
