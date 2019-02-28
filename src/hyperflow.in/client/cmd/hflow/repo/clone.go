package repo


import (
  "os"
  "fmt"
  "path/filepath"

  "github.com/spf13/cobra"

  "hyperflow.in/client"
  "hyperflow.in/client/config"
  cmd_utils "hyperflow.in/client/cmd/hflow/utils"
)

func buildCloneCmd(repo_name, branch_name, commit_id string) *cobra.Command {
  clone_cmd := &cobra.Command {
    Use: "clone",
    Short: "Clone a Repo", 
    Long: `Clone a given repo`, 
    Run: func(cmd *cobra.Command, args []string) {
      
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      fmt.Println("directory: ", current_dir)
      
      if len(args) == 0 {
        cmd_utils.ExitWithError("Please enter Repo name")
        cmd.Usage()
      }

      repo_name := args[0]
      if repo_name == "" {
        cmd_utils.ExitWithError("Repo name can not be null")
      }

      if branch_name == "" {
        branch_name = "master"
      }

      // args override 
      if len(args) >= 2 {
        branch_name = args[1]
      }

      // args override 
      if len(args) >= 3 {
        commit_id = args[2]
      }

      repo_dir := dirNameForRepo(repo_name)
      if repo_dir == "" {
        cmd_utils.ExitWithError("Directory name can not be null")
      }

      r, _ := config.ReadRepoParams(repo_dir, "REPO_NAME")
      if r != "" {
        cmd_utils.ExitWithError("A Repo is already initialized in this directory:", r)
      }  

      repo_dir = filepath.Join(current_dir, repo_dir)
      c, _ := client.New(repo_dir)

      clone_commit_id, err := c.CloneCommit(repo_name, branch_name, commit_id)
      if err != nil {
        cmd_utils.ExitWithError(err.Error())
      }

      _ = config.WriteRepoParams(repo_dir, "REPO_NAME", repo_name)
      _ = config.WriteRepoParams(repo_dir, "BRANCH_NAME", branch_name)
      _ = config.WriteRepoParams(repo_dir, "COMMIT_ID", clone_commit_id)
    
    },
  }  

  clone_cmd.PersistentFlags().StringVarP(&commit_id, "commit", "c", "", "commit Id")
  clone_cmd.PersistentFlags().StringVarP(&branch_name, "branch", "b", "", "branch name")

  return  clone_cmd
} 
