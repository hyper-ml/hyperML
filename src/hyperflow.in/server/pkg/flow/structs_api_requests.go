package flow

import(
  "hyperflow.in/server/pkg/tasks"
  ws "hyperflow.in/server/pkg/workspace"
)

type FlowMessageType int

const (
  FlowRequest FlowMessageType = 10
  FlowResponse FlowMessageType = 20
  FlowData FlowMessageType = 30
)

type FlowMessage struct {
  Type  FlowMessageType
  Flow  *Flow
  Tasks  *[]tasks.Task
  FlowStatusStr string
  Task *tasks.Task
  TaskStatusStr string
  EnvVars map[string]string

  Repos []*ws.RepoMessage
  CmdStr string
}


type FlowAttrsMessage struct {
  Type  FlowMessageType
  FlowAttrs  *FlowAttrs
  TasksAttrs  *[]tasks.TaskAttrs
  TaskAttrs *tasks.TaskAttrs
}

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

type FlowOutRepoRequest struct {
}

type FlowOutRepoResponse struct {
  Repo *ws.Repo
  Branch *ws.Branch
  Commit *ws.Commit
}
