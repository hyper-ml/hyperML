package dataset
 
import ( 
  "os" 
  "path/filepath"
  "github.com/spf13/cobra"

  "hyperflow.in/server/pkg/base"
  "hyperflow.in/client/config"
  "hyperflow.in/client"

  cmd_utils "hyperflow.in/client/cmd/hflow/utils"
)
 
func DsCommand() *cobra.Command {
  var dataset_name string

  ds_cmd := &cobra.Command{
    Use: "dataset",
    Short: "manage dataset ",
    Long: `manage dataset `, 
    Run: func(cmd *cobra.Command, args []string) {
      cmd.Usage()
    },
  }

  new_cmd :=  &cobra.Command{
    Use: "new",
    Short: "new dataset ",
    Long: `new dataset `, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

      if len(args) == 0 {
        cmd_utils.ExitWithError("Dataset name is mandatory")
        cmd.Usage()
      }
      
      dataset_name = args[0]
      base.Debug("name of ds: ", dataset_name)

      c, _ := client.New(current_dir)
      err := c.InitDataRepo(current_dir, dataset_name)
      if err != nil {
        cmd_utils.ExitWithError(err.Error())
      }

      err = config.WriteRepoParams(current_dir, "REPO_NAME", dataset_name)
      // TODO: use the server response
      err = config.WriteRepoParams(current_dir, "REPO_TYPE", "DATASET")  
      if err != nil {
        cmd_utils.ExitWithError(err.Error())
      } 

    },
  }

  new_cmd.Flags().StringVar(&dataset_name, "name", "", "dataset name")
  ds_cmd.AddCommand(new_cmd)

  return ds_cmd
}
