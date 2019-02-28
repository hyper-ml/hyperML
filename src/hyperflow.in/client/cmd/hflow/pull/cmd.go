package pull


import ( 
  "os" 
  "path/filepath"
  "github.com/spf13/cobra"
  
  "hyperflow.in/server/pkg/base"

  "hyperflow.in/client/config"
  client "hyperflow.in/client"
  cmd_utils "hyperflow.in/client/cmd/hflow/utils"
)
 

func PullCommand() *cobra.Command {
  var task_id string

  pull_cmd := &cobra.Command{
    Use: "pull",
    Short: "Pull new or changed files to server repo",
    //TODO: add command details
    Long: `Pull new or changed files to server repo`, 
    Run: func(cmd *cobra.Command, args []string) {
      cmd.Usage()
    },
  }

  pull_res_cmd := &cobra.Command{
    Use: "results",
    Short: "Pull Results from task run", 
    Long: `Pull Results from task run`, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

      switch task_id {
        case "" :
          switch {
            case len(args) > 0:   
              task_id = args[0] 
            case len(args) == 0:
              task_id, _ = config.ReadRepoParams(current_dir, "FLOW_ID")   
          }

        default: 
          base.Debug("[pull_res_cmd] Task Id: ", task_id)
      }
 
      if task_id == "" {
        cmd_utils.ExitWithError("Task (Id) is mandatory")
      }

      c, _ := client.New(current_dir) 
      out_dir, out_repo, out_branch, out_commit, err := c.PullResults(task_id)

      if err != nil {
        cmd_utils.ExitWithError("pull_results_error: %s", err)
      } 

      _ = config.WriteRepoParams(out_dir, "REPO_NAME", out_repo.Name)
      _ = config.WriteRepoParams(out_dir, "BRANCH_NAME", out_branch.Name)
      
      if out_commit != nil {
        _ = config.WriteRepoParams(out_dir, "COMMIT_ID", out_commit.Id)
      }

    },
  }  

  pull_res_cmd.PersistentFlags().StringVar(&task_id, "task", "", "Task Id")
  pull_cmd.AddCommand(pull_res_cmd)

  pull_models_cmd := &cobra.Command{
    Use: "model",
    Short: "Pull saved models from task run",
    Long: `Pull saved models  from task run`, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      switch task_id {
        case "" :
          switch {
            case len(args) > 0:   
              task_id = args[0] 
            case len(args) == 0:
              task_id, _ = config.ReadRepoParams(current_dir, "FLOW_ID")   
          }

        default: 
          base.Debug("[pull_res_cmd] Task Id: ", task_id)
      }
 
      if task_id == "" {
        cmd_utils.ExitWithError("Task (Id) is mandatory")
      }

      c, _ := client.New(current_dir) 
      saved_dir, saved_repo, saved_branch, saved_commit, err := c.PullSavedModels(task_id)

      if err != nil {
        cmd_utils.ExitWithError("pull_models_error: %s", err)
      }

      _ = config.WriteRepoParams(saved_dir, "REPO_NAME", saved_repo.Name)
      _ = config.WriteRepoParams(saved_dir, "BRANCH_NAME", saved_branch.Name)
      _ = config.WriteRepoParams(saved_dir, "COMMIT_ID", saved_commit.Id)

    },
  }  

  pull_models_cmd.PersistentFlags().StringVar(&task_id, "task", "", "Task Id")
  pull_cmd.AddCommand(pull_models_cmd)

  return pull_cmd
}
