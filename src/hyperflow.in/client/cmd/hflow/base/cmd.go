package base


import(
  "os"
  "github.com/spf13/cobra"

  repo_cmd "hyperflow.in/client/cmd/hflow/repo"
  flow_cmd "hyperflow.in/client/cmd/hflow/flow"
  log_cmd "hyperflow.in/client/cmd/hflow/log"
  pull_cmd "hyperflow.in/client/cmd/hflow/pull"
  config_cmd "hyperflow.in/client/cmd/hflow/config"
  branch_cmd "hyperflow.in/client/cmd/hflow/branch"
  dataset_cmd "hyperflow.in/client/cmd/hflow/dataset"
 
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
  
  root_cmd.AddCommand(version_cmd)
  root_cmd.AddCommand(auth_cmd)

  repo_cmds := repo_cmd.RepoCommands()
  for _, cmd := range repo_cmds {
    root_cmd.AddCommand(cmd)
  }

  flow_cmds := flow_cmd.FlowCommands()
  for _, cmd := range flow_cmds {
    root_cmd.AddCommand(cmd)
  }

  log_cmds := log_cmd.LogCommands()
  for _, cmd := range log_cmds {
    root_cmd.AddCommand(cmd)
  }

  root_cmd.AddCommand(pull_cmd.PullCommand())
  root_cmd.AddCommand(config_cmd.ConfigCommand())
  root_cmd.AddCommand(branch_cmd.BranchCommand())
  root_cmd.AddCommand(dataset_cmd.DsCommand())
  return root_cmd, nil
}

 