package cmd


import(
  "os"
  "github.com/spf13/cobra"

  repocmd "hyperview.in/client/cmd/hflow/repocmd"
  flowcmd "hyperview.in/client/cmd/hflow/flowcmd"
  logcmd "hyperview.in/client/cmd/hflow/logcmd"
  pullcmd "hyperview.in/client/cmd/hflow/pullcmd"
  configcmd "hyperview.in/client/cmd/hflow/configcmd"
  branchcmd "hyperview.in/client/cmd/hflow/branchcmd"
  dscmd "hyperview.in/client/cmd/hflow/dscmd"
)


func RootCmd() (*cobra.Command, error) {
	var verbose bool 

  root_cmd := &cobra.Command{
    Use: os.Args[0],
    Short: "Hyperflow Client", 
    Long: `One-clik deploy and train Models with Hyperview client `,
    //TODO: add logger PersistentPreRun
  }

  root_cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Output verbose logs")
  
  root_cmd.AddCommand(versionCmd)

  repo_cmds := repocmd.RepoCommands()
  for _, cmd := range repo_cmds {
    root_cmd.AddCommand(cmd)
  }

  flow_cmds := flowcmd.FlowCommands()
  for _, cmd := range flow_cmds {
    root_cmd.AddCommand(cmd)
  }

  log_cmds := logcmd.LogCommands()
  for _, cmd := range log_cmds {
    root_cmd.AddCommand(cmd)
  }

  root_cmd.AddCommand(pullcmd.PullCommand())
  root_cmd.AddCommand(configcmd.ConfigCommand())
  root_cmd.AddCommand(branchcmd.BranchCommand())
  root_cmd.AddCommand(dscmd.DsCommand())
  return root_cmd, nil
}

 