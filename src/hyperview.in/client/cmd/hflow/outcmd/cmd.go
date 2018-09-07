package outcmd


import (
  "fmt"
  "os"
  "strings"
  "path/filepath"
  "github.com/spf13/cobra"

  client "hyperview.in/client"
)

func exitWithError(format string, args ...interface{}) {
  if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
    fmt.Fprintf(os.Stderr, "%s\n", errString)
  }
  os.Exit(1)
}

func OutCommand() *cobra.Command {
  var flow_id string

  out_cmd := &cobra.Command{
    Use: "out",
    Short: "Pull output of flow Id ",
    //TODO: add command details
    Long: `Pull output of flow Id `, 
    Run: func(cmd *cobra.Command, args []string) {
      current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

      cl, _ := client.NewApiClient(current_dir)
      log_bytes, err := cl.RequestLog(flow_id)
      if err != nil {
        exitWithError(err.Error())
      }
      fmt.Println(string(log_bytes))
    },
  }

  out_cmd.Flags().StringVar(&flow_id, "task", "", "task Id")

  
  return out_cmd
}
