package workspace 

import(
  task_pkg "hyperview.in/server/core/tasks"
)

type NewFlowLaunchRequest struct {
  Repo Repo
  Commit  Commit
  Command string
}

type NewFlowLaunchResponse struct {
  TaskStatus task_pkg.TaskStatus
}
