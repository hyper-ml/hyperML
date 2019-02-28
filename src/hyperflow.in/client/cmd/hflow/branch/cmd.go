package branch
 

import ( 
  "os" 
  "path/filepath"
  "github.com/spf13/cobra"

  "hyperflow.in/server/pkg/base"

  "hyperflow.in/client"
  "hyperflow.in/client/config"
 
)

func addBranch(current_dir string, brname string) error {
  
  c, _ := client.New(current_dir)
  repo_name, _ := config.ReadRepoParams(current_dir, "REPO_NAME")
  current_commit_id,_ := config.ReadRepoParams(current_dir, "COMMIT_ID")  
  
  err := c.InitBranch(repo_name, brname, current_commit_id)
  if err != nil {
    return err
  }

  _ = config.WriteRepoParams(current_dir, "BRANCH_NAME", brname)
  return nil
}


func checkout(current_dir string, brname string) error {
  // pull repo branch 
  // set branch 
  repo_name, _ := config.ReadRepoParams(current_dir, "REPO_NAME")

  c, _ := client.New(current_dir)
  commit_id, err := c.CloneBranch(repo_name, brname)
  if err != nil {
    return err 
  }

  _ = config.WriteRepoParams(current_dir, "BRANCH_NAME", brname)
  _ = config.WriteRepoParams(current_dir, "COMMIT_ID", commit_id)
  return nil
}
  
func BranchCommand() *cobra.Command {
  var branch_name string
  var err error 
  branch_cmd := &cobra.Command{
    Use: "experiment",
    Short: "code checkout experiments ",
    Long: `code checkout experiments `, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

      if len(args) > 0 {
        // received a branch name 
        base.Log("[BranchCommand] Received a branch checkout")
        err = checkout(current_dir, args[0])
      }  

      if branch_name != "" {
        base.Log("[BranchCommand] New branch request: ", branch_name)
        err = addBranch(current_dir, branch_name)
      }

    },
  }
  branch_cmd.Flags().StringVarP(&branch_name, "new-branch", "n", "", "new branch name")

  return branch_cmd
}
