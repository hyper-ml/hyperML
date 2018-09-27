package repocmd

import (
  "os"
  "fmt"
  "strings"
  "path/filepath"

  "github.com/spf13/cobra"
  
  //"hyperview.in/server/base"

  "hyperview.in/client"
  "hyperview.in/client/config"
  "hyperview.in/client/fs"

)

func dirNameForRepo(repoName string) string {
  return fs.DirNameForRepo(repoName)
} 

func exitWithError(format string, args ...interface{}) {
  if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
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
        exitWithError("Repo name is mandatory. \nformat: hflow init <<repo_name>>")
      }  

      repo_name = args[0]
      
      if len(args) > 1 || validRepoName(repo_name) {
        exitWithError("Repo name can not have spaces")
      }

      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

      r, _ := config.ReadRepoParams(current_dir, "REPO_NAME")
      if r != "" {
        exitWithError("A Repo is already initialized in this directory:", r)
      }  

      c, _ := client.New(current_dir)      
      err := c.InitRepo(repo_name)

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
    Long: `Clones a given repo`, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
      fmt.Println("directory: ", current_dir)
      
      if len(args) == 0 {
        exitWithError("Please enter Repo name")
        cmd.Usage()
      }

      repo_name := args[0]
      if repo_name == "" {
        exitWithError("Repo name can not be null")
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
        exitWithError("Directory name can not be null")
      }

      r, _ := config.ReadRepoParams(repo_dir, "REPO_NAME")
      if r != "" {
        exitWithError("A Repo is already initialized in this directory:", r)
      }  


      repo_dir = filepath.Join(current_dir, repo_dir)
      c, _ := client.New(repo_dir)

      clone_commit_id, err := c.CloneCommit(repo_name, branch_name, commit_id)
      if err != nil {
        exitWithError(err.Error())
      }

      _ = config.WriteRepoParams(repo_dir, "REPO_NAME", repo_name)
      _ = config.WriteRepoParams(repo_dir, "BRANCH_NAME", branch_name)
      _ = config.WriteRepoParams(repo_dir, "COMMIT_ID", clone_commit_id)
    
    },
  }  
  clone_cmd.PersistentFlags().StringVarP(&commit_id, "commit", "c", "", "commit Id")
  clone_cmd.PersistentFlags().StringVarP(&branch_name, "branch", "b", "", "branch name")

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
      commit_id, _  = config.ReadRepoParams(current_dir, "COMMIT_ID")

      if repo_name == "" {
        exitWithError("This command works only from the top repo directory. If you aleady in top directory then initialize a repo with - hflow init -n <<name>> ")
      }
  
      c, _ := client.New(current_dir)
      commit, err := c.PushRepo(repo_name, branch_name, commit_id)
      
      if err != nil { 
        exitWithError(err.Error())
      }

      _ = config.WriteRepoParams(current_dir, "COMMIT_ID", commit.Id)
    },
  }

  // not sure you should push to a specific commit from client 
  // push_cmd.PersistentFlags().StringVar(&commit_id, "commit", "", "commit Id")



  var result []*cobra.Command
  result = append(result, init_cmd, repo_cmd, push_cmd)

  return result
}


