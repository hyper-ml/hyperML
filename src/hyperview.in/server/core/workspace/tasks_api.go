package workspace

import(
  task_pkg "hyperview.in/server/core/tasks"
)


funct (a *apiServer) CreateTask(config *task_pkg.TaskConfig) (*Task, error) {
  t := task_pkg.NewTask(config)
  
  return t
}