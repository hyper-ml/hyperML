package flow

import(
  "hyperview.in/server/core/tasks"
  ws "hyperview.in/server/core/workspace"
)

type TaskStatusChangeRequest struct {
  Flow Flow
  Task tasks.Task
  TaskStatus tasks.TaskStatus 
  Message string
}

type TaskStatusChangeResponse struct {
  FlowAttrs *FlowAttrs
}

type NewFlowLaunchRequest struct {
  Repo ws.Repo
  Branch ws.Branch
  Commit  ws.Commit
  CmdString string
}

type NewFlowLaunchResponse struct {
  TaskStatus tasks.TaskStatus
  TaskStatusStr string
  Task *tasks.Task
  Flow *Flow
}
