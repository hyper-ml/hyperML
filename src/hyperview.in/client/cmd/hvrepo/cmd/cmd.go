package cmd


import(
  "os"
  "hyperview.in/client/cmd/hvrepo/repocmd"
  "github.com/spf13/cobra"
)



func RunnerCmd() (*cobra.Command, error) {
	var verbose bool 

  rootCmd := &cobra.Command{
    Use: os.Args[0],
    Short: "Clone Deploy and Run Models with Hyperview", 
    Long: `Clone Deploy and Run Models with Hyperview `,
    //TODO: add logger PersistentPreRun
  }

  rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Output verbose logs")
  
  repoCommands := repocmd.RepoCommands()
  for _, cmd := range repoCommands {
    rootCmd.AddCommand(cmd)
  }

  return rootCmd, nil
}

 