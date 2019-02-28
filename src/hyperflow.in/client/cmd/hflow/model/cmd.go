package model


import(
  "github.com/spf13/cobra"
)

func validRepoName(s string) bool{
  if strings.ContainsAny(s, " ") || strings.ContainsAny(s, "/") {
    return true
  }

  return  false
}

func modelCmd() *cobra.Command {
  var model_name string

  model_cmd := &cobra.Command {
    Use: "model",
    Short: "Manage external models",
    Long: `Manage external models `, 
    Run: func(cmd *cobra.Command, args []string) {
      cmd.Usage()
    },
  }

  new_cmd := &cobra.Command {
    Use: "init",
    Short: "Define an external model",
    Long: `Define an external model `, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      
      if r, _ := config.ReadRepoParams(current_dir, "REPO_NAME"); r != "" {
        cmd_utils.ExitWithError("Repo is already initialised in this directory")
      }
 
      if len(args) != 0 {
        model_name = args[0]
      } else {
        cmd_utils.ExitWithError("Model name is mandatory. e.g. - hflow model init <<model_name>>")
      }
  
      
      // create model
      c, _ := client.New(current_dir)
      if err := c.InitModelRepo(current_dir, model_name); err != nil {
        cmd_utils.ExitWithError(err.Error())
      }     

      // save repo params - repo name and branch
      if err = config.WriteRepoParams(current_dir, "REPO_NAME", model_name); err != nil {
        cmd_utils.ExitWithError(err.Error())
      }
      
      if err = config.WriteRepoParams(current_dir, "REPO_TYPE", "MODEL"); err != nil {
        cmd_utils.ExitWithError(err.Error())
      }

      if err = config.WriteRepoParams(current_dir, "BRANCH_NAME", "master"); err != nil {
        cmd_utils.ExitWithError(err.Error())
      }
      
    },
  }


  upload_cmd := &cobra.Command {
    Use: "upload",
    Short: "Upload external models",
    Long: `Upload external models `, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      model_name, _ = config.ReadRepoParams(current_dir, "REPO_NAME") 
      branch_name, _ = config.ReadRepoParams(current_dir, "BRANCH_NAME") 
      
      c, _ := client.New(current_dir)
      commit, err := c.PushRepo(model_name, branch_name, "")
      if err != nil {
        cmd_utils.ExitWithError(err.Error())
      }

      if err = config.WriteRepoParams(current_dir, "COMMIT_ID", commit.Id); err != nil {
        cmd_utils.ExitWithError(err.Error())
      }
    },
  }

  model_cmd.AddCommand(upload_cmd)
  return model_cmd
}