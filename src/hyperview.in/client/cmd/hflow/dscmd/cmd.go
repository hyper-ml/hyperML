package dscmd
 
import (
  "fmt"
  "os"
  "strings"
  "path/filepath"
  "github.com/spf13/cobra"

  "hyperview.in/server/base"
  "hyperview.in/client/config"
  "hyperview.in/client"

)

func exitWithError(format string, args ...interface{}) {
  if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}

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
      base.Debug("Current Directory: ", current_dir)

      if len(args) == 0 {
        exitWithError("Dataset name is mandatory")
        cmd.Usage()
      }
      
      dataset_name = args[0]
      base.Debug("name of ds: ", dataset_name)

      c, _ := client.New(current_dir)
      err := c.InitDataRepo(current_dir, dataset_name)
      if err != nil {
        exitWithError(err.Error())
      }

      err = config.WriteRepoParams(current_dir, "REPO_NAME", dataset_name)
      // TODO: use the server response
      err = config.WriteRepoParams(current_dir, "REPO_TYPE", "DATASET")  
      if err != nil {
        exitWithError(err.Error())
      } 

    },
  }

  new_cmd.Flags().StringVar(&dataset_name, "name", "", "dataset name")
  ds_cmd.AddCommand(new_cmd)

  return ds_cmd
}
