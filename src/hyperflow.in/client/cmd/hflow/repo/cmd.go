package repo

import (
  "os"  
  "strings"
  "path/filepath"

  "github.com/spf13/cobra"
  
  "hyperflow.in/client"
  "hyperflow.in/client/fs" 
  "hyperflow.in/client/config"
  //"hyperflow.in/server/pkg/base"
  cmd_utils "hyperflow.in/client/cmd/hflow/utils"
)

func dirNameForRepo(repoName string) string {
  return fs.DirNameForRepo(repoName)
} 
 

func validRepoName(s string) bool{
  if strings.ContainsAny(s, " ") || strings.ContainsAny(s, "/") {
    return true
  }

  return  false
}

const (
	concurrency = 10
)

func RepoCommands() []*cobra.Command {
  var repo_name string
  var branch_name string
  var commit_id string 

  repo_cmd := &cobra.Command{
    Use: "repo",
    Short: "Repo command Usage",
    //TODO: add command details
    Long: `Repo command Usage`, 
    Run: func(cmd *cobra.Command, args []string) {
      cmd.Usage()
    },
  }

  init_cmd := &cobra.Command{
    Use: "init",
    Short: "create a new Repo",
    //TODO: add command details
    Long: `Initializes a new repo. Name is mandatory parameter.`, 
    Run: func(cmd *cobra.Command, args []string) {      
      if len(args) == 0 {
        cmd_utils.ExitWithError("Repo name is mandatory. \nformat: hflow init <<repo_name>>")
      }  

      repo_name = args[0]
      
      if len(args) > 1 || validRepoName(repo_name) {
        cmd_utils.ExitWithError("Repo name can not have spaces")
      }

      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

      r, _ := config.ReadRepoParams(current_dir, "REPO_NAME")
      if r != "" {
        cmd_utils.ExitWithError("A Repo is already initialized in this directory:", r)
      }  

      c, _ := client.New(current_dir)      
      err := c.InitRepo(repo_name)

      if err != nil {
        cmd_utils.ExitWithError(err.Error())
      }

      _ = config.WriteRepoParams(current_dir, "REPO_NAME", repo_name)
      _ = config.WriteRepoParams(current_dir, "BRANCH_NAME", "master")

      // create components
      repo_full_path := filepath.Dir(os.Args[0])
      if err := cmd_utils.CreateComponentDirs(repo_full_path, repo_name); err != nil {
        cmd_utils.ExitWithError(err.Error())
      }
    },
  }  

  init_cmd.PersistentFlags().StringVarP(&repo_name, "repo","r", "", "Repo Name")

  clone_cmd := buildCloneCmd(repo_name, branch_name, commit_id)
  repo_cmd.AddCommand(clone_cmd)

  push_cmd  := buildPushCmd(repo_name, branch_name, commit_id)
  repo_cmd.AddCommand(push_cmd)

  var result []*cobra.Command
  result = append(result, init_cmd, repo_cmd, push_cmd)

  return result
}


