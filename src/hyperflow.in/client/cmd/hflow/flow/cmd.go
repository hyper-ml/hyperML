package flow

import ( 
  "os" 
  "fmt"
  "path/filepath"
  "github.com/spf13/cobra"
  "hyperflow.in/client/config"

  client "hyperflow.in/client"

  cmd_utils "hyperflow.in/client/cmd/hflow/utils"
)

const (
  concurrency = 10
)

func checkNullRepo(repo_name string) error {
  if repo_name == "" {
    return fmt.Errorf("Please initialise a repo before running task.")
  }
  return nil
} 

func checkNullTask(task string) error {
  if task == "" {
    return fmt.Errorf("Task Command is mandatory")
  }
  return nil
}

func isEmpty(v string) bool {
  if v == "" {
    return true
  }
  return false
}

func FlowCommands() []*cobra.Command {
  var task string

  var env_vars map[string]string

  flow_cmd := &cobra.Command{
    Use: "run",
    Short: "Start a hyperflow task for a given repo and commit",
    //TODO: add command details
    Long: `start a hyperflow task for a given repo and commit`, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      
      repo_name, _ := config.ReadRepoParams(current_dir, "REPO_NAME")
      branch_name, _ := config.ReadRepoParams(current_dir, "BRANCH_NAME")
      commit_id, _ := config.ReadRepoParams(current_dir, "COMMIT_ID")
      cmd_str := task 

      if len(args) > 0 {
        cmd_str = args[0]
      } else {
        cmd_str = task
      }

      switch {
      case isEmpty(repo_name):  
        cmd_utils.ExitWithError("Fatal Error: Please initialise a repo before running a task")      
      case isEmpty(cmd_str): 
        cmd_utils.ExitWithError("Fatal Error: A Task command was not found. Please see hflow run --help for command usage ")      
      }

      c, _ := client.New(current_dir)
      flow, commit,  err := c.RunTask(repo_name, branch_name, commit_id, cmd_str, env_vars)
      
      if err != nil {
        cmd_utils.ExitWithError(err.Error())
      }

      fmt.Println("Flow Id: ", flow.Id) 
 
      _ = config.WriteRepoParams(current_dir, "FLOW_ID", flow.Id)
      _ = config.WriteRepoParams(current_dir, "TASK_ID", flow.Id)
      _ = config.WriteRepoParams(current_dir, "COMMIT_ID", commit.Id)  
    
    },
  }
  flow_cmd.Flags().StringVarP(&task, "task","t", "", "task command")
  flow_cmd.Flags().StringToStringVarP(&env_vars, "env", "e", nil, "Environment vars")
    
  flow_status_cmd:= &cobra.Command {
    Use: "status",
    Short: "Status of flow/task status ", 
    Long: `Status of flow/task status`, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      var flow_id string

      last_flow_id, _ := config.ReadRepoParams(current_dir, "FLOW_ID")

      switch {
      case task != "":
        flow_id = task
      case last_flow_id != "":
        flow_id = last_flow_id 
      default: 
        cmd.Usage()
        cmd_utils.ExitWithError("Task or Flow ID is mandatory ") 
      }  
 
      c, _ := client.New(current_dir)
      flow_status, err := c.GetFlowStatus(flow_id)
      if err != nil {
        cmd_utils.ExitWithError(err.Error())
      }

      fmt.Println("Task: " + flow_id)
      fmt.Println("Status: ", flow_status)

    },
  }
  flow_status_cmd.Flags().StringVarP(&task, "task","t", "", "task command")

  var result []*cobra.Command
  result = append(result, flow_cmd, flow_status_cmd)
  return result
}


