package repo


import (
  "os"
  "fmt"
  "strings"
  "path/filepath"
  "github.com/spf13/cobra"

  "hyperflow.in/client"
  "hyperflow.in/client/config"
  cmd_utils "hyperflow.in/client/cmd/hflow/utils"
)

func readIgnoresFromPath(dir string) []string {
  if cmd_utils.PathExists(cmd_utils.IgnoreFileName) {
    ignore_path := filepath.Join(dir, cmd_utils.IgnoreFileName)
    content, err := cmd_utils.GetFileContent(ignore_path)
    
    if err != nil {
      cmd_utils.ExitWithError("failed to read ignore file", err)
    }  

    return strings.Split(string(content), cmd_utils.ListSeparator)
  }

  return nil
}

func buildPushCmd(repo_name, branch_name, commit_id string) *cobra.Command {

  push_cmd := &cobra.Command {
    Use: "push",
    Short: "Push new or changed files to server repo",
    //TODO: add command details
    Long: `Push new or changed files to server repo`, 
    Run: func(cmd *cobra.Command, args []string) {

      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      ignore_list := readIgnoresFromPath(current_dir)
      fmt.Println("ignore list: ", ignore_list)

      repo_name, _  = config.ReadRepoParams(current_dir, "REPO_NAME")
      branch_name, _  = config.ReadRepoParams(current_dir, "BRANCH_NAME")
      commit_id, _  = config.ReadRepoParams(current_dir, "COMMIT_ID")

      if repo_name == "" {
        cmd_utils.ExitWithError("This command works only from the top repo directory. If you aleady in top directory then initialize a repo with - hflow init -n <<name>> ")
      }
  
      c, _ := client.New(current_dir)
      commit, err := c.PushRepo(repo_name, branch_name, commit_id, ignore_list)
      
      if err != nil { 
        cmd_utils.ExitWithError(err.Error())
      }
      if commit == nil {
        cmd_utils.ExitWithError("Err.. Push returned empty commit. Please retry.")
      }

      _ = config.WriteRepoParams(current_dir, "COMMIT_ID", commit.Id)
    },
  }

  return push_cmd
}