package repocmd

import (

  "fmt"
  "github.com/spf13/cobra"
)

const (
	concurrency = 10
)

func RepoCommands() []*cobra.Command {

  repoCmd := &cobra.Command{
    Use: "clone-repo",
    Short: "Clone a new Repo",
    //TODO: add command details
    Long: `Repos are cloned with clone-repo command`, 
    Run: func(cmd *cobra.Command, args []string) {
      cmd.Usage()
    },
  }

  var result []*cobra.Command
  result = append(result, repoCmd)
  return result
}


