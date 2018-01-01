package repocmd

import (
  "os"
  "fmt"
  "strings"
  "path/filepath"

  "github.com/spf13/cobra"
  
  //"hyperview.in/server/base"

  client "hyperview.in/client"
  "hyperview.in/client/config"

)
 

func exitWithError(format string, args ...interface{}) {
  if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}

func hasWhiteSpace(s string) bool{
  return strings.ContainsAny(s, " ")
}

const (
	concurrency = 10
)

func RepoCommands() []*cobra.Command {
  var task_id string
  var repo_name string
  var branch_name string
  var curr_commit_id string

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
        exitWithError("Repo name is mandatory. \nformat: hflow init <<repo_name>>")
      }  

      repo_name = args[0]
      if len(args) > 1 || hasWhiteSpace(repo_name) {
        exitWithError("Repo name can not have spaces")
      }
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

      r, _ := config.ReadRepoParams(current_dir, "REPO_NAME")
      if r != "" {
        exitWithError("A Repo is already initialized in this directory:", r)
      }  

      cl, _ := client.NewApiClient(current_dir)      
      err := cl.InitRepo(repo_name)

      if err != nil {
        exitWithError(err.Error())
      }

      _ = config.WriteRepoParams(current_dir, "REPO_NAME", repo_name)
      // assumption: repo always starts with a master branch
      _ = config.WriteRepoParams(current_dir, "BRANCH_NAME", "master")

    },
  }  

  init_cmd.PersistentFlags().StringVarP(&repo_name, "repo","r", "", "Repo Name")

  clone_cmd := &cobra.Command{
    Use: "clone",
    Short: "clone a new Repo",
    //TODO: add command details
    Long: `Clones a new repo from remote URL.`, 
    Run: func(cmd *cobra.Command, args []string) {
      fmt.Println("repo clone")
      cmd.Usage()
    },
  }  

  repo_cmd.AddCommand(clone_cmd)


  push_cmd := &cobra.Command {
    Use: "push",
    Short: "Push new or changed files to server repo",
    //TODO: add command details
    Long: `Push new or changed files to server repo`, 
    Run: func(cmd *cobra.Command, args []string) {

      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      fmt.Println("directory: ", current_dir)

      repo_name, _  = config.ReadRepoParams(current_dir, "REPO_NAME")
      branch_name, _  = config.ReadRepoParams(current_dir, "BRANCH_NAME")
      curr_commit_id, _  = config.ReadRepoParams(current_dir, "COMMIT_ID")

      if repo_name == "" {
        exitWithError("This command works only from the top repo directory. If you aleady in top directory then initialize a repo with - hflow init -n <<name>> ")
      }
  
      c, _ := client.NewApiClient(current_dir)
      commit, err := c.PushRepo(repo_name, branch_name, curr_commit_id)
      if err != nil { 
        exitWithError(err.Error())
      }

      _ = config.WriteRepoParams(current_dir, "COMMIT_ID", commit.Id)
    },
  }

  // not sure you should push to a specific commit from client 
  // push_cmd.PersistentFlags().StringVar(&commit_id, "commit", "", "commit Id")


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
    //TODO: add command details
    Long: `Pull Results from task run`, 
    Run: func(cmd *cobra.Command, args []string) {
      fmt.Println("repo clone")
      cmd.Usage()
    },
  }  

  pull_res_cmd.PersistentFlags().StringVar(&task_id, "task", "", "Task Id")
  pull_cmd.AddCommand(pull_res_cmd)

  var result []*cobra.Command
  result = append(result, init_cmd, repo_cmd, push_cmd, pull_cmd)

  return result
}


